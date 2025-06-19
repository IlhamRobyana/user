package main

//go:generate go run github.com/swaggo/swag/cmd/swag init
//go:generate go run github.com/google/wire/cmd/wire

import (
	"github.com/IlhamRobyana/user/configs"
	"github.com/IlhamRobyana/user/shared/logger"
)

var configServiceGen *configs.Config

// @securityDefinitions.apikey EVMOauthToken
// @in header
// @name Authorization
func main() {
	// Initialize logger
	logger.InitLogger()

	// Initialize config
	configServiceGen = configs.Get()

	// Set desired log level
	logger.SetLogLevel(configServiceGen)

	// Wire everything up
	httpServiceGen := InitializeServiceServiceGen()

	// Run server
	httpServiceGen.SetupAndServe()
}
