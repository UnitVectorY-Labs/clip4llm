package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Helper function to find and load the .clip4llm file from home or current directory
func loadConfig(verbose bool) map[string]string {
	config := make(map[string]string)

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if verbose {
			log.Printf("Error getting home directory: %v", err)
		}
	} else {
		homeConfigPath := filepath.Join(homeDir, ".clip4llm")
		loadConfigFromFile(homeConfigPath, config, verbose)
	}

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		if verbose {
			log.Printf("Error getting current directory: %v", err)
		}
	} else {
		currentConfigPath := filepath.Join(currentDir, ".clip4llm")
		loadConfigFromFile(currentConfigPath, config, verbose)
	}

	return config
}

// Helper function to load configuration from a file and add to the config map
func loadConfigFromFile(path string, config map[string]string, verbose bool) {
	if verbose {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Config file exists: %s\n", path)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		// It's OK if the file doesn't exist
		if !os.IsNotExist(err) {
			if verbose {
				log.Printf("Error reading config file %s: %v", path, err)
			}
		}
		return
	}
	defer file.Close()

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
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		if verbose {
			log.Printf("Error scanning config file %s: %v", path, err)
		}
	}
}
