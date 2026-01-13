package config

import (
	"fmt"
	"os"
)

// RabbitMQConfig holds RabbitMQ connection configuration
type RabbitMQConfig struct {
	URL      string
	User     string
	Password string
	Host     string
	Port     string
	VHost    string
}

// LoadRabbitMQConfig loads RabbitMQ configuration from environment variables
func LoadRabbitMQConfig() *RabbitMQConfig {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		// Build URL from components
		user := getEnv("RABBITMQ_USER", "woragis")
		password := getEnv("RABBITMQ_PASSWORD", "woragis")
		host := getEnv("RABBITMQ_HOST", "rabbitmq")
		port := getEnv("RABBITMQ_PORT", "5672")
		vhost := getEnv("RABBITMQ_VHOST", "/")
		
		// URL encode the vhost for the connection string
		// For RabbitMQ, "/" becomes "%2F" in the URL, but other vhosts keep their slashes
		if vhost == "/" {
			vhost = ""
		} else if len(vhost) > 0 && vhost[0] == '/' {
			// Keep the slash but remove it for URL encoding (will be added back as %2F)
			vhost = vhost[1:]
		}
		
		url = fmt.Sprintf("amqp://%s:%s@%s:%s/%s", user, password, host, port, vhost)
	}

	return &RabbitMQConfig{
		URL:      url,
		User:     getEnv("RABBITMQ_USER", "woragis"),
		Password: getEnv("RABBITMQ_PASSWORD", "woragis"),
		Host:     getEnv("RABBITMQ_HOST", "rabbitmq"),
		Port:     getEnv("RABBITMQ_PORT", "5672"),
		VHost:    getEnv("RABBITMQ_VHOST", "/"),
	}
}

