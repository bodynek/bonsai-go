package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Running Bonsai Tests...")

	// Example: Check health endpoint
	err := checkHealth()
	if err != nil {
		log.Fatalf("Health check failed: %v\n", err)
	} else {
		fmt.Println("Health check passed!")
	}

	// Add more test cases as needed
}

func checkHealth() error {
	url := "http://localhost:8081/api/health" // Update as needed
	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
