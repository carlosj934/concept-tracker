package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	// DB
	DBHost     string
	DBPort     int64
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode string

	// API
	ServerPort int64

	// Clerk
	ClerkSecretKey string
}

func Load() (*Config, error) {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return nil, errors.New("DB_HOST is required")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		return nil, errors.New("DB_PORT is required")
	}
	dbPortInt, err := strconv.ParseInt(dbPort, 10, 64)
	if err != nil {
		return nil, errors.New("DB_PORT must be a valid integer")
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return nil, errors.New("DB_USER is required")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return nil, errors.New("DB_PASSWORD is required")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return nil, errors.New("DB_NAME is required")
	}

	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	serverPortInt, err := strconv.ParseInt(serverPort, 10, 64)
	if err != nil {
		return nil, errors.New("SERVERPORT must be a valid integer")
	}

	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		return nil, errors.New("CLERK_SECRET_KEY is required")
	}

	c := &Config{
		DBHost:         dbHost,
		DBPort:         dbPortInt,
		DBUser:         dbUser,
		DBPassword:     dbPassword,
		DBName:         dbName,
		SSLMode: sslMode,
		ServerPort:     serverPortInt,
		ClerkSecretKey: clerkSecretKey,
	}

	return c, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.SSLMode)
}
