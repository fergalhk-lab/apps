package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Common struct {
		ECR struct {
			PushIAMRoleARN string `yaml:"pushIAMRoleARN"`
			Region         string `yaml:"region"`
			ImagePrefix    string `yaml:"imagePrefix"`
		} `yaml:"ecr"`
	} `yaml:"common"`
	Apps []App `yaml:"apps"`
}

type App struct {
	Name          string   `yaml:"name"`
	Path          string   `yaml:"path"`
	Image         string   `yaml:"image"`
	SetupCommands []string `yaml:"setupCommands"`
	Tests         []Test   `yaml:"tests"`
}

type Test struct {
	Name string `yaml:"name" json:"name"`
	Run  string `yaml:"run"  json:"run"`
}

type MatrixEntry struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	Dockerfile    string   `json:"dockerfile"`
	Image         string   `json:"image"`
	SetupCommands []string `json:"setup_commands"`
	Tests         []Test   `json:"tests"`
	ECRRegion     string   `json:"ecr_region"`
	ECRRoleARN    string   `json:"ecr_role_arn"`
}

type Matrix struct {
	Include []MatrixEntry `json:"include"`
}

func buildMatrix(cfg Config) Matrix {
	entries := make([]MatrixEntry, 0, len(cfg.Apps))
	for _, app := range cfg.Apps {
		image := cfg.Common.ECR.ImagePrefix + "/" + app.Name
		if app.Image != "" {
			image = app.Image
		}
		entries = append(entries, MatrixEntry{
			Name:          app.Name,
			Path:          app.Path,
			Dockerfile:    app.Path + "Dockerfile",
			Image:         image,
			SetupCommands: app.SetupCommands,
			Tests:         app.Tests,
			ECRRegion:     cfg.Common.ECR.Region,
			ECRRoleARN:    cfg.Common.ECR.PushIAMRoleARN,
		})
	}
	return Matrix{Include: entries}
}

func main() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config.yaml: %v\n", err)
		os.Exit(1)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config.yaml: %v\n", err)
		os.Exit(1)
	}

	matrix := buildMatrix(cfg)
	out, err := json.Marshal(matrix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding matrix: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
