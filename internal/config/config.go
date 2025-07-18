package config

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	AppEnv              string        `mapstructure:"APP_ENV" validate:"oneof=development production"`
	AppPort             string        `mapstructure:"APP_PORT"`
	DBUrl               string        `mapstructure:"DB_URL"`
	TestDBUrl           string        `mapstructure:"TEST_DB_URL" validate:"required"`
	JWTAccessSecret     string        `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret    string        `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessExpiresIn  time.Duration `mapstructure:"JWT_ACCESS_EXPIRES_IN"`
	JWTRefreshExpiresIn time.Duration `mapstructure:"JWT_REFRESH_EXPIRES_IN"`
	MinioEndpoint       string        `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey      string        `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey      string        `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucketName     string        `mapstructure:"MINIO_BUCKET_NAME"`
	MinioUseSSL         bool          `mapstructure:"MINIO_USE_SSL"`
	RedisAddr           string        `mapstructure:"REDIS_ADDR"`
	CorsAllowedOrigins  []string      `mapstructure:"CORS_ALLOWED_ORIGINS"`
	RabbitMQUrl         string        `mapstructure:"RABBITMQ_URL" validate:"required"`
	SmtpHost            string        `mapstructure:"SMTP_HOST" validate:"required"`
	SmtpPort            int           `mapstructure:"SMTP_PORT" validate:"required"`
	SmtpUsername        string        `mapstructure:"SMTP_USERNAME"`
	SmtpPassword        string        `mapstructure:"SMTP_PASSWORD"`
	SmtpSender          string        `mapstructure:"SMTP_SENDER" validate:"required,email"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (config Config, err error) {
	// Set the file name of the configurations file
	viper.SetConfigName(".env")
	// Set the type of the configuration file
	viper.SetConfigType("env")
	// Add the path to look for the configurations file
	viper.AddConfigPath(".")
	// Add the path to look for the configurations file inside the container
	viper.AddConfigPath("/app")

	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %s", err)
			return
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("Unable to decode into struct: %v", err)
		return
	}

	validate := validator.New()
	if err = validate.Struct(&config); err != nil {
		log.Printf("Missing required configuration: %v", err)
		return
	}

	return
}
