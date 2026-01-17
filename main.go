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

	defer func() {
		connections.CloseKafka()
		connections.CloseDB()
	}()

	cmd.Execute()
}
