package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vivek-ray/connections"
	"vivek-ray/middleware"
	"vivek-ray/server"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "api-server",
	Short: "Start the API server",
	Long:  `Start the API server for the Connectra API`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func startServer() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(gzip.Gzip(gzip.DefaultCompression))

	router.Use(middleware.RateLimiter())
	router.Use(middleware.APIKeyAuth())

	router.SetTrustedProxies(nil)

	router.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status": "ok",
			"timestamp": gin.H{
				"current": time.Now().Format(time.RFC3339),
			},
			"components": gin.H{},
		}

		// Check database connectivity
		if connections.PgDBConnection != nil && connections.PgDBConnection.Client != nil {
			if err := connections.PgDBConnection.Client.Ping(); err != nil {
				health["components"].(gin.H)["database"] = gin.H{"status": "down", "error": err.Error()}
				health["status"] = "degraded"
			} else {
				health["components"].(gin.H)["database"] = gin.H{"status": "up"}
			}
		} else {
			health["components"].(gin.H)["database"] = gin.H{"status": "not_initialized"}
			health["status"] = "degraded"
		}

		// Check Kafka connectivity
		if connections.KafkaService != nil {
			client := connections.KafkaService.Client()
			if client == nil {
				health["components"].(gin.H)["kafka"] = gin.H{"status": "not_initialized"}
				health["status"] = "degraded"
			} else {
				brokers := client.Brokers()
				if len(brokers) == 0 {
					health["components"].(gin.H)["kafka"] = gin.H{"status": "down", "error": "no brokers available"}
					health["status"] = "degraded"
				} else {
					health["components"].(gin.H)["kafka"] = gin.H{"status": "up", "brokers": len(brokers)}
				}
			}
		} else {
			health["components"].(gin.H)["kafka"] = gin.H{"status": "not_initialized"}
			health["status"] = "degraded"
		}

		statusCode := 200
		if health["status"] == "degraded" {
			statusCode = 503
		}

		c.JSON(statusCode, health)
	})

	server.Routes(router.Group("/jobs"))

	srv := &http.Server{Addr: ":8000", Handler: router}

	go func() {
		log.Info().Msg("Starting server on :8000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Error starting server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}
}
