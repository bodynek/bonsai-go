package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	apiHost string
	apiPort string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bonsai-cli",
		Short: "CLI for Bonsai shortening service",
	}

	rootCmd.PersistentFlags().StringVar(&apiHost, "api_host", "localhost", "Host of the Bonsai API service")
	rootCmd.PersistentFlags().StringVar(&apiPort, "api_port", "8081", "Port of the Bonsai API service")

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "add [key] [url]",
			Short: "Add a shortened URL",
			Args:  cobra.ExactArgs(2),
			Run:   addShortenedURL,
		},
		&cobra.Command{
			Use:   "get [key]",
			Short: "Get the original URL by shortened key",
			Args:  cobra.ExactArgs(1),
			Run:   getShortenedURL,
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all shortened URLs",
			Args:  cobra.NoArgs,
			Run:   listShortenedURLs,
		},
		&cobra.Command{
			Use:   "delete [key]",
			Short: "Delete a shortened URL",
			Args:  cobra.ExactArgs(1),
			Run:   deleteShortenedURL,
		},
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error running bonsai-cli: %v", err)
	}
}

func addShortenedURL(cmd *cobra.Command, args []string) {
	key, url := args[0], args[1]
	data := fmt.Sprintf(`{"key": "%s", "original_url": "%s"}`, key, url)
	resp, err := http.Post(fmt.Sprintf("http://%s:%s/api/url", apiHost, apiPort), "application/json", bytes.NewBufferString(data))
	if err != nil {
		log.Fatalf("Error adding shortened URL: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Shortened URL added successfully")
}

func listShortenedURLs(cmd *cobra.Command, args []string) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/urls", apiHost, apiPort))
	if err != nil {
		log.Fatalf("Error retrieving shortened URLs: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response status is not 200 OK
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to fetch URLs, received status code: %d", resp.StatusCode)
	}

	// Parse JSON response
	var urls map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&urls); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	// Display URLs
	fmt.Println("Shortened URLs:")
	for shortKey, originalURL := range urls {
		fmt.Printf("- %s -> %s\n", shortKey, originalURL)
	}
}

func getShortenedURL(cmd *cobra.Command, args []string) {
	key := args[0]
	resp, err := http.Get(fmt.Sprintf("http://%s:%s/api/url/%s", apiHost, apiPort, key))
	if err != nil {
		log.Fatalf("Error retrieving shortened URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to retrieve URL: %s", resp.Status)
	}

	var url map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&url); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	fmt.Printf("Shortened URL:\n  %s -> %s\n", url["key"], url["original_url"])
}

func deleteShortenedURL(cmd *cobra.Command, args []string) {
	key := args[0]
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s:%s/api/url/%s", apiHost, apiPort, key), nil)
	if err != nil {
		log.Fatalf("Error creating delete request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error deleting shortened URL: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println("Shortened URL deleted successfully")
}
