// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the typed configuration for clip4llm
type Config struct {
	Delimiter   string
	MaxSizeKB   int
	NoRecursive bool
	Include     []string
	Exclude     []string
}

// ConfigLayer represents a parsed configuration from a specific directory
type ConfigLayer struct {
	Directory string
	Config    *Config
}

// ConfigStack manages the stack of configuration layers during traversal
type ConfigStack struct {
	layers    []*ConfigLayer
	cache     map[string]*Config // Cache parsed configs by directory path
	verbose   bool
	rootDir   string
	homeDir   string
}

// NewConfigStack creates a new ConfigStack with home and root configs preloaded
func NewConfigStack(rootDir string, verbose bool) *ConfigStack {
	cs := &ConfigStack{
		layers:  make([]*ConfigLayer, 0),
		cache:   make(map[string]*Config),
		verbose: verbose,
		rootDir: rootDir,
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			log.Printf("Error getting home directory: %v", err)
		}
	} else {
		cs.homeDir = homeDir
		homeConfigPath := filepath.Join(homeDir, ".clip4llm")
		if cfg := cs.loadConfigFromFile(homeConfigPath); cfg != nil {
			cs.layers = append(cs.layers, &ConfigLayer{
				Directory: homeDir,
				Config:    cfg,
			})
		}
	}

	// Load root directory config if different from home
	if rootDir != homeDir {
		rootConfigPath := filepath.Join(rootDir, ".clip4llm")
		if cfg := cs.loadConfigFromFile(rootConfigPath); cfg != nil {
			cs.layers = append(cs.layers, &ConfigLayer{
				Directory: rootDir,
				Config:    cfg,
			})
		}
	}

	return cs
}

// PushIfExists checks if a .clip4llm file exists in the given directory,
// and if so, parses and pushes it onto the stack. Returns true if pushed.
func (cs *ConfigStack) PushIfExists(dir string) bool {
	// Don't re-push root or home directory configs
	if dir == cs.rootDir || dir == cs.homeDir {
		return false
	}

	configPath := filepath.Join(dir, ".clip4llm")

	// Check cache first
	if _, cached := cs.cache[dir]; cached {
		cfg := cs.cache[dir]
		if cfg != nil {
			cs.layers = append(cs.layers, &ConfigLayer{
				Directory: dir,
				Config:    cfg,
			})
			if cs.verbose {
				fmt.Printf("Loading scoped config from (cached): %s\n", configPath)
			}
			return true
		}
		return false
	}

	// Try to load the config
	cfg := cs.loadConfigFromFile(configPath)
	cs.cache[dir] = cfg // Cache even if nil (no config exists)

	if cfg != nil {
		cs.layers = append(cs.layers, &ConfigLayer{
			Directory: dir,
			Config:    cfg,
		})
		if cs.verbose {
			fmt.Printf("Loading scoped config from: %s\n", configPath)
		}
		return true
	}

	return false
}

// Pop removes the top layer from the stack if it matches the given directory
func (cs *ConfigStack) Pop(dir string) {
	if len(cs.layers) > 0 {
		top := cs.layers[len(cs.layers)-1]
		if top.Directory == dir {
			cs.layers = cs.layers[:len(cs.layers)-1]
			if cs.verbose {
				fmt.Printf("Leaving scoped config from: %s\n", dir)
			}
		}
	}
}

// GetEffectiveConfig computes the effective configuration by merging all layers
func (cs *ConfigStack) GetEffectiveConfig() *Config {
	effective := &Config{
		Delimiter:   "```", // Default values
		MaxSizeKB:   32,
		NoRecursive: false,
		Include:     []string{},
		Exclude:     []string{},
	}

	for _, layer := range cs.layers {
		cs.mergeConfig(effective, layer.Config)
	}

	return effective
}

// mergeConfig merges source config into target config
// - Scalar values (Delimiter, MaxSizeKB, NoRecursive): last writer wins
// - List values (Include, Exclude): additive concatenation
func (cs *ConfigStack) mergeConfig(target, source *Config) {
	if source == nil {
		return
	}

	if source.Delimiter != "" {
		target.Delimiter = source.Delimiter
	}
	if source.MaxSizeKB > 0 {
		target.MaxSizeKB = source.MaxSizeKB
	}
	if source.NoRecursive {
		target.NoRecursive = true
	}
	if len(source.Include) > 0 {
		target.Include = append(target.Include, source.Include...)
	}
	if len(source.Exclude) > 0 {
		target.Exclude = append(target.Exclude, source.Exclude...)
	}
}

// loadConfigFromFile loads and parses a .clip4llm file into a Config struct
func (cs *ConfigStack) loadConfigFromFile(path string) *Config {
	if cs.verbose {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Config file exists: %s\n", path)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		// It's OK if the file doesn't exist
		if !os.IsNotExist(err) {
			if cs.verbose {
				log.Printf("Error reading config file %s: %v", path, err)
			}
		}
		return nil
	}
	defer file.Close()

	config := &Config{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// Skip empty lines and comments
			continue
		}
		// Expect lines in the format "key=value"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "delimiter":
				config.Delimiter = value
			case "max-size":
				if parsedVal, err := strconv.Atoi(value); err == nil {
					config.MaxSizeKB = parsedVal
				}
			case "no-recursive":
				config.NoRecursive = value == "true"
			case "include":
				config.Include = parseCommaSeparated(value)
			case "exclude":
				config.Exclude = parseCommaSeparated(value)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if cs.verbose {
			log.Printf("Error scanning config file %s: %v", path, err)
		}
	}

	return config
}
