package config

import (
	"os"
)

type Config struct {
	HTTPAddr       string
	DSN            string
	MigrationsPath string
	RabbitMQURL    string
	APIKey         string
	APIUrl         string
}

func Read() Config {
	var config Config

	httpAddr, exists := os.LookupEnv("HTTP_ADDR")
	if exists {
		config.HTTPAddr = httpAddr
	} else {
		config.HTTPAddr = ":8080" // Default to localhost on port 8080
	}

	dsn, exists := os.LookupEnv("DSN")
	if exists {
		config.DSN = dsn
	} else {
		config.DSN = "" // Default PostgreSQL DSN with localhost
	}

	migrationsPath, exists := os.LookupEnv("MIGRATIONS_PATH")
	if exists {
		config.MigrationsPath = migrationsPath
	} else {
		config.MigrationsPath = "file://internal/app/migrations" // Default migrations path
	}

	rabbitMQURL, exists := os.LookupEnv("RABBITMQ_URL")
	if exists {
		config.RabbitMQURL = rabbitMQURL
	} else {
		config.RabbitMQURL = "amqp://guest:guest@localhost:5673/" // Default RabbitMQ URL with localhost
	}

	apiKey, exists := os.LookupEnv("API_KEY")
	if exists {
		config.APIKey = apiKey
	} else {
		config.APIKey = "" // Replace with your default API key
	}

	apiUrl, exists := os.LookupEnv("API_URL")
	if exists {
		config.APIUrl = apiUrl
	} else {
		config.APIUrl = "https://api.nasa.gov/mars-photos/api/v1/rovers/curiosity/photos" // Default NASA API URL
	}

	return config
}
