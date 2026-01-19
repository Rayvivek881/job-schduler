package main

import (
	"vivek-ray/cmd"
	"vivek-ray/conf"
	"vivek-ray/connections"
)

func main() {
	viper := &conf.Viper{}
	viper.Init()

	connections.InitDB()
	connections.InitKafka()
	connections.InitSearchEngine()
	connections.InitS3()

	defer func() {
		connections.CloseSearchEngine()
		connections.CloseKafka()
		connections.CloseDB()
		connections.CloseS3()
	}()

	cmd.Execute()
}
