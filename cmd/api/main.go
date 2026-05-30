package main

import (
	"log"

	"github.com/joho/godotenv"

	"concept-tracker/config"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("godotenv: .env file not found and could not be loaded: %v", err)
	}

	load, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	n, w, err := New(load)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := w.Stop(); err != nil {
			log.Printf("worker: error stopping worker: %v", err)
		}
	}()

	n.Start()
}
