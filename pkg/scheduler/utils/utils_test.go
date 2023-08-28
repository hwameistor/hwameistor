package utils

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file to ensure it exists
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	testCases := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			"FileExists_ExistingFile", // Test for a file that exists
			tmpfile.Name(),            // Path to temporary file
			true,                      // We expect the function to return true
		},
		{
			"FileExists_NonExistentFile", // Test for a file that does not exist
			"/non/existent/path",         // Path to non-existent file
			false,                        // We expect the function to return false
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists := FileExists(tc.filePath)
			if exists != tc.expected {
				t.Errorf("FileExists(%s) = %t; want %t", tc.filePath, exists, tc.expected)
			}
		})
	}
}
