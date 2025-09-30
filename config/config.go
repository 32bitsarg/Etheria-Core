package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"sslmode"`
	} `mapstructure:"database"`

	Redis struct {
		Host         string        `mapstructure:"host"`
		Port         int           `mapstructure:"port"`
		Password     string        `mapstructure:"password"`
		DB           int           `mapstructure:"db"`
		PoolSize     int           `mapstructure:"pool_size"`
		MinIdleConns int           `mapstructure:"min_idle_conns"`
		MaxRetries   int           `mapstructure:"max_retries"`
		DialTimeout  time.Duration `mapstructure:"dial_timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
	} `mapstructure:"redis"`

	JWT struct {
		SecretKey     string        `mapstructure:"secret_key"`
		TokenDuration time.Duration `mapstructure:"token_duration"`
	} `mapstructure:"jwt"`

	Server struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`

	RateLimit struct {
		IPLimit        int `mapstructure:"ip_limit"`
		IPWindow       int `mapstructure:"ip_window"`
		UserLimit      int `mapstructure:"user_limit"`
		UserWindow     int `mapstructure:"user_window"`
		EndpointLimit  int `mapstructure:"endpoint_limit"`
		EndpointWindow int `mapstructure:"endpoint_window"`
	} `mapstructure:"rate_limit"`

	TimeZone string `mapstructure:"timezone"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
