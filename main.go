// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/UnitVectorY-Labs/isplaintextfile"
	"github.com/atotto/clipboard"
)

// Define the max total size limit in bytes (1MB = 1,048,576 bytes)
const maxTotalSize = 1 * 1024 * 1024 // 1MB in bytes

var Version = "dev" // This will be set by the build systems to the release version

func main() {
	// Define existing flags
	delimiter := flag.String("delimiter", "", "Set the delimiter for file content (default: ```)")
	maxSize := flag.Int("max-size", 0, "Maximum file size to include in KB (default: 32 KB)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	// Define new flags for include and exclude with support for wildcards
	include := flag.String("include", "", "Comma-separated list of patterns to include, even if hidden (e.g., .github,*.env)")
	exclude := flag.String("exclude", "", "Comma-separated list of patterns to exclude (e.g., LICENSE,*.md)")

	noRecursive := flag.Bool("no-recursive", false, "Disable recursive directory traversal (only process files in current directory)")

	showVersion := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *showVersion {
		fmt.Println("Version:", Version)
		return
	}

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Create the config stack with home and root configs preloaded
	configStack := NewConfigStack(dir, *verbose)

	// Determine if flags were set by the user
	delimiterSet := false
	maxSizeSet := false
	includeSetFlag := false
	excludeSetFlag := false
	noRecursiveSet := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "delimiter" {
			delimiterSet = true
		}
		if f.Name == "max-size" {
			maxSizeSet = true
		}
		if f.Name == "include" {
			includeSetFlag = true
		}
		if f.Name == "exclude" {
			excludeSetFlag = true
		}
		if f.Name == "no-recursive" {
			noRecursiveSet = true
		}
	})

	// Get effective config from the stack (without scoped configs yet)
	effectiveConfig := configStack.GetEffectiveConfig()

	// Apply CLI overrides for initial traversal settings
	if delimiterSet {
		effectiveConfig.Delimiter = *delimiter
	}
	if maxSizeSet {
		effectiveConfig.MaxSizeKB = *maxSize
	}
	if noRecursiveSet {
		effectiveConfig.NoRecursive = *noRecursive
	}

	// CLI patterns for include/exclude - these will be appended to all effective configs
	var cliIncludePatterns []string
	var cliExcludePatterns []string

	if includeSetFlag && *include != "" {
		cliIncludePatterns = parseCommaSeparated(*include)
	}
	if excludeSetFlag && *exclude != "" {
		cliExcludePatterns = parseCommaSeparated(*exclude)
	}

	if *verbose {
		// Print out the configuration values
		fmt.Println("Initial Configuration:")
		fmt.Printf("\tDelimiter: %s\n", effectiveConfig.Delimiter)
		fmt.Printf("\tMax Size: %d KB\n", effectiveConfig.MaxSizeKB)
		fmt.Printf("\tInclude Patterns: %v\n", effectiveConfig.Include)
		fmt.Printf("\tExclude Patterns: %v\n", effectiveConfig.Exclude)
		fmt.Printf("\tCLI Include Patterns: %v\n", cliIncludePatterns)
		fmt.Printf("\tCLI Exclude Patterns: %v\n", cliExcludePatterns)
		fmt.Printf("\tNo Recursive: %v\n", effectiveConfig.NoRecursive)
	}

	var builder strings.Builder
	totalSize := 0 // Track total size of the output

	// Track directories where we pushed a config layer and track current path for popping
	pushedDirs := make(map[string]bool)
	var currentPath string // Track the current directory path for proper popping

	// Walk through the current folder and process files using WalkDir for better performance
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Pop configs when we leave directories (detect when path is not under currentPath)
		if currentPath != "" && d.IsDir() {
			// Pop any configs from directories we've left
			for len(pushedDirs) > 0 {
				// Check if any pushed directory is no longer a parent of the current path
				toRemove := []string{}
				for pushedDir := range pushedDirs {
					// If current path doesn't start with pushedDir, we've left that directory
					if !strings.HasPrefix(path, pushedDir+string(filepath.Separator)) && path != pushedDir {
						configStack.Pop(pushedDir)
						toRemove = append(toRemove, pushedDir)
					}
				}
				if len(toRemove) == 0 {
					break
				}
				for _, d := range toRemove {
					delete(pushedDirs, d)
				}
			}
		}

		// Get the base name of the file/directory
		name := d.Name()

		// Get the effective config at this point (including any scoped configs)
		effectiveConfig := configStack.GetEffectiveConfig()

		// Apply CLI overrides (CLI always wins)
		if delimiterSet {
			effectiveConfig.Delimiter = *delimiter
		}
		if maxSizeSet {
			effectiveConfig.MaxSizeKB = *maxSize
		}
		if noRecursiveSet {
			effectiveConfig.NoRecursive = *noRecursive
		}

		// Combine config patterns with CLI patterns (CLI patterns take priority)
		includePatterns := append([]string{}, effectiveConfig.Include...)
		excludePatterns := append([]string{}, effectiveConfig.Exclude...)

		if includeSetFlag {
			// When CLI include is set, use only CLI include patterns for final decision
			includePatterns = cliIncludePatterns
		}

		// Always append CLI exclude patterns (exclusions are cumulative)
		excludePatterns = append(excludePatterns, cliExcludePatterns...)

		// Get relative path for pattern matching
		relPath, _ := filepath.Rel(dir, path)
		if relPath == "." {
			relPath = ""
		}
		if relPath != "" && !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}

		// Check if the file/directory matches any exclude patterns
		excluded := matchesAnyPatternWithPath(name, relPath, excludePatterns)

		// Check if explicitly included (config include can rescue from config exclude,
		// but CLI exclude always wins)
		explicitlyIncluded := matchesAnyPatternWithPath(name, relPath, includePatterns)

		// CLI exclude patterns always win - if matched by CLI exclude, exclude it
		cliExcluded := matchesAnyPatternWithPath(name, relPath, cliExcludePatterns)

		// Final exclusion decision:
		// - If CLI excludes it, it's excluded (CLI always wins)
		// - If config excludes it but config also includes it, it's included (include can rescue)
		// - If config excludes it and not explicitly included, it's excluded
		shouldExclude := excluded && (!explicitlyIncluded || cliExcluded)

		if shouldExclude {
			if d.IsDir() {
				if *verbose {
					fmt.Printf("Excluding directory (matched exclude pattern): %s\n", path)
				}
				return filepath.SkipDir // Skip the entire directory
			}
			if *verbose {
				fmt.Printf("Excluding file (matched exclude pattern): %s\n", path)
			}
			return nil // Skip the file
		}

		// Handle hidden files and directories
		if strings.HasPrefix(name, ".") {
			// Check if the hidden file/directory matches any include patterns
			included := matchesAnyPatternWithPath(name, relPath, includePatterns)

			if !included {
				if *verbose {
					fmt.Printf("Skipping hidden file/directory: %s\n", path)
				}
				if d.IsDir() {
					return filepath.SkipDir // Skip the entire hidden directory
				}
				return nil // Skip the hidden file
			}
			// If the hidden file/directory is in the include patterns, proceed
			if *verbose {
				fmt.Printf("Including hidden file/directory (matched include pattern): %s\n", path)
			}
		}

		// If it's a directory (and not skipped), handle config stack and continue traversing
		if d.IsDir() {
			if *verbose {
				fmt.Printf("Entering directory: %s\n", path)
			}

			// Update current path
			currentPath = path

			// Check for scoped config in this directory (not root)
			if path != dir {
				if configStack.PushIfExists(path) {
					pushedDirs[path] = true
				}
			}

			// If no-recursive is set, skip subdirectories
			if effectiveConfig.NoRecursive && path != dir {
				if *verbose {
					fmt.Printf("Skipping subdirectory (no-recursive enabled): %s\n", path)
				}
				// Pop if we pushed for this directory
				if pushedDirs[path] {
					configStack.Pop(path)
					delete(pushedDirs, path)
				}
				return filepath.SkipDir
			}
			return nil
		}

		// Get file info for size check
		info, err := d.Info()
		if err != nil {
			if *verbose {
				fmt.Printf("Error getting file info for %s: %v\n", path, err)
			}
			return nil
		}

		// Skip files larger than the specified max size
		maxSizeBytes := int64(effectiveConfig.MaxSizeKB) * 1024
		if info.Size() > maxSizeBytes {
			if *verbose {
				fmt.Printf("Skipping large file (%.2f KB): %s\n", float64(info.Size())/1024, path)
			}
			return nil
		}

		// Check if the file is binary
		isText, err := isplaintextfile.FilePreview(path, effectiveConfig.MaxSizeKB)
		if err != nil {
			if *verbose {
				fmt.Printf("Error checking if file is binary: %s\n", path)
			}
			return nil
		}
		if !isText {
			if *verbose {
				fmt.Printf("Skipping binary file: %s\n", path)
			}
			return nil
		}

		// Read the content of the file using os.ReadFile
		content, err := os.ReadFile(path)
		if err != nil {
			if *verbose {
				fmt.Printf("Failed to read file: %s\n", path)
			}
			return nil
		}

		// Get the relative path of the file, ensuring it starts with "./"
		relPathForOutput, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(relPathForOutput, ".") {
			relPathForOutput = "./" + relPathForOutput
		}

		// Prepare the content to append using effective delimiter
		fileContent := fmt.Sprintf("\nFile: %s\n\n%s\n%s\n%s\n\n", relPathForOutput, effectiveConfig.Delimiter, content, effectiveConfig.Delimiter)
		fileSize := len(fileContent)

		// Check if the total size exceeds the 1MB limit
		if totalSize+fileSize > maxTotalSize {
			return fmt.Errorf("total output size exceeds 1MB limit; content not copied to the clipboard")
		}

		// Append the file path and content to the builder
		builder.WriteString(fileContent)
		totalSize += fileSize

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Copy the final content to the clipboard
	err = clipboard.WriteAll(builder.String())
	if err != nil {
		fmt.Println("Failed to copy to clipboard:", err)
		return
	}

	fmt.Println("Content copied to clipboard successfully.")
}

// matchesAnyPatternWithPath checks if the given name or relative path matches any pattern in the list.
// It returns true if a match is found. Errors are silently ignored (treated as no match).
func matchesAnyPatternWithPath(name, relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		// First try to match against basename
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}

		// Also try to match against relative path for path-based patterns
		if relPath != "" {
			matched, err = filepath.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}
			// Try matching without the ./ prefix
			trimmedPath := strings.TrimPrefix(relPath, "./")
			if trimmedPath != relPath {
				matched, err = filepath.Match(pattern, trimmedPath)
				if err == nil && matched {
					return true
				}
			}
		}
	}
	return false
}

// Helper function to parse comma-separated strings into a slice
func parseCommaSeparated(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
