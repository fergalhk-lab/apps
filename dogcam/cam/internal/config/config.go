package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ServerAddr   string `envconfig:"SERVER_ADDR"   required:"true"`
	CameraDevice string `envconfig:"CAMERA_DEVICE" default:"/dev/video0"`
	CamAPIKey    string `envconfig:"CAM_API_KEY"   required:"true"`
}

func Load() (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return Config{}, err
	}
	return c, nil
}
