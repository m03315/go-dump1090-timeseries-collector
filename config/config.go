package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all the application configuration settings.
type Config struct {
	Dump1090Host      string
	Dump1090Port      string
	InfluxHost        string
	InfluxToken       string
	InfluxDatabase    string
	BatchSize         int
	BatchInterval     time.Duration
	ConnectRetryDelay time.Duration
	ConnectMaxRetries int
	OutputDBType      string // New field to select the output database type
}

const (
	defaultBatchSize     = 50
	defaultBatchInterval = 5 * time.Second
	defaultRetryDelay    = 5 * time.Second
	defaultMaxRetries    = 0 // 0 means infinite retries
)

// LoadConfig loads configuration from environment variables and provides defaults.
// In a more complex application, this might also read from a config file (e.g., YAML, JSON).
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Dump1090Host:   getEnv("DUMP1090_HOST", "localhost"),
		Dump1090Port:   getEnv("DUMP1090_PORT", "30003"),
		InfluxHost:     os.Getenv("INFLUX_URL"),              // No default, mandatory for InfluxDB type
		InfluxToken:    os.Getenv("INFLUXDB_TOKEN"),          // No default
		InfluxDatabase: os.Getenv("INFLUXDB_DATABASE"),       // No default
		OutputDBType:   getEnv("OUTPUT_DB_TYPE", "influxdb"), // Default to influxdb

		BatchSize:         getEnvAsInt("BATCH_SIZE", defaultBatchSize),
		BatchInterval:     getEnvAsDuration("BATCH_INTERVAL", defaultBatchInterval),
		ConnectRetryDelay: getEnvAsDuration("CONNECT_RETRY_DELAY", defaultRetryDelay),
		ConnectMaxRetries: getEnvAsInt("CONNECT_MAX_RETRIES", defaultMaxRetries),
	}

	// You could add validation logic here
	if cfg.OutputDBType == "influxdb" {
		if cfg.InfluxHost == "" || cfg.InfluxToken == "" || cfg.InfluxDatabase == "" {
			return nil, fmt.Errorf("INFLUX_URL, INFLUXDB_TOKEN, and INFLUXDB_DATABASE must be set for InfluxDB output type")
		}
	} else {
		// Add validation for other DB types here if they have mandatory fields
		return nil, fmt.Errorf("unsupported OUTPUT_DB_TYPE: %s", cfg.OutputDBType)
	}

	return cfg, nil
}

// Helper function to get environment variable or use a default.
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// Helper function to get environment variable as an integer.
func getEnvAsInt(key string, defaultVal int) int {
	strVal := getEnv(key, "")
	if strVal == "" {
		return defaultVal
	}
	if intVal, err := strconv.Atoi(strVal); err == nil {
		return intVal
	}
	return defaultVal // Fallback if parsing fails
}

// Helper function to get environment variable as a time.Duration.
func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	strVal := getEnv(key, "")
	if strVal == "" {
		return defaultVal
	}
	if durVal, err := time.ParseDuration(strVal); err == nil {
		return durVal
	}
	return defaultVal // Fallback if parsing fails
}
