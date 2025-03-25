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
	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}


	// Fetch env file path 
	envPath := viper.GetString("env_file")
	if envPath == "" {
		envPath = ".env"
	}

	// Check if the .env file exists at the given path
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".env file not found at '%s'", envPath)
	}

	// Load the env file
	err := godotenv.Load(envPath); if err != nil {
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