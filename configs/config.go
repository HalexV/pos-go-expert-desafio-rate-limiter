package configs

import (
	"fmt"

	"github.com/go-chi/jwtauth"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type conf struct {
	IpMaxReqsBySec   int32  `mapstructure:"IP_MAX_REQS_BY_SEC" validate:"required"`
	IpBlockTimeBySec int32  `mapstructure:"IP_BLOCK_TIME_BY_SEC" validate:"required"`
	WebServerPort    string `mapstructure:"WEB_SERVER_PORT" validate:"required"`
	RedisHost        string `mapstructure:"REDIS_HOST" validate:"required"`
	RedisPort        string `mapstructure:"REDIS_PORT" validate:"required"`
	JWTSecret        string `mapstructure:"JWT_SECRET" validate:"required"`
	JWTExpiresIn     int    `mapstructure:"JWT_EXPIRES_IN" validate:"required"`
	TokenAuth        *jwtauth.JWTAuth
}

func LoadConfig(path string) (*conf, error) {
	var cfg conf

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// ENV
	viper.AutomaticEnv()

	// bind ENV VARS explicitamente
	keys := []string{
		"IP_MAX_REQS_BY_SEC",
		"IP_BLOCK_TIME_BY_SEC",
		"WEB_SERVER_PORT",
		"REDIS_HOST",
		"REDIS_PORT",
		"JWT_SECRET",
		"JWT_EXPIRES_IN",
	}
	for _, key := range keys {
		viper.BindEnv(key)
	}

	// Try load .env
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println(".env não encontrado — usando apenas variáveis do ambiente")
		} else {
			return nil, err
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&cfg); err != nil {
		return nil, err
	}

	cfg.TokenAuth = jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	return &cfg, nil
}
