package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func baseConfig() Config {
	cfg := Config{}
	cfg.Common.ECR.ImagePrefix = "123456789.dkr.ecr.eu-west-1.amazonaws.com/apps"
	cfg.Common.ECR.Region = "eu-west-1"
	cfg.Common.ECR.PushIAMRoleARN = "arn:aws:iam::123456789:role/gh-ecr"
	return cfg
}

func TestBuildMatrix_SingleApp(t *testing.T) {
	cfg := baseConfig()
	cfg.Apps = []App{
		{
			Name: "myapp",
			Path: "myapp/",
			Tests: []Test{
				{Name: "go", Run: "go test ./..."},
			},
		},
	}

	m := buildMatrix(cfg)

	if len(m.Include) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Include))
	}
	e := m.Include[0]
	if e.Name != "myapp" {
		t.Errorf("name: got %q, want %q", e.Name, "myapp")
	}
	if e.Dockerfile != "myapp/Dockerfile" {
		t.Errorf("dockerfile: got %q, want %q", e.Dockerfile, "myapp/Dockerfile")
	}
	if e.Image != "123456789.dkr.ecr.eu-west-1.amazonaws.com/apps/myapp" {
		t.Errorf("image: got %q, want %q", e.Image, "123456789.dkr.ecr.eu-west-1.amazonaws.com/apps/myapp")
	}
	if e.ECRRegion != "eu-west-1" {
		t.Errorf("ecr_region: got %q, want %q", e.ECRRegion, "eu-west-1")
	}
	if e.ECRRoleARN != "arn:aws:iam::123456789:role/gh-ecr" {
		t.Errorf("ecr_role_arn: got %q, want %q", e.ECRRoleARN, "arn:aws:iam::123456789:role/gh-ecr")
	}
	if len(e.Tests) != 1 {
		t.Fatalf("tests length: got %d, want 1", len(e.Tests))
	}
	if e.Tests[0].Name != "go" {
		t.Errorf("tests[0].name: got %q, want %q", e.Tests[0].Name, "go")
	}
	if e.Tests[0].Run != "go test ./..." {
		t.Errorf("tests[0].run: got %q, want %q", e.Tests[0].Run, "go test ./...")
	}
}

func TestBuildMatrix_MultipleApps(t *testing.T) {
	cfg := baseConfig()
	cfg.Apps = []App{
		{Name: "alpha", Path: "alpha/", Tests: []Test{{Name: "go", Run: "go test ./alpha/..."}}},
		{Name: "beta", Path: "services/beta/", Tests: []Test{{Name: "go", Run: "go test ./services/beta/..."}}},
	}

	m := buildMatrix(cfg)

	if len(m.Include) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Include))
	}
	if m.Include[0].Dockerfile != "alpha/Dockerfile" {
		t.Errorf("alpha dockerfile: got %q, want %q", m.Include[0].Dockerfile, "alpha/Dockerfile")
	}
	if m.Include[1].Dockerfile != "services/beta/Dockerfile" {
		t.Errorf("beta dockerfile: got %q, want %q", m.Include[1].Dockerfile, "services/beta/Dockerfile")
	}
	if m.Include[1].Image != "123456789.dkr.ecr.eu-west-1.amazonaws.com/apps/beta" {
		t.Errorf("beta image: got %q, want %q", m.Include[1].Image, "123456789.dkr.ecr.eu-west-1.amazonaws.com/apps/beta")
	}
}

func TestBuildMatrix_ClusterField(t *testing.T) {
	cfg := baseConfig()
	cfg.Apps = []App{
		{
			Name:       "myapp",
			Path:       "myapp/",
			Deployment: Deployment{Cluster: "mgmt"},
			Tests:      []Test{{Name: "go", Run: "go test ./..."}},
		},
	}

	m := buildMatrix(cfg)

	require.Len(t, m.Include, 1)
	require.Equal(t, "mgmt", m.Include[0].Cluster)
}

func TestBuildMatrix_NoDeployment(t *testing.T) {
	cfg := baseConfig()
	cfg.Apps = []App{
		{
			Name:  "myapp",
			Path:  "myapp/",
			Tests: []Test{{Name: "go", Run: "go test ./..."}},
		},
	}

	m := buildMatrix(cfg)

	require.Len(t, m.Include, 1)
	require.Equal(t, "", m.Include[0].Cluster)
}
