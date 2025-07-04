package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	AppPort             string        `mapstructure:"APP_PORT"`
	DBUrl               string        `mapstructure:"DB_URL"`
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
		// If the config file is not found, it's not a fatal error.
		// We can rely on environment variables.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %s", err)
			return
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Printf("Unable to decode into struct: %v", err)
	}

	return
}
