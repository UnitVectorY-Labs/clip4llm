// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigMerge(t *testing.T) {
	tests := []struct {
		name     string
		target   *Config
		source   *Config
		expected *Config
	}{
		{
			name: "merge delimiter",
			target: &Config{
				Delimiter: "```",
				MaxSizeKB: 32,
			},
			source: &Config{
				Delimiter: "---",
			},
			expected: &Config{
				Delimiter: "---",
				MaxSizeKB: 32,
			},
		},
		{
			name: "merge max-size",
			target: &Config{
				Delimiter: "```",
				MaxSizeKB: 32,
			},
			source: &Config{
				MaxSizeKB: 128,
			},
			expected: &Config{
				Delimiter: "```",
				MaxSizeKB: 128,
			},
		},
		{
			name: "merge no-recursive",
			target: &Config{
				NoRecursive: false,
			},
			source: &Config{
				NoRecursive: true,
			},
			expected: &Config{
				NoRecursive: true,
			},
		},
		{
			name: "additive include patterns",
			target: &Config{
				Include: []string{".github", "*.env"},
			},
			source: &Config{
				Include: []string{".config"},
			},
			expected: &Config{
				Include: []string{".github", "*.env", ".config"},
			},
		},
		{
			name: "additive exclude patterns",
			target: &Config{
				Exclude: []string{"*.md", "LICENSE"},
			},
			source: &Config{
				Exclude: []string{"*.txt"},
			},
			expected: &Config{
				Exclude: []string{"*.md", "LICENSE", "*.txt"},
			},
		},
		{
			name: "nil source does not change target",
			target: &Config{
				Delimiter: "```",
				MaxSizeKB: 32,
				Include:   []string{".github"},
				Exclude:   []string{"*.md"},
			},
			source: nil,
			expected: &Config{
				Delimiter: "```",
				MaxSizeKB: 32,
				Include:   []string{".github"},
				Exclude:   []string{"*.md"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &ConfigStack{}
			cs.mergeConfig(tt.target, tt.source)

			if tt.target.Delimiter != tt.expected.Delimiter {
				t.Errorf("Delimiter: got %q, want %q", tt.target.Delimiter, tt.expected.Delimiter)
			}
			if tt.target.MaxSizeKB != tt.expected.MaxSizeKB {
				t.Errorf("MaxSizeKB: got %d, want %d", tt.target.MaxSizeKB, tt.expected.MaxSizeKB)
			}
			if tt.target.NoRecursive != tt.expected.NoRecursive {
				t.Errorf("NoRecursive: got %v, want %v", tt.target.NoRecursive, tt.expected.NoRecursive)
			}
			if !slicesEqual(tt.target.Include, tt.expected.Include) {
				t.Errorf("Include: got %v, want %v", tt.target.Include, tt.expected.Include)
			}
			if !slicesEqual(tt.target.Exclude, tt.expected.Exclude) {
				t.Errorf("Exclude: got %v, want %v", tt.target.Exclude, tt.expected.Exclude)
			}
		})
	}
}

func TestMatchesAnyPatternWithPath(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		relPath  string
		patterns []string
		expected bool
	}{
		{
			name:     "match basename with wildcard",
			filename: "test.md",
			relPath:  "./docs/test.md",
			patterns: []string{"*.md"},
			expected: true,
		},
		{
			name:     "no match",
			filename: "test.go",
			relPath:  "./src/test.go",
			patterns: []string{"*.md"},
			expected: false,
		},
		{
			name:     "match exact basename",
			filename: "LICENSE",
			relPath:  "./LICENSE",
			patterns: []string{"LICENSE"},
			expected: true,
		},
		{
			name:     "match relative path",
			filename: "file.txt",
			relPath:  "./docs/file.txt",
			patterns: []string{"docs/*"},
			expected: true,
		},
		{
			name:     "match without leading dot slash",
			filename: "file.txt",
			relPath:  "./docs/file.txt",
			patterns: []string{"docs/file.txt"},
			expected: true,
		},
		{
			name:     "empty patterns list",
			filename: "test.md",
			relPath:  "./test.md",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "match hidden file by name",
			filename: ".gitignore",
			relPath:  "./.gitignore",
			patterns: []string{".gitignore"},
			expected: true,
		},
		{
			name:     "match hidden directory",
			filename: ".github",
			relPath:  "./.github",
			patterns: []string{".github"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAnyPatternWithPath(tt.filename, tt.relPath, tt.patterns)
			if result != tt.expected {
				t.Errorf("matchesAnyPatternWithPath(%q, %q, %v) = %v, want %v",
					tt.filename, tt.relPath, tt.patterns, result, tt.expected)
			}
		})
	}
}

func TestConfigStackWithNestedConfigs(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "clip4llm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	docsDir := filepath.Join(tmpDir, "docs")
	apiDir := filepath.Join(tmpDir, "api")
	apiV2Dir := filepath.Join(apiDir, "v2")

	for _, dir := range []string{docsDir, apiDir, apiV2Dir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create root config
	rootConfig := `exclude=*.md
max-size=32
delimiter=---
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".clip4llm"), []byte(rootConfig), 0644); err != nil {
		t.Fatalf("Failed to write root config: %v", err)
	}

	// Create docs config that includes README.md
	docsConfig := `include=README.md
`
	if err := os.WriteFile(filepath.Join(docsDir, ".clip4llm"), []byte(docsConfig), 0644); err != nil {
		t.Fatalf("Failed to write docs config: %v", err)
	}

	// Create api config with different max-size
	apiConfig := `max-size=128
`
	if err := os.WriteFile(filepath.Join(apiDir, ".clip4llm"), []byte(apiConfig), 0644); err != nil {
		t.Fatalf("Failed to write api config: %v", err)
	}

	// Test 1: Create config stack for root
	cs := NewConfigStack(tmpDir, false)
	effectiveConfig := cs.GetEffectiveConfig()

	if effectiveConfig.MaxSizeKB != 32 {
		t.Errorf("Root MaxSizeKB: got %d, want 32", effectiveConfig.MaxSizeKB)
	}
	if effectiveConfig.Delimiter != "---" {
		t.Errorf("Root Delimiter: got %q, want '---'", effectiveConfig.Delimiter)
	}

	// Test 2: Push docs config and verify additive include
	if !cs.PushIfExists(docsDir) {
		t.Error("Expected to push docs config")
	}
	effectiveConfig = cs.GetEffectiveConfig()

	if !containsPattern(effectiveConfig.Include, "README.md") {
		t.Errorf("Docs Include should contain README.md, got %v", effectiveConfig.Include)
	}
	if !containsPattern(effectiveConfig.Exclude, "*.md") {
		t.Errorf("Docs Exclude should still contain *.md, got %v", effectiveConfig.Exclude)
	}

	// Pop docs and push api
	cs.Pop(docsDir)
	if !cs.PushIfExists(apiDir) {
		t.Error("Expected to push api config")
	}
	effectiveConfig = cs.GetEffectiveConfig()

	if effectiveConfig.MaxSizeKB != 128 {
		t.Errorf("API MaxSizeKB: got %d, want 128", effectiveConfig.MaxSizeKB)
	}
	// Delimiter should still be from root
	if effectiveConfig.Delimiter != "---" {
		t.Errorf("API Delimiter: got %q, want '---'", effectiveConfig.Delimiter)
	}
}

func TestConfigStackDoesNotRepushRootOrHome(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create root config
	if err := os.WriteFile(filepath.Join(tmpDir, ".clip4llm"), []byte("max-size=64\n"), 0644); err != nil {
		t.Fatalf("Failed to write root config: %v", err)
	}

	cs := NewConfigStack(tmpDir, false)
	initialLayerCount := len(cs.layers)

	// Try to push root again - should not add another layer
	pushed := cs.PushIfExists(tmpDir)
	if pushed {
		t.Error("Should not re-push root directory config")
	}
	if len(cs.layers) != initialLayerCount {
		t.Errorf("Layer count changed: got %d, want %d", len(cs.layers), initialLayerCount)
	}
}

func TestConfigStackCaching(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create config in subdir
	if err := os.WriteFile(filepath.Join(subDir, ".clip4llm"), []byte("max-size=64\n"), 0644); err != nil {
		t.Fatalf("Failed to write subdir config: %v", err)
	}

	cs := NewConfigStack(tmpDir, false)

	// Push subdirectory config
	cs.PushIfExists(subDir)
	cs.Pop(subDir)

	// After popping, the cache should still have the parsed config
	cachedConfig, exists := cs.cache[subDir]
	if !exists {
		t.Error("Expected cached config to exist after pop")
	}
	if cachedConfig == nil {
		t.Error("Expected cached config to not be nil")
	}
	if cachedConfig.MaxSizeKB != 64 {
		t.Errorf("Cached MaxSizeKB: got %d, want 64", cachedConfig.MaxSizeKB)
	}
}

func TestConfigStackNonExistentConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clip4llm-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Don't create a config file in subdir

	cs := NewConfigStack(tmpDir, false)
	initialLayerCount := len(cs.layers)

	// Try to push config for directory without .clip4llm
	pushed := cs.PushIfExists(subDir)
	if pushed {
		t.Error("Should not push when no .clip4llm exists")
	}
	if len(cs.layers) != initialLayerCount {
		t.Errorf("Layer count changed: got %d, want %d", len(cs.layers), initialLayerCount)
	}

	// Cache should record that no config exists
	cachedConfig, exists := cs.cache[subDir]
	if !exists {
		t.Error("Expected cache entry to exist (even for nil)")
	}
	if cachedConfig != nil {
		t.Error("Expected cached config to be nil for non-existent file")
	}
}

func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{"  a  ,  b  ,  c  ", []string{"a", "b", "c"}},
		{"*.md,*.txt", []string{"*.md", "*.txt"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{"a,,b", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseCommaSeparated(tt.input)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("parseCommaSeparated(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper functions for tests
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsPattern(patterns []string, pattern string) bool {
	for _, p := range patterns {
		if p == pattern {
			return true
		}
	}
	return false
}
