package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"vivek-ray/jobs"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var ConsumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Start the Kafka jobs consumer",
	Long:  "Start the Kafka jobs consumer to process jobs from the queue",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		consumer := jobs.NewBaseConsumer()
		if consumer == nil {
			log.Fatal().Msg("Failed to initialize consumer")
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Info().Msg("Starting consumer...")
			if err := consumer.ConsumeJobs(ctx); err != nil {
				log.Error().Err(err).Msg("Consumer error")
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info().Msg("Shutting down consumer...")
		cancel()
		wg.Wait()
		log.Info().Msg("Consumer stopped")
	},
}

func init() {
	rootCmd.AddCommand(ConsumerCmd)
}
