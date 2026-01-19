package connections

import (
	"vivek-ray/clients"
	"vivek-ray/conf"

	"github.com/rs/zerolog/log"
)

var KafkaService *clients.KafkaService

func InitKafka() {
	kafkaService, err := clients.NewKafkaService(&clients.KafkaConfig{
		Brokers:       conf.KafkaConfig.Brokers,
		Username:      conf.KafkaConfig.Username,
		Password:      conf.KafkaConfig.Password,
		SASLMechanism: conf.KafkaConfig.SASLMechanism,
		Version:       conf.KafkaConfig.Version,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize kafka service")
	}
	KafkaService = kafkaService
}

func CloseKafka() {
	KafkaService.Close()
}
