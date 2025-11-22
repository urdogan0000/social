package config

import (
	"github.com/urdogan0000/social/internal/env"
)

type Config struct {
	Server   ServerConfig
	DB       DBConfig
	JWT      JWTConfig
	EventBus EventBusConfig
}

type ServerConfig struct {
	Addr            string
	RateLimitRPM    int
	EnableCORS      bool
	AllowedOrigins  []string
	IsDevelopment   bool
}

type DBConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type JWTConfig struct {
	SecretKey     string
	ExpirationHours int
}

type EventBusConfig struct {
	Type string
	Kafka KafkaConfig
	NATS  NATSConfig
}

type KafkaConfig struct {
	Brokers     []string
	TopicPrefix string
}

type NATSConfig struct {
	URL string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:            env.GetString("ADDR", ":8080"),
			RateLimitRPM:    env.GetInt("RATE_LIMIT_RPM", 100),
			EnableCORS:      env.GetBool("ENABLE_CORS", true),
			AllowedOrigins:  env.GetStringSlice("ALLOWED_ORIGINS", []string{"*"}),
			IsDevelopment:   env.GetBool("IS_DEVELOPMENT", false),
		},
		DB: DBConfig{
			Addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		JWT: JWTConfig{
			SecretKey:       env.GetString("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpirationHours: env.GetInt("JWT_EXPIRATION_HOURS", 24),
		},
		EventBus: EventBusConfig{
			Type: env.GetString("EVENT_BUS_TYPE", "inmemory"),
			Kafka: KafkaConfig{
				Brokers:     env.GetStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
				TopicPrefix: env.GetString("KAFKA_TOPIC_PREFIX", "social"),
			},
			NATS: NATSConfig{
				URL: env.GetString("NATS_URL", "nats://localhost:4222"),
			},
		},
	}
}
