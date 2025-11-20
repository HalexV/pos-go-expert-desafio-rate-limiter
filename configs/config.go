package configs

import (
	"github.com/go-chi/jwtauth"
	"github.com/spf13/viper"
)

type conf struct {
	IpMaxReqsBySec   int32  `mapstructure:"IP_MAX_REQS_BY_SEC"`
	IpBlockTimeBySec int32  `mapstructure:"IP_BLOCK_TIME_BY_SEC"`
	WebServerPort    string `mapstructure:"WEB_SERVER_PORT"`
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        string `mapstructure:"REDIS_PORT"`
	JWTSecret        string `mapstructure:"JWT_SECRET"`
	JWTExpiresIn     int    `mapstructure:"JWT_EXPIRES_IN"`
	TokenAuth        *jwtauth.JWTAuth
}

func LoadConfig(path string) (*conf, error) {

	var cfg *conf

	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	// Para assinar e gerar tokens JWT
	cfg.TokenAuth = jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)
	return cfg, err
}
