package utils

import (
	"github.com/joho/godotenv"
	"log"
)

func EnvLoad(path string) {
	if path == "" {
		path = ".env"
	}
	err := godotenv.Load(path)
	if err != nil {
		log.Fatalf("Error loading env target")
	}
}
