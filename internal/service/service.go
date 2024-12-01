package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type BonsaiService struct {
	redisClient *redis.Client
	usageChan   chan string
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

// Regular expression for key validation
var validKey = regexp.MustCompile(`^[a-zA-Z0-9._-]{2,32}$`)

// NewService creates a new Service instance with the given Redis client.
func NewService(redisClient *redis.Client) *BonsaiService {
	s := &BonsaiService{
		redisClient: redisClient,
		usageChan:   make(chan string, 100), // Buffered channel for usage updates
		stopChan:    make(chan struct{}),
	}
	go s.processUsageUpdates()
	return s
}

// processUsageUpdates processes usage updates asynchronously.
func (s *BonsaiService) processUsageUpdates() {
	ctx := context.Background()
	for {
		select {
		case key := <-s.usageChan:
			_, err := s.redisClient.Incr(ctx, fmt.Sprintf("s_usage:%s", key)).Result()
			if err != nil {
				log.Printf("Failed to update usage for key %s: %v", key, err)
			}
		case <-s.stopChan:
			// Drain remaining keys when stop signal is received
			for key := range s.usageChan {
				_, err := s.redisClient.Incr(ctx, fmt.Sprintf("s_usage:%s", key)).Result()
				if err != nil {
					log.Printf("Failed to update usage for key %s: %v", key, err)
				}
			}
			s.wg.Done() // Signal that processing is complete
			return
		}
	}
}

// Stop gracefully shuts down the service, ensuring all updates are processed.
func (s *BonsaiService) Stop() {
	close(s.stopChan)  // Signal the goroutine to stop
	s.wg.Add(1)        // Increment the wait group counter
	close(s.usageChan) // Close the usage channel to prevent further sends
	s.wg.Wait()        // Wait for the goroutine to finish processing
}

// NewShorteningHandler initializes the shortening service handler.
func (s *BonsaiService) NewShorteningHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")
		ctx := context.Background()

		// Check if key is valid
		if !validKey.MatchString(key) {
			return c.Status(http.StatusBadRequest).SendString("Invalid key format")
		}

		// Fetch the original URL from Redis
		originalURL, err := s.redisClient.Get(ctx, fmt.Sprintf("s:%s", key)).Result()
		if err == redis.Nil {
			return c.Status(http.StatusNotFound).SendString("Shortened URL not found")
		} else if err != nil {
			log.Printf("Error fetching URL: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Send the key to the usage update channel
		select {
		case s.usageChan <- key:
		default:
			log.Printf("Usage update channel full, dropping key: %s", key)
		}

		// Redirect to the original URL
		return c.Redirect(originalURL, http.StatusFound)
	}
}

// NewAPIHandler initializes the API service handler.
func (s *BonsaiService) NewAPIHandler() *fiber.App {
	router := fiber.New()

	router.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// Add URL
	router.Post("/url", func(c *fiber.Ctx) error {
		type AddURLRequest struct {
			Key         string `json:"key"`
			OriginalURL string `json:"original_url"`
		}

		var req AddURLRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(http.StatusBadRequest).SendString("Invalid request")
		}

		if req.Key == "" {
			return c.Status(http.StatusBadRequest).SendString("Empty request")
		}

		// Check if key is valid
		if !validKey.MatchString(req.Key) {
			return c.Status(http.StatusBadRequest).SendString("Invalid key format")
		}

		// Store in Redis
		ctx := context.Background()
		err := s.redisClient.Set(ctx, fmt.Sprintf("s:%s", req.Key), req.OriginalURL, 0).Err()
		if err != nil {
			log.Printf("Error saving URL: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}

		return c.SendString("Successfully added URL")
	})

	// Get URL
	router.Get("/url/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		ctx := context.Background()

		// Check if key is valid
		if !validKey.MatchString(key) {
			return c.Status(http.StatusBadRequest).SendString("Invalid key format")
		}

		originalURL, err := s.redisClient.Get(ctx, fmt.Sprintf("s:%s", key)).Result()
		if err == redis.Nil {
			return c.Status(http.StatusNotFound).SendString("URL not found")
		} else if err != nil {
			log.Printf("Error fetching URL: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}

		return c.JSON(fiber.Map{"key": key, "original_url": originalURL})
	})

	// List all URLs
	router.Get("/urls", func(c *fiber.Ctx) error {
		ctx := context.Background()
		keys, err := s.redisClient.Keys(ctx, "s:*").Result()
		if err != nil {
			log.Printf("Error listing keys: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}

		urls := make(map[string]string)
		for _, key := range keys {
			originalURL, err := s.redisClient.Get(ctx, key).Result()
			if err == nil {
				shortKey := key[2:] // Remove the "s:" prefix
				urls[shortKey] = originalURL
			}
		}

		return c.JSON(urls)
	})

	// Delete URL
	router.Delete("/url/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")
		ctx := context.Background()

		// Check if key is valid
		if !validKey.MatchString(key) {
			return c.Status(http.StatusBadRequest).SendString("Invalid key format")
		}

		// Delete the shortened URL
		_, err := s.redisClient.Del(ctx, fmt.Sprintf("s:%s", key)).Result()
		if err != nil {
			log.Printf("Error deleting key: %v", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}

		return c.SendString("Successfully deleted URL")
	})

	return router
}
