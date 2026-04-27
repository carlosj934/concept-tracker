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

	n, err := New(load)
	if err != nil {
		log.Fatal(err)
	}
	
	n.Start()

}
