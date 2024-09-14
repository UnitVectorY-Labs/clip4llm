package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
)

func main() {
	// Define the --delimiter flag with a default value of triple backticks
	delimiter := flag.String("delimiter", "```", "Set the delimiter for file content (default: ```)")
	maxSize := flag.Int("max-size", 32, "Maximum file size to include in KB (default: 32 KB)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	// Call loadConfig to load configuration from .clip4llm file
	config := loadConfig(*verbose)

	// Check if the flags were set by the user
	delimiterSet := false
	maxSizeSet := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "delimiter" {
			delimiterSet = true
		}
		if f.Name == "max-size" {
			maxSizeSet = true
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

	if *verbose {
		// Print out the configuration values
		fmt.Println("Configuration:")
		fmt.Printf("\tDelimiter: %s\n", *delimiter)
		fmt.Printf("\tMax Size: %d KB\n", *maxSize)
	}

	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var builder strings.Builder

	// Walk through the current folder and process files
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip files and directories that begin with a dot
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				if *verbose {
					// Print out the configuration values
					fmt.Printf("Skipping directory: %s\n", path)
				}
				return filepath.SkipDir // Skip the entire directory
			}
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Skipping file: %s\n", path)
			}
			return nil // Skip the file
		}

		// If it's a directory, continue traversing
		if info.IsDir() {
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Entering directory: %s\n", path)
			}
			return nil
		}

		// Skip files larger than the specified max size
		maxSizeBytes := int64(*maxSize) * 1024
		if info.Size() > maxSizeBytes {
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Skipping large file (%.2f KB): %s\n", float64(info.Size())/1024, path)
			}
			return nil
		}

		// Check if the file is binary
		isBinary, err := isBinaryFile(path, *maxSize)
		if err != nil {
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Error checking if file is binary: %s\n", path)
			}
			return nil
		}
		if isBinary {
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Skipping binary file: %s\n", path)
			}
			return nil
		}

		// Read the content of the file using os.ReadFile
		content, err := os.ReadFile(path)
		if err != nil {
			if *verbose {
				// Print out the configuration values
				fmt.Printf("Failed to read file: %s\n", path)
			}
			return nil
		}

		// Get the relative path of the file, ensuring it starts with "./"
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(relPath, "./") {
			relPath = "./" + relPath
		}

		// Append the file path and content to the builder with the specified delimiter
		builder.WriteString(fmt.Sprintf("\nFile: %s\n\n%s\n%s\n%s\n\n", relPath, *delimiter, content, *delimiter))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Copy the final content to the clipboard
	err = clipboard.WriteAll(builder.String())
	if err != nil {
		fmt.Println("Failed to copy to clipboard:", err)
	}

	fmt.Println("Content copied to clipboard successfully.")
}
