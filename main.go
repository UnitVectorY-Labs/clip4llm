// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/UnitVectorY-Labs/isplaintextfile"
	"github.com/atotto/clipboard"
)

// Define the max total size limit in bytes (1MB = 1,048,576 bytes)
const maxTotalSize = 1 * 1024 * 1024 // 1MB in bytes

var Version = "dev" // This will be set by the build systems to the release version

func main() {
	// Define existing flags
	delimiter := flag.String("delimiter", "```", "Set the delimiter for file content (default: ```)")
	maxSize := flag.Int("max-size", 32, "Maximum file size to include in KB (default: 32 KB)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	// Define new flags for include and exclude with support for wildcards
	include := flag.String("include", "", "Comma-separated list of patterns to include, even if hidden (e.g., .github,*.env)")
	exclude := flag.String("exclude", "", "Comma-separated list of patterns to exclude (e.g., LICENSE,*.md)")

	showVersion := flag.Bool("version", false, "Print version")

	flag.Parse()

	if *showVersion {
		fmt.Println("Version:", Version)
		return
	}

	// Load configuration from .clip4llm files
	config := loadConfig(*verbose)

	// Determine if flags were set by the user
	delimiterSet := false
	maxSizeSet := false
	includeSetFlag := false
	excludeSetFlag := false

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
	})

	// Override flag values with config values if the flag was not set by the user
	if !delimiterSet {
		if val, ok := config["delimiter"]; ok {
			*delimiter = val
		}
	}

	if !maxSizeSet {
		if val, ok := config["max-size"]; ok {
			if parsedVal, err := strconv.Atoi(val); err == nil {
				*maxSize = parsedVal
			}
		}
	}

	if !includeSetFlag {
		if val, ok := config["include"]; ok {
			*include = val
		}
	}

	if !excludeSetFlag {
		if val, ok := config["exclude"]; ok {
			*exclude = val
		}
	}

	// Parse include and exclude patterns from flags or config
	var includePatterns []string
	if *include != "" {
		includePatterns = parseCommaSeparated(*include)
	} else if val, ok := config["include"]; ok {
		includePatterns = parseCommaSeparated(val)
	}

	var excludePatterns []string
	if *exclude != "" {
		excludePatterns = parseCommaSeparated(*exclude)
	}

	if *verbose {
		// Print out the configuration values
		fmt.Println("Configuration:")
		fmt.Printf("\tDelimiter: %s\n", *delimiter)
		fmt.Printf("\tMax Size: %d KB\n", *maxSize)
		fmt.Printf("\tInclude Patterns: %v\n", includePatterns)
		fmt.Printf("\tExclude Patterns: %v\n", excludePatterns)
	}

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var builder strings.Builder
	totalSize := 0 // Track total size of the output

	// Walk through the current folder and process files
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the base name of the file/directory
		name := info.Name()

		// Check if the file/directory matches any exclude patterns
		excluded, err := matchesAnyPattern(name, excludePatterns)
		if err != nil {
			if *verbose {
				fmt.Printf("Error matching exclude patterns for %s: %v\n", path, err)
			}
			// In case of error, do not exclude
			excluded = false
		}
		if excluded {
			if info.IsDir() {
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
			included, err := matchesAnyPattern(name, includePatterns)
			if err != nil {
				if *verbose {
					fmt.Printf("Error matching include patterns for %s: %v\n", path, err)
				}
				// In case of error, do not include
				included = false
			}

			if !included {
				if *verbose {
					fmt.Printf("Skipping hidden file/directory: %s\n", path)
				}
				if info.IsDir() {
					return filepath.SkipDir // Skip the entire hidden directory
				}
				return nil // Skip the hidden file
			}
			// If the hidden file/directory is in the include patterns, proceed
			if *verbose {
				fmt.Printf("Including hidden file/directory (matched include pattern): %s\n", path)
			}
		}

		// If it's a directory (and not skipped), continue traversing
		if info.IsDir() {
			if *verbose {
				fmt.Printf("Entering directory: %s\n", path)
			}
			return nil
		}

		// Skip files larger than the specified max size
		maxSizeBytes := int64(*maxSize) * 1024
		if info.Size() > maxSizeBytes {
			if *verbose {
				fmt.Printf("Skipping large file (%.2f KB): %s\n", float64(info.Size())/1024, path)
			}
			return nil
		}

		// Check if the file is binary
		isText, err := isplaintextfile.FilePreview(path, *maxSize)
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
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}

		// Prepare the content to append
		fileContent := fmt.Sprintf("\nFile: %s\n\n%s\n%s\n%s\n\n", relPath, *delimiter, content, *delimiter)
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

// matchesAnyPattern checks if the given name matches any pattern in the list.
// It returns true if a match is found.
func matchesAnyPattern(name string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
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
