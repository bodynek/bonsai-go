package main

import (
	"bonsai-go/internal/config"
	"bonsai-go/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var (
	noAPI      bool
	apiPort    int
	svcPort    int
	redisHost  string
	redisPort  int
	configFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bonsaid",
		Short: "Bonsai shortening daemon",
		Run:   runDaemon,
	}

	rootCmd.Flags().BoolVar(&noAPI, "no_api", false, "Disable API service")
	rootCmd.Flags().IntVar(&apiPort, "api_port", 0, "Port for the API service")
	rootCmd.Flags().IntVar(&svcPort, "svc_port", 0, "Port for the shortening service")
	rootCmd.Flags().StringVar(&redisHost, "redis_host", "", "Host for the Redis database")
	rootCmd.Flags().IntVar(&redisPort, "redis_port", 0, "Port for the Redis database")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to the configuration file")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error starting bonsaid: %v", err)
	}
}

func runDaemon(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Override with command-line parameters if provided
	if cmd.Flag("no_api").Changed {
		cfg.NoAPI = noAPI
	}
	if cmd.Flag("api_port").Changed {
		cfg.APIPort = apiPort
	}
	if cmd.Flag("svc_port").Changed {
		cfg.SvcPort = svcPort
	}
	if cmd.Flag("redis_host").Changed {
		cfg.RedisHost = redisHost
	}
	if cmd.Flag("redis_port").Changed {
		cfg.RedisPort = redisPort
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})

	// Initialize service
	bonsaiService := service.NewService(redisClient)

	if !cfg.NoAPI {
		// Create Fiber app for API
		apiApp := fiber.New()
		apiApp.Mount("/api", bonsaiService.NewAPIHandler())
		go func() {
			if err := apiApp.Listen(fmt.Sprintf(":%d", cfg.APIPort)); err != nil {
				log.Fatalf("Failed to start API: %v", err)
			}
		}()
	}

	// Create Fiber app for shortening service
	shortApp := fiber.New()
	shortApp.Get("/:key", bonsaiService.NewShorteningHandler())

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := shortApp.Listen(fmt.Sprintf(":%d", cfg.SvcPort)); err != nil {
			log.Fatalf("Failed to start shortening service: %v", err)
		}
	}()

	<-stop // Wait for interrupt signal

	log.Println("Shutting down shortening service...")
	bonsaiService.Stop() // Stop the service gracefully
	if err := shortApp.Shutdown(); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}
	log.Println("Service stopped.")
}
