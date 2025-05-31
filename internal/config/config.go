package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Firebase      FirebaseConfig      `mapstructure:"firebase"`
	OpenTelemetry OpenTelemetryConfig `mapstructure:"openTelemetry"`
	GinMode       string              `mapstructure:"ginMode"` // e.g., "debug", "release"
}

// ServerConfig holds server-specific configuration.
type ServerConfig struct {
	HTTPPort string `mapstructure:"httpPort"`
	GRPCPort string `mapstructure:"grpcPort"`
}

// FirebaseConfig holds Firebase-specific configuration.
type FirebaseConfig struct {
	ProjectID             string `mapstructure:"projectId"`
	ServiceAccountKeyPath string `mapstructure:"serviceAccountKeyPath"` // Optional, for local dev
}

// OpenTelemetryConfig holds OpenTelemetry-specific configuration.
type OpenTelemetryConfig struct {
	ServiceName      string  `mapstructure:"serviceName"`
	ExporterEndpoint string  `mapstructure:"exporterEndpoint"` // For OTLP (Tempo, Loki)
	PrometheusPort   string  `mapstructure:"prometheusPort"` // Port for Prometheus scrape endpoint
	TraceHeaderName  string  `mapstructure:"traceHeaderName"`  // e.g., X-TRACE-ID
	SampleRate       float64 `mapstructure:"sampleRate"`
}

var AppConfig Config

// LoadConfig loads configuration from file and environment variables.
func LoadConfig(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath) // Use specific config file path
	} else {
		viper.AddConfigPath("./configs") // Path to look for the config file in
		viper.AddConfigPath(".")         // Look for config in current directory
		viper.SetConfigName("config")    // Name of config file (without extension)
		viper.SetConfigType("yaml")      // REQUIRED if the config file does not have the extension in the name
	}

	viper.AutomaticEnv() // Read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // Replace . with _ in env var names

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Error reading config file: %s. Using defaults or env vars if available.", err)
		// It might not be fatal if all configs can be supplied by Env or have defaults
	}

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
		return nil, err
	}

	// You can set default values here if needed, after trying to load from file/env
	// For example:
	if AppConfig.Server.HTTPPort == "" {
		AppConfig.Server.HTTPPort = "8080"
	}
	if AppConfig.Server.GRPCPort == "" {
		AppConfig.Server.GRPCPort = "50051"
	}
	if AppConfig.OpenTelemetry.ServiceName == "" {
		AppConfig.OpenTelemetry.ServiceName = "profile-service"
	}
	if AppConfig.OpenTelemetry.TraceHeaderName == "" {
		AppConfig.OpenTelemetry.TraceHeaderName = "X-Trace-Id" // Default to X-Trace-Id
	}
    if AppConfig.OpenTelemetry.SampleRate == 0 { // Check for 0 as it's a float64
         AppConfig.OpenTelemetry.SampleRate = 1.0 // Default to sample all traces
    }
	if AppConfig.GinMode == "" {
		AppConfig.GinMode = "debug"
	}

	log.Printf("Configuration loaded: %+v", AppConfig)
	return &AppConfig, nil
}
