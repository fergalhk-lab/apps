package kubeconfig_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fergalhk-lab/apps/capi-argocd-bootstrapper/internal/kubeconfig"
)

var (
	testCA   = []byte("test-ca-cert-data")
	testCert = []byte("test-client-cert-data")
	testKey  = []byte("test-client-key-data")
)

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func certKubeconfig(server string) []byte {
	return []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: test-user
  user:
    client-certificate-data: %s
    client-key-data: %s
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
`, server, b64(testCA), b64(testCert), b64(testKey)))
}

func tokenKubeconfig(server string) []byte {
	return []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: test-user
  user:
    token: test-bearer-token
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
`, server, b64(testCA)))
}

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		wantErr    bool
		wantServer string
		wantCA     []byte
		wantCert   []byte
		wantKey    []byte
		wantToken  string
	}{
		{
			name:       "cert auth kubeconfig",
			input:      certKubeconfig("https://1.2.3.4:6443"),
			wantServer: "https://1.2.3.4:6443",
			wantCA:     testCA,
			wantCert:   testCert,
			wantKey:    testKey,
		},
		{
			name:       "token auth kubeconfig",
			input:      tokenKubeconfig("https://1.2.3.4:6443"),
			wantServer: "https://1.2.3.4:6443",
			wantCA:     testCA,
			wantToken:  "test-bearer-token",
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "malformed YAML",
			input:   []byte("not: valid: yaml: {{{"),
			wantErr: true,
		},
		{
			name: "missing server URL",
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    certificate-authority-data: %s
users:
- name: test-user
  user:
    token: mytoken
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
`, b64(testCA))),
			wantErr: true,
		},
		{
			name: "missing credentials",
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://1.2.3.4:6443
    certificate-authority-data: %s
users:
- name: test-user
  user: {}
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
`, b64(testCA))),
			wantErr: true,
		},
		{
			name:    "context not in map",
			input:   []byte(`apiVersion: v1
kind: Config
clusters: []
users: []
contexts: []
current-context: nonexistent
`),
			wantErr: true,
		},
		{
			name:       "no current-context with single context",
			wantServer: "https://1.2.3.4:6443",
			wantCA:     testCA,
			wantToken:  "test-bearer-token",
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://1.2.3.4:6443
    certificate-authority-data: %s
users:
- name: test-user
  user:
    token: test-bearer-token
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
`, b64(testCA))),
		},
		{
			name:    "no current-context with multiple contexts",
			wantErr: true,
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://1.2.3.4:6443
    certificate-authority-data: %s
users:
- name: test-user
  user:
    token: test-bearer-token
contexts:
- name: context-one
  context:
    cluster: test-cluster
    user: test-user
- name: context-two
  context:
    cluster: test-cluster
    user: test-user
`, b64(testCA))),
		},
		{
			name:    "no current-context with no contexts",
			wantErr: true,
			input:   []byte(`apiVersion: v1
kind: Config
clusters: []
users: []
contexts: []
`),
		},
		{
			name: "cluster not in map",
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters: []
users:
- name: test-user
  user:
    token: mytoken
contexts:
- name: test-context
  context:
    cluster: missing-cluster
    user: test-user
current-context: test-context
`)),
			wantErr: true,
		},
		{
			name: "missing CA data",
			input: []byte(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://1.2.3.4:6443
users:
- name: test-user
  user:
    token: mytoken
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: test-user
current-context: test-context
`),
			wantErr: true,
		},
		{
			name: "user not in map",
			input: []byte(fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: test-cluster
  cluster:
    server: https://1.2.3.4:6443
    certificate-authority-data: %s
users: []
contexts:
- name: test-context
  context:
    cluster: test-cluster
    user: missing-user
current-context: test-context
`, b64(testCA))),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kubeconfig.Parse(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantServer, got.Server)
			require.Equal(t, tt.wantCA, got.CAData)
			require.Equal(t, tt.wantCert, got.CertData)
			require.Equal(t, tt.wantKey, got.KeyData)
			require.Equal(t, tt.wantToken, got.Token)
		})
	}
}

func TestToArgoCDConfigJSON(t *testing.T) {
	tests := []struct {
		name      string
		parsed    *kubeconfig.Parsed
		wantToken string
		wantCert  bool
	}{
		{
			name: "cert auth",
			parsed: &kubeconfig.Parsed{
				Server:   "https://1.2.3.4:6443",
				CAData:   testCA,
				CertData: testCert,
				KeyData:  testKey,
			},
			wantCert: true,
		},
		{
			name: "token auth",
			parsed: &kubeconfig.Parsed{
				Server: "https://1.2.3.4:6443",
				CAData: testCA,
				Token:  "mytoken",
			},
			wantToken: "mytoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.parsed.ToArgoCDConfigJSON()
			require.NoError(t, err)

			var cfg map[string]interface{}
			require.NoError(t, json.Unmarshal(data, &cfg))

			tls, ok := cfg["tlsClientConfig"].(map[string]interface{})
			require.True(t, ok)
			require.Equal(t, b64(testCA), tls["caData"])

			if tt.wantToken != "" {
				require.Equal(t, tt.wantToken, cfg["bearerToken"])
				require.Nil(t, tls["certData"])
				require.Nil(t, tls["keyData"])
			} else {
				require.Equal(t, b64(testCert), tls["certData"])
				require.Equal(t, b64(testKey), tls["keyData"])
				require.Nil(t, cfg["bearerToken"])
			}
		})
	}
}
