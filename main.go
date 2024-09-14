package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
)

// Function to determine if a file is likely plain text or binary
func isBinaryFile(path string) (bool, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read the first 512 bytes to analyze the file content
	reader := bufio.NewReader(file)
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil {
		return false, err
	}

	// Check for non-printable characters
	for i := 0; i < n; i++ {
		// If we encounter a non-ASCII or non-printable character, treat it as binary
		if buffer[i] > unicode.MaxASCII || (buffer[i] < 32 && buffer[i] != '\n' && buffer[i] != '\r' && buffer[i] != '\t') {
			return true, nil
		}
	}
	// Assume it's a text file if no binary-like content is found
	return false, nil
}

func main() {
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
				log.Printf("Skipping directory: %s\n", path)
				return filepath.SkipDir // Skip the entire directory
			}
			log.Printf("Skipping file: %s\n", path)
			return nil // Skip the file
		}

		// If it's a directory, continue traversing
		if info.IsDir() {
			log.Printf("Entering directory: %s\n", path)
			return nil
		}

		// Check if the file is binary
		isBinary, err := isBinaryFile(path)
		if err != nil {
			log.Printf("Error checking if file is binary: %s\n", path)
			return nil
		}
		if isBinary {
			log.Printf("Skipping binary file: %s\n", path)
			return nil
		}

		// Read the content of the file using os.ReadFile
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file: %s\n", path)
			return nil
		}

		// Get the relative path of the file
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Append the file path and content to the builder with triple backticks
		builder.WriteString(fmt.Sprintf("File: %s\n```\n%s\n```\n\n", relPath, content))
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Copy the final content to the clipboard
	err = clipboard.WriteAll(builder.String())
	if err != nil {
		log.Fatal("Failed to copy to clipboard:", err)
	}

	fmt.Println("Content copied to clipboard successfully.")
}
