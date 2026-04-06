package kubeconfig

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
)

// Parsed holds connection info extracted from a kubeconfig.
type Parsed struct {
	Server   string
	CAData   []byte
	CertData []byte // nil for token-based auth
	KeyData  []byte // nil for token-based auth
	Token    string // empty for cert-based auth
}

type argoCDConfig struct {
	BearerToken     string          `json:"bearerToken,omitempty"`
	TLSClientConfig tlsClientConfig `json:"tlsClientConfig"`
}

type tlsClientConfig struct {
	CAData   string `json:"caData"`
	CertData string `json:"certData,omitempty"`
	KeyData  string `json:"keyData,omitempty"`
}

// Parse extracts connection info from a kubeconfig YAML byte slice.
// Returns a non-retryable error if the kubeconfig is invalid or missing required fields.
func Parse(data []byte) (*Parsed, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("kubeconfig is empty")
	}
	cfg, err := clientcmd.Load(data)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}

	ctx, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok || ctx == nil {
		return nil, fmt.Errorf("current context %q not found", cfg.CurrentContext)
	}

	cluster, ok := cfg.Clusters[ctx.Cluster]
	if !ok || cluster == nil {
		return nil, fmt.Errorf("cluster %q not found", ctx.Cluster)
	}
	if cluster.Server == "" {
		return nil, fmt.Errorf("cluster %q has no server URL", ctx.Cluster)
	}
	if len(cluster.CertificateAuthorityData) == 0 {
		return nil, fmt.Errorf("cluster %q has no CA data", ctx.Cluster)
	}

	user, ok := cfg.AuthInfos[ctx.AuthInfo]
	if !ok || user == nil {
		return nil, fmt.Errorf("user %q not found", ctx.AuthInfo)
	}

	p := &Parsed{
		Server: cluster.Server,
		CAData: cluster.CertificateAuthorityData,
	}

	switch {
	case user.Token != "":
		p.Token = user.Token
	case len(user.ClientCertificateData) > 0 && len(user.ClientKeyData) > 0:
		p.CertData = user.ClientCertificateData
		p.KeyData = user.ClientKeyData
	default:
		return nil, fmt.Errorf("user %q has no supported credentials (need token or client cert/key)", ctx.AuthInfo)
	}

	return p, nil
}

// ToArgoCDConfigJSON returns the JSON blob for the ArgoCD cluster secret's "config" data key.
func (p *Parsed) ToArgoCDConfigJSON() ([]byte, error) {
	cfg := argoCDConfig{
		TLSClientConfig: tlsClientConfig{
			CAData: base64.StdEncoding.EncodeToString(p.CAData),
		},
	}
	if p.Token != "" {
		cfg.BearerToken = p.Token
	} else {
		cfg.TLSClientConfig.CertData = base64.StdEncoding.EncodeToString(p.CertData)
		cfg.TLSClientConfig.KeyData = base64.StdEncoding.EncodeToString(p.KeyData)
	}
	return json.Marshal(cfg)
}
