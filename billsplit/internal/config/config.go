package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	S3Bucket     string `envconfig:"S3_BUCKET"      required:"true"`
	S3Endpoint   string `envconfig:"S3_ENDPOINT"`
	JWTSecret    string `envconfig:"JWT_SECRET"     required:"true"`
	Port         string `envconfig:"PORT"           default:"8080"`
	SecureCookie bool   `envconfig:"SECURE_COOKIE"  default:"true"`
}

func Load() (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return Config{}, err
	}
	return c, nil
}
