package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Raft        RaftConfig
	Telemetry   TelemetryConfig
	Security    SecurityConfig
	RateLimit   RateLimitConfig
	Environment string
	LogLevel    string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	MaxRetries int
	PoolSize   int
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret              string
	Expiration          time.Duration
	RefreshExpiration   time.Duration
	Issuer              string
}

// RaftConfig holds Raft consensus configuration
type RaftConfig struct {
	NodeID             string
	BindAddr           string
	AdvertiseAddr      string
	DataDir            string
	SnapshotInterval   time.Duration
	SnapshotThreshold  uint64
	HeartbeatTimeout   time.Duration
	ElectionTimeout    time.Duration
	Bootstrap          bool
	JoinAddresses      []string
}

// TelemetryConfig holds OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled             bool
	ServiceName         string
	ServiceVersion      string
	OTLPEndpoint        string
	OTLPTracesEndpoint  string
	OTLPMetricsEndpoint string
	TraceRatio          float64
	MetricsInterval     time.Duration
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BCryptCost      int
	APIKeyLength    int
	APIKeyPrefix    string
	MaxRequestSize  int64
	MaxResponseSize int64
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled               bool
	RequestsPerSecond     int
	Burst                 int
	CleanupInterval       time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "cfguardian"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		
		Redis: RedisConfig{
			Host:       getEnv("REDIS_HOST", "localhost"),
			Port:       getEnvInt("REDIS_PORT", 6379),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getEnvInt("REDIS_DB", 0),
			MaxRetries: getEnvInt("REDIS_MAX_RETRIES", 3),
			PoolSize:   getEnvInt("REDIS_POOL_SIZE", 10),
		},
		
		JWT: JWTConfig{
			Secret:            mustGetEnv("JWT_SECRET"),
			Expiration:        getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			RefreshExpiration: getEnvDuration("JWT_REFRESH_EXPIRATION", 168*time.Hour),
			Issuer:            getEnv("JWT_ISSUER", "cfguardian"),
		},
		
		Raft: RaftConfig{
			NodeID:            getEnv("RAFT_NODE_ID", "node1"),
			BindAddr:          getEnv("RAFT_BIND_ADDR", "127.0.0.1:7000"),
			AdvertiseAddr:     getEnv("RAFT_ADVERTISE_ADDR", "127.0.0.1:7000"),
			DataDir:           getEnv("RAFT_DATA_DIR", "./raft-data"),
			SnapshotInterval:  getEnvDuration("RAFT_SNAPSHOT_INTERVAL", 30*time.Second),
			SnapshotThreshold: uint64(getEnvInt("RAFT_SNAPSHOT_THRESHOLD", 1024)),
			HeartbeatTimeout:  getEnvDuration("RAFT_HEARTBEAT_TIMEOUT", 1*time.Second),
			ElectionTimeout:   getEnvDuration("RAFT_ELECTION_TIMEOUT", 1*time.Second),
			Bootstrap:         getEnvBool("RAFT_BOOTSTRAP", true),
			JoinAddresses:     getEnvSlice("RAFT_JOIN_ADDRESSES", []string{}),
		},
		
		Telemetry: TelemetryConfig{
			Enabled:             getEnvBool("OTEL_ENABLED", true),
			ServiceName:         getEnv("OTEL_SERVICE_NAME", "cfguardian"),
			ServiceVersion:      getEnv("OTEL_SERVICE_VERSION", "1.0.0"),
			OTLPEndpoint:        getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
			OTLPTracesEndpoint:  getEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "http://localhost:4318/v1/traces"),
			OTLPMetricsEndpoint: getEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "http://localhost:4318/v1/metrics"),
			TraceRatio:          getEnvFloat("OTEL_TRACE_RATIO", 1.0),
			MetricsInterval:     getEnvDuration("OTEL_METRICS_INTERVAL", 15*time.Second),
		},
		
		Security: SecurityConfig{
			BCryptCost:      getEnvInt("BCRYPT_COST", 10),
			APIKeyLength:    getEnvInt("API_KEY_LENGTH", 32),
			APIKeyPrefix:    getEnv("API_KEY_PREFIX", "cfg_"),
			MaxRequestSize:  int64(getEnvInt("MAX_REQUEST_SIZE", 10*1024*1024)), // 10MB
			MaxResponseSize: int64(getEnvInt("MAX_RESPONSE_SIZE", 10*1024*1024)),
		},
		
		RateLimit: RateLimitConfig{
			Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: getEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 100),
			Burst:             getEnvInt("RATE_LIMIT_BURST", 200),
			CleanupInterval:   getEnvDuration("RATE_LIMIT_CLEANUP_INTERVAL", 1*time.Minute),
		},
	}
	
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	
	if c.Security.BCryptCost < 4 || c.Security.BCryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}
	
	return nil
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	// Simple comma-separated parsing
	var result []string
	for _, v := range splitAndTrim(valueStr, ",") {
		if v != "" {
			result = append(result, v)
		}
	}
	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	// Simple split implementation
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}

