// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegrationNestedConfigs tests the full traversal with nested .clip4llm files
func TestIntegrationNestedConfigs(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "clip4llm-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   .clip4llm (exclude=*.md)
	//   file1.txt
	//   file1.md
	//   docs/
	//     .clip4llm (include=README.md)
	//     README.md
	//     other.md
	//     guide.txt
	//   api/
	//     .clip4llm (max-size=128)
	//     handler.go
	//   backend/
	//     server.go

	// Create directories
	docsDir := filepath.Join(tmpDir, "docs")
	apiDir := filepath.Join(tmpDir, "api")
	backendDir := filepath.Join(tmpDir, "backend")

	for _, dir := range []string{docsDir, apiDir, backendDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	files := map[string]string{
		filepath.Join(tmpDir, "file1.txt"):       "root text file",
		filepath.Join(tmpDir, "file1.md"):        "root markdown file",
		filepath.Join(docsDir, "README.md"):      "docs readme",
		filepath.Join(docsDir, "other.md"):       "other docs",
		filepath.Join(docsDir, "guide.txt"):      "guide text",
		filepath.Join(apiDir, "handler.go"):      "api handler",
		filepath.Join(backendDir, "server.go"):   "backend server",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Create config files
	configs := map[string]string{
		filepath.Join(tmpDir, ".clip4llm"):    "exclude=*.md\n",
		filepath.Join(docsDir, ".clip4llm"):   "include=README.md\n",
		filepath.Join(apiDir, ".clip4llm"):    "max-size=128\n",
	}

	for path, content := range configs {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create config file %s: %v", path, err)
		}
	}

	// Simulate the traversal and collect processed files
	var processedFiles []string
	configStack := NewConfigStack(tmpDir, false)
	pushedDirs := make(map[string]bool)

	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Get effective config
		effectiveConfig := configStack.GetEffectiveConfig()

		// Get relative path
		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath == "." {
			relPath = ""
		}
		if relPath != "" && !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}

		// Check exclusions and inclusions (matching main.go behavior)
		excluded := matchesAnyPatternWithPath(name, relPath, effectiveConfig.Exclude)
		explicitlyIncluded := matchesAnyPatternWithPath(name, relPath, effectiveConfig.Include)
		shouldExclude := excluded && !explicitlyIncluded

		if shouldExclude {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle hidden files
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle directories
		if d.IsDir() {
			if path != tmpDir {
				if configStack.PushIfExists(path) {
					pushedDirs[path] = true
				}
			}
			return nil
		}

		// Process files
		processedFiles = append(processedFiles, relPath)
		return nil
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	// Verify results
	// Expected files:
	// - ./file1.txt (not excluded)
	// - ./docs/README.md (excluded by root *.md but rescued by docs include=README.md)
	// - ./docs/guide.txt (not excluded, not .md)
	// - ./api/handler.go (not excluded)
	// - ./backend/server.go (not excluded)
	// Not expected:
	// - ./file1.md (excluded by root *.md)
	// - ./docs/other.md (excluded by root *.md and not included)

	expectedFiles := []string{
		"./file1.txt",
		"./docs/README.md",
		"./docs/guide.txt",
		"./api/handler.go",
		"./backend/server.go",
	}

	unexpectedFiles := []string{
		"./file1.md",
		"./docs/other.md",
	}

	for _, expected := range expectedFiles {
		found := false
		for _, processed := range processedFiles {
			if processed == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %q to be processed, but it wasn't. Processed: %v", expected, processedFiles)
		}
	}

	for _, unexpected := range unexpectedFiles {
		for _, processed := range processedFiles {
			if processed == unexpected {
				t.Errorf("File %q should have been excluded but was processed", unexpected)
				break
			}
		}
	}
}

// TestIntegrationScopeContainment verifies that configs in one directory don't affect siblings
func TestIntegrationScopeContainment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-scope-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   frontend/
	//     .clip4llm (exclude=*.css)
	//     app.js
	//     styles.css
	//   backend/
	//     server.go
	//     styles.css  (should NOT be excluded since frontend's config doesn't apply)

	frontendDir := filepath.Join(tmpDir, "frontend")
	backendDir := filepath.Join(tmpDir, "backend")

	for _, dir := range []string{frontendDir, backendDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(frontendDir, "app.js"):      "frontend js",
		filepath.Join(frontendDir, "styles.css"):  "frontend css",
		filepath.Join(backendDir, "server.go"):    "backend go",
		filepath.Join(backendDir, "styles.css"):   "backend css",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Only frontend has a config that excludes *.css
	if err := os.WriteFile(filepath.Join(frontendDir, ".clip4llm"), []byte("exclude=*.css\n"), 0644); err != nil {
		t.Fatalf("Failed to create frontend config: %v", err)
	}

	// Track files processed with their directory context
	type fileWithContext struct {
		path    string
		cssExcluded bool
	}
	var results []fileWithContext

	configStack := NewConfigStack(tmpDir, false)
	pushedDirs := make(map[string]bool)

	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		effectiveConfig := configStack.GetEffectiveConfig()

		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath == "." {
			relPath = ""
		}
		if relPath != "" && !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}

		// Check exclusions
		excluded := matchesAnyPatternWithPath(name, relPath, effectiveConfig.Exclude)

		// Handle hidden files
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if path != tmpDir {
				if configStack.PushIfExists(path) {
					pushedDirs[path] = true
				}
			}
			return nil
		}

		if !excluded {
			results = append(results, fileWithContext{path: relPath, cssExcluded: false})
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	// Verify backend/styles.css is processed (not excluded)
	backendCSSProcessed := false
	for _, r := range results {
		if r.path == "./backend/styles.css" {
			backendCSSProcessed = true
			break
		}
	}

	if !backendCSSProcessed {
		t.Error("backend/styles.css should have been processed (frontend config shouldn't affect it)")
	}

	// Verify frontend/styles.css is NOT processed
	frontendCSSProcessed := false
	for _, r := range results {
		if r.path == "./frontend/styles.css" {
			frontendCSSProcessed = true
			break
		}
	}

	if frontendCSSProcessed {
		t.Error("frontend/styles.css should have been excluded by frontend's config")
	}
}

// TestIntegrationMaxSizePrecedence tests that nested max-size configs take precedence
func TestIntegrationMaxSizePrecedence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-maxsize-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   .clip4llm (max-size=32)
	//   api/
	//     .clip4llm (max-size=128)

	apiDir := filepath.Join(tmpDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatalf("Failed to create api directory: %v", err)
	}

	// Create configs
	if err := os.WriteFile(filepath.Join(tmpDir, ".clip4llm"), []byte("max-size=32\n"), 0644); err != nil {
		t.Fatalf("Failed to create root config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, ".clip4llm"), []byte("max-size=128\n"), 0644); err != nil {
		t.Fatalf("Failed to create api config: %v", err)
	}

	configStack := NewConfigStack(tmpDir, false)

	// Check root max-size
	rootConfig := configStack.GetEffectiveConfig()
	if rootConfig.MaxSizeKB != 32 {
		t.Errorf("Root MaxSizeKB: got %d, want 32", rootConfig.MaxSizeKB)
	}

	// Push api config and check max-size
	configStack.PushIfExists(apiDir)
	apiConfig := configStack.GetEffectiveConfig()
	if apiConfig.MaxSizeKB != 128 {
		t.Errorf("API MaxSizeKB: got %d, want 128", apiConfig.MaxSizeKB)
	}

	// Pop api config and verify back to root
	configStack.Pop(apiDir)
	afterPopConfig := configStack.GetEffectiveConfig()
	if afterPopConfig.MaxSizeKB != 32 {
		t.Errorf("After pop MaxSizeKB: got %d, want 32", afterPopConfig.MaxSizeKB)
	}
}

// TestIntegrationDeeplyNestedConfigs tests three levels of nested configs
func TestIntegrationDeeplyNestedConfigs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-deep-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   .clip4llm (max-size=32, exclude=*.log)
	//   a/
	//     .clip4llm (max-size=64)
	//     b/
	//       .clip4llm (max-size=128, exclude=*.tmp)
	//       c/
	//         file.txt

	aDir := filepath.Join(tmpDir, "a")
	bDir := filepath.Join(aDir, "b")
	cDir := filepath.Join(bDir, "c")

	for _, dir := range []string{aDir, bDir, cDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	configs := map[string]string{
		filepath.Join(tmpDir, ".clip4llm"): "max-size=32\nexclude=*.log\n",
		filepath.Join(aDir, ".clip4llm"):   "max-size=64\n",
		filepath.Join(bDir, ".clip4llm"):   "max-size=128\nexclude=*.tmp\n",
	}

	for path, content := range configs {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create config %s: %v", path, err)
		}
	}

	configStack := NewConfigStack(tmpDir, false)

	// Verify root
	config := configStack.GetEffectiveConfig()
	if config.MaxSizeKB != 32 {
		t.Errorf("Root MaxSizeKB: got %d, want 32", config.MaxSizeKB)
	}
	if len(config.Exclude) != 1 || config.Exclude[0] != "*.log" {
		t.Errorf("Root Exclude: got %v, want [*.log]", config.Exclude)
	}

	// Push a/
	configStack.PushIfExists(aDir)
	config = configStack.GetEffectiveConfig()
	if config.MaxSizeKB != 64 {
		t.Errorf("a/ MaxSizeKB: got %d, want 64", config.MaxSizeKB)
	}
	// Exclude should still have *.log (additive)
	if len(config.Exclude) != 1 || config.Exclude[0] != "*.log" {
		t.Errorf("a/ Exclude: got %v, want [*.log]", config.Exclude)
	}

	// Push a/b/
	configStack.PushIfExists(bDir)
	config = configStack.GetEffectiveConfig()
	if config.MaxSizeKB != 128 {
		t.Errorf("a/b/ MaxSizeKB: got %d, want 128", config.MaxSizeKB)
	}
	// Exclude should now have both *.log and *.tmp
	if len(config.Exclude) != 2 {
		t.Errorf("a/b/ Exclude: got %v, want [*.log, *.tmp]", config.Exclude)
	}
	hasLog := false
	hasTmp := false
	for _, e := range config.Exclude {
		if e == "*.log" {
			hasLog = true
		}
		if e == "*.tmp" {
			hasTmp = true
		}
	}
	if !hasLog || !hasTmp {
		t.Errorf("a/b/ Exclude: got %v, want [*.log, *.tmp]", config.Exclude)
	}

	// Pop back to a/
	configStack.Pop(bDir)
	config = configStack.GetEffectiveConfig()
	if config.MaxSizeKB != 64 {
		t.Errorf("After pop to a/ MaxSizeKB: got %d, want 64", config.MaxSizeKB)
	}

	// Pop back to root
	configStack.Pop(aDir)
	config = configStack.GetEffectiveConfig()
	if config.MaxSizeKB != 32 {
		t.Errorf("After pop to root MaxSizeKB: got %d, want 32", config.MaxSizeKB)
	}
}

// TestIntegrationNoRecursiveMode tests that --no-recursive prevents traversal into subdirectories
func TestIntegrationNoRecursiveMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-norecursive-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   .clip4llm (no-recursive=true)
	//   file.txt
	//   sub/
	//     .clip4llm (max-size=128) - should NOT be loaded
	//     nested.txt - should NOT be processed

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create sub directory: %v", err)
	}

	files := map[string]string{
		filepath.Join(tmpDir, "file.txt"):   "root file",
		filepath.Join(subDir, "nested.txt"): "nested file",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	configs := map[string]string{
		filepath.Join(tmpDir, ".clip4llm"): "no-recursive=true\n",
		filepath.Join(subDir, ".clip4llm"): "max-size=128\n",
	}

	for path, content := range configs {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create config %s: %v", path, err)
		}
	}

	configStack := NewConfigStack(tmpDir, false)
	var processedFiles []string

	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		effectiveConfig := configStack.GetEffectiveConfig()

		// Handle hidden files
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if path == tmpDir {
				return nil
			}
			// If no-recursive is set, skip subdirectories
			if effectiveConfig.NoRecursive {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(tmpDir, path)
		if !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}
		processedFiles = append(processedFiles, relPath)
		return nil
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	// Should only process ./file.txt
	if len(processedFiles) != 1 {
		t.Errorf("Expected 1 file, got %d: %v", len(processedFiles), processedFiles)
	}

	if len(processedFiles) > 0 && processedFiles[0] != "./file.txt" {
		t.Errorf("Expected ./file.txt, got %s", processedFiles[0])
	}

	// Verify nested config was never loaded
	_, exists := configStack.cache[subDir]
	if exists {
		t.Error("Nested config should not have been checked when no-recursive is enabled")
	}
}

// TestIntegrationCLIFlagsWin tests that CLI flags always override config file settings
func TestIntegrationCLIFlagsWin(t *testing.T) {
	// This test verifies that CLI exclude patterns override config include patterns
	// as specified in acceptance criteria #5

	tmpDir, err := os.MkdirTemp("", "clip4llm-cli-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure:
	// tmpDir/
	//   .clip4llm (include=*.go)  <- config tries to include Go files
	//   app.go                    <- should be excluded by CLI --exclude=*.go
	//   main.txt                  <- should be included

	files := map[string]string{
		filepath.Join(tmpDir, "app.go"):    "go code",
		filepath.Join(tmpDir, "main.txt"):  "text file",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	// Config tries to include Go files
	if err := os.WriteFile(filepath.Join(tmpDir, ".clip4llm"), []byte("include=*.go\n"), 0644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	configStack := NewConfigStack(tmpDir, false)

	// Simulate CLI exclude pattern (--exclude=*.go)
	cliExcludePatterns := []string{"*.go"}
	var processedFiles []string

	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		effectiveConfig := configStack.GetEffectiveConfig()

		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath == "." {
			relPath = ""
		}
		if relPath != "" && !strings.HasPrefix(relPath, ".") {
			relPath = "./" + relPath
		}

		// Apply the same logic as main.go
		excludePatterns := append([]string{}, effectiveConfig.Exclude...)
		excludePatterns = append(excludePatterns, cliExcludePatterns...)

		excluded := matchesAnyPatternWithPath(name, relPath, excludePatterns)
		explicitlyIncluded := matchesAnyPatternWithPath(name, relPath, effectiveConfig.Include)

		// CLI exclude patterns always win
		cliExcluded := matchesAnyPatternWithPath(name, relPath, cliExcludePatterns)
		shouldExclude := excluded && (!explicitlyIncluded || cliExcluded)

		// Handle hidden files
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldExclude {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		processedFiles = append(processedFiles, relPath)
		return nil
	})

	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}

	// app.go should be excluded (CLI wins over config include)
	for _, f := range processedFiles {
		if f == "./app.go" {
			t.Error("app.go should have been excluded by CLI --exclude=*.go")
		}
	}

	// main.txt should be included
	foundTxt := false
	for _, f := range processedFiles {
		if f == "./main.txt" {
			foundTxt = true
			break
		}
	}
	if !foundTxt {
		t.Error("main.txt should have been included")
	}
}

// printTestDirectoryTree is a helper to visualize directory structure (for debugging)
func printTestDirectoryTree(root string) {
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		indent := strings.Repeat("  ", strings.Count(rel, string(os.PathSeparator)))
		if d.IsDir() {
			fmt.Printf("%s[%s/]\n", indent, d.Name())
		} else {
			fmt.Printf("%s%s\n", indent, d.Name())
		}
		return nil
	})
}
