package connections

import (
	"vivek-ray/clients"
	"vivek-ray/conf"
)

var KafkaService *clients.KafkaService

func InitKafka() {
	KafkaService, _ = clients.NewKafkaService(&clients.KafkaConfig{
		Brokers:       conf.KafkaConfig.Brokers,
		Username:      conf.KafkaConfig.Username,
		Password:      conf.KafkaConfig.Password,
		SASLMechanism: conf.KafkaConfig.SASLMechanism,
		Version:       conf.KafkaConfig.Version,
	})
}

func CloseKafka() {
	KafkaService.Close()
}
