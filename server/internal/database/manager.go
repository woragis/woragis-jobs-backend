package database

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Manager handles all database connections
type Manager struct {
	Postgres *gorm.DB
	Redis    *redis.Client
	RabbitMQ *RabbitMQConnection
}

// Config holds configuration for all database connections
type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
	RabbitMQ string // URL
}

// NewManager creates a new database manager with all connections
func NewManager(config Config) (*Manager, error) {
	manager := &Manager{}

	// Initialize PostgreSQL connection
	postgresDB, err := NewPostgres(config.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}
	manager.Postgres = postgresDB

	// Initialize Redis connection
	redisClient, err := NewRedis(config.Redis)
	if err != nil {
		// Close PostgreSQL connection if Redis fails
		if closeErr := ClosePostgres(postgresDB); closeErr != nil {
			log.Printf("Failed to close PostgreSQL connection: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	manager.Redis = redisClient

	// Initialize RabbitMQ connection (optional, warn if fails)
	if config.RabbitMQ != "" {
		rabbitMQConn, err := NewRabbitMQ(config.RabbitMQ)
		if err != nil {
			log.Printf("Warning: failed to initialize RabbitMQ: %v", err)
			// Don't fail completely, but log the warning
		} else {
			manager.RabbitMQ = rabbitMQConn
		}
	}

	log.Println("Database manager initialized successfully")
	return manager, nil
}

// Close closes all database connections
func (m *Manager) Close() error {
	var errs []error

	// Close PostgreSQL connection
	if m.Postgres != nil {
		if err := ClosePostgres(m.Postgres); err != nil {
			errs = append(errs, fmt.Errorf("failed to close PostgreSQL: %w", err))
		}
	}

	// Close Redis connection
	if m.Redis != nil {
		if err := CloseRedis(m.Redis); err != nil {
			errs = append(errs, fmt.Errorf("failed to close Redis: %w", err))
		}
	}

	// Close RabbitMQ connection
	if m.RabbitMQ != nil {
		if err := m.RabbitMQ.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close RabbitMQ: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errs)
	}

	log.Println("All database connections closed successfully")
	return nil
}

// HealthCheck performs health checks on all database connections
func (m *Manager) HealthCheck() error {
	var errs []error

	// Check PostgreSQL connection
	if m.Postgres != nil {
		if err := HealthCheck(m.Postgres); err != nil {
			errs = append(errs, fmt.Errorf("PostgreSQL health check failed: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("PostgreSQL connection is nil"))
	}

	// Check Redis connection
	if m.Redis != nil {
		if err := RedisHealthCheck(m.Redis); err != nil {
			errs = append(errs, fmt.Errorf("redis health check failed: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("redis connection is nil"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("database health check failed: %v", errs)
	}

	return nil
}

// GetPostgres returns the PostgreSQL database connection
func (m *Manager) GetPostgres() *gorm.DB {
	return m.Postgres
}

// GetRedis returns the Redis client connection
func (m *Manager) GetRedis() *redis.Client {
	return m.Redis
}

// GetRabbitMQ returns the RabbitMQ connection
func (m *Manager) GetRabbitMQ() *RabbitMQConnection {
	return m.RabbitMQ
}
