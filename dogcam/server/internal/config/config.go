package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	CamAPIKey      string `envconfig:"CAM_API_KEY"     required:"true"`
	ViewerPassword string `envconfig:"VIEWER_PASSWORD" required:"true"`
	GRPCPort       string `envconfig:"GRPC_PORT"       default:"50051"`
	HTTPPort       string `envconfig:"HTTP_PORT"       default:"8080"`
	MetricsPort    string `envconfig:"METRICS_PORT"    default:"9090"`
	FrameIntervalMs int32 `envconfig:"FRAME_INTERVAL_MS" default:"2000"`
}

func Load() (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return Config{}, err
	}
	return c, nil
}
