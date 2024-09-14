// Copyright (c) 2024 UnitVectorY Labs
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
package main

import (
	"bufio"
	"os"
	"unicode"
)

// Function to determine if a file is likely plain text or binary
func isBinaryFile(path string, maxKB int) (bool, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Calculate the maximum number of bytes to read
	maxBytes := maxKB * 1024

	// Read up to maxBytes to analyze the file content
	reader := bufio.NewReader(file)
	buffer := make([]byte, maxBytes)
	n, err := reader.Read(buffer)
	if err != nil {
		return false, err
	}

	// Check for non-printable characters
	for i := 0; n > 0 && i < n; i++ {
		// If we encounter a non-ASCII or non-printable character, treat it as binary
		if buffer[i] > unicode.MaxASCII || (buffer[i] < 32 && buffer[i] != '\n' && buffer[i] != '\r' && buffer[i] != '\t') {
			return true, nil
		}
	}
	// Assume it's a text file if no binary-like content is found
	return false, nil
}
