package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	osClipboard "github.com/atotto/clipboard"
)

// clipboardFile is the path to the JSON file for persistent storage
const clipboardFile = "clipboard.json"

// clipboardData represents the in-memory storage for key-value pairs
// Key design decision: Using map for O(1) lookup performance, persisted to JSON
var clipboardData = make(map[string]string)

// loadClipboard reads the clipboard data from JSON file
// If the file doesn't exist, it initializes with an empty clipboard
// Important: This function is called at startup to restore persisted data
func loadClipboard() error {
	file, err := os.Open(clipboardFile)
	if err != nil {
		// File doesn't exist, start with empty clipboard
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open clipboard file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&clipboardData)
	if err != nil {
		return fmt.Errorf("failed to decode clipboard data: %v", err)
	}

	return nil
}

// saveClipboard writes the current clipboard data to JSON file
// Important: This function is called after every modification to persist data
func saveClipboard() error {
	file, err := os.Create(clipboardFile)
	if err != nil {
		return fmt.Errorf("failed to create clipboard file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(clipboardData)
	if err != nil {
		return fmt.Errorf("failed to encode clipboard data: %v", err)
	}

	return nil
}

// printUsage displays the correct usage of the CLI tool
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  clipboard add <key> <value>     - Store a value with the given key")
	fmt.Println("  clipboard retrieve <key>        - Retrieve and display the value")
	fmt.Println("  clipboard copy <key>            - Retrieve value and copy to OS clipboard")
	fmt.Println("  clipboard list                  - List all stored keys")
	fmt.Println("\nExamples:")
	fmt.Println("  clipboard add mykey \"Hello World\"")
	fmt.Println("  clipboard retrieve mykey")
	fmt.Println("  clipboard copy mykey")
}

// addToClipboard stores a key-value pair in the clipboard and persists to file
// Args:
//   - key: The identifier for the stored value
//   - value: The string content to store
//
// Important: Keys are case-sensitive and will overwrite existing values
// Data is automatically saved to JSON file for persistence
func addToClipboard(key, value string) {
	if key == "" {
		fmt.Println("Error: Key cannot be empty")
		return
	}
	
	clipboardData[key] = value
	
	// Save to file for persistence
	if err := saveClipboard(); err != nil {
		fmt.Printf("Error saving clipboard: %v\n", err)
		return
	}
	
	fmt.Printf("Added '%s' with key '%s'\n", value, key)
}

// retrieveFromClipboard gets a value from the clipboard using its key
// Args:
//   - key: The identifier for the value to retrieve
//   - copyToOS: Whether to copy the value to the OS clipboard
//
// Returns the stored value or an error message if key doesn't exist
// Important: Automatically copies retrieved value to OS clipboard for easy pasting
func retrieveFromClipboard(key string, copyToOS bool) {
	if key == "" {
		fmt.Println("Error: Key cannot be empty")
		return
	}
	
	value, exists := clipboardData[key]
	if !exists {
		fmt.Printf("Error: No value found for key '%s'\n", key)
		return
	}
	
	// Copy to OS clipboard if requested
	if copyToOS {
		err := osClipboard.WriteAll(value)
		if err != nil {
			fmt.Printf("Warning: Failed to copy to OS clipboard: %v\n", err)
		} else {
			fmt.Printf("Copied to OS clipboard: %s\n", value)
			return
		}
	}
	
	fmt.Println(value)
}

// listAllKeys displays all stored keys in the clipboard
// Useful for debugging and seeing what's currently stored
func listAllKeys() {
	if len(clipboardData) == 0 {
		fmt.Println("Clipboard is empty")
		return
	}
	
	fmt.Println("Stored keys:")
	for key := range clipboardData {
		fmt.Printf("  - %s\n", key)
	}
}

// main is the entry point of the CLI application
// Handles command-line argument parsing and delegates to appropriate functions
// Important: Loads persisted data from JSON file at startup
func main() {
	// Load existing clipboard data from file
	if err := loadClipboard(); err != nil {
		fmt.Printf("Error loading clipboard: %v\n", err)
		os.Exit(1)
	}
	
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	command := strings.ToLower(os.Args[1])
	
	switch command {
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Error: 'add' command requires both key and value")
			fmt.Println("Usage: clipboard add <key> <value>")
			os.Exit(1)
		}
		
		key := os.Args[2]
		// Join remaining arguments to support values with spaces
		value := strings.Join(os.Args[3:], " ")
		addToClipboard(key, value)
		
	case "retrieve":
		if len(os.Args) < 3 {
			fmt.Println("Error: 'retrieve' command requires a key")
			fmt.Println("Usage: clipboard retrieve <key>")
			os.Exit(1)
		}
		
		key := os.Args[2]
		retrieveFromClipboard(key, false)
		
	case "copy":
		if len(os.Args) < 3 {
			fmt.Println("Error: 'copy' command requires a key")
			fmt.Println("Usage: clipboard copy <key>")
			os.Exit(1)
		}
		
		key := os.Args[2]
		retrieveFromClipboard(key, true)
		
	case "list":
		listAllKeys()
		
	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}