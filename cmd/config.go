package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config struct holding migoration configuration
type Config struct {
	DatabaseURL string `mapstructure:"database_url"`
	MigrationsDir string `mapstructure:"migration_path"`
}


// Loads configuration from specific file
func loadConfig(filename string) (*Config, error){
	// Load the .env file
	err := godotenv.Load(); if err != nil {
		return nil, err
	}

	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Replace placeholders with actual env values
	for _, key := range viper.AllKeys() {
		val := viper.GetString(key)
		
		// If val include ${}, extract the env variable and find it
		if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
			envKey := strings.Trim(val, "${}")
			if envVal, exists := os.LookupEnv(envKey); exists {
				viper.Set(envKey, envVal)
			} else {
				fmt.Printf("Warning: Environment variable %s not found\n", envKey)
				viper.Set(key, "")
			}
		} else {
			viper.Set(key, val)
		}
	}

	// Unmarshal the config into a struct
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}