package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/victorximenis/multitenant"
	"github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment
	config, err := multitenant.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create multitenant client
	client, err := multitenant.NewMultitenantClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create multitenant client: %v", err)
	}
	defer client.Close(ctx)

	// Setup Gin router
	router := gin.Default()

	// Add tenant middleware
	router.Use(client.GinMiddleware())

	// Add routes
	router.GET("/api/tenant", func(c *gin.Context) {
		tenant, ok := tenantcontext.GetTenant(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not found in context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tenant": tenant,
		})
	})

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown gracefully
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server shutdown complete")
}
