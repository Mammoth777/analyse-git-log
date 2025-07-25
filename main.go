package main

import (
	"bufio"
	"fmt"
	"git-log-analyzer/cmd"
	"os"
	"strings"
)

func main() {
	// Load environment variables from .env file if it exists
	loadEnvFile()
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Try to find .env file in current directory
	envPaths := []string{
		".env",           // Current directory
	}
	
	var envFile *os.File
	var err error
	
	for _, path := range envPaths {
		envFile, err = os.Open(path)
		if err == nil {
			break
		}
	}
	
	if envFile == nil {
		// No .env file found, that's okay
		return
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		
		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
			// fmt.Println("Loaded env:", key, "=", value)
		}
	}
}
