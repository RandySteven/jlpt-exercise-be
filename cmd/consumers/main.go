package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/RandySteven/go-kopi/apps"
	"github.com/RandySteven/go-kopi/configs"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("./files/env/.env")
	if err != nil {
		log.Fatalln(`failed to load .env `, err)
	}
}

func main() {
	configPath, err := configs.ParseFlags()
	if err != nil {
		log.Fatalln(err)
	}

	config, err := configs.NewConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := apps.NewApp(config)
	if err != nil {
		log.Fatalln(`Error starting app `, err)
	}

	consumerRunner := app.PrepareConsumer(ctx)
	if err := consumerRunner.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalln(`Error running consumers `, err)
	}

	log.Println("Consumers exiting")
}
