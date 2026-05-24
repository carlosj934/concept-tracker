package main

import (
	"log"

	"concept-tracker/config"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

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
