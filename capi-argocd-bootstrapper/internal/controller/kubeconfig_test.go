package controller_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/stretchr/testify/require"

	"github.com/fergalhk-lab/apps/capi-argocd-bootstrapper/internal/controller"
)

const testTimeout = 5 * time.Second

var (
	k8sClient  client.Client
	testCtx    context.Context
	testCancel context.CancelFunc
	testCA     = []byte("test-ca-cert-data")
	testCert   = []byte("test-client-cert-data")
	testKey    = []byte("test-client-key-data")
)

func TestMain(m *testing.M) {
	testEnv := &envtest.Environment{}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(fmt.Sprintf("start envtest: %v", err))
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		panic(fmt.Sprintf("create client: %v", err))
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme})
	if err != nil {
		panic(fmt.Sprintf("create manager: %v", err))
	}

	if err := (&controller.KubeconfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		panic(fmt.Sprintf("setup controller: %v", err))
	}

	testCtx, testCancel = context.WithCancel(context.Background())
	go func() {
		if err := mgr.Start(testCtx); err != nil {
			panic(fmt.Sprintf("manager: %v", err))
		}
	}()

	if err := k8sClient.Create(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "argocd"},
	}); err != nil {
		panic(fmt.Sprintf("create argocd namespace: %v", err))
	}

	code := m.Run()

	testCancel()
	if err := testEnv.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "stop envtest: %v\n", err)
	}
	os.Exit(code)
}

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

func makeKubeconfigSecret(ns, name, clusterName string, kcfg []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"caph.environment":               "owned",
				"cluster.x-k8s.io/cluster-name": clusterName,
			},
		},
		Data: map[string][]byte{"value": kcfg},
	}
}

func waitFor(t *testing.T, condition func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(testTimeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func getArgoCDSecret(clusterName string) (*corev1.Secret, error) {
	var s corev1.Secret
	err := k8sClient.Get(testCtx, types.NamespacedName{Namespace: "argocd", Name: clusterName}, &s)
	return &s, err
}

func createNS(t *testing.T, name string) {
	t.Helper()
	require.NoError(t, k8sClient.Create(testCtx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}))
	t.Cleanup(func() {
		k8sClient.Delete(testCtx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}) //nolint:errcheck
	})
}

func assertArgoCDSecretCreated(t *testing.T, clusterName, server string) *corev1.Secret {
	t.Helper()
	require.True(t, waitFor(t, func() bool {
		_, err := getArgoCDSecret(clusterName)
		return err == nil
	}), "ArgoCD secret not created within timeout")

	s, err := getArgoCDSecret(clusterName)
	require.NoError(t, err)
	require.Equal(t, "true", s.Labels["capi-argocd-bootstrapper/managed"])
	require.Equal(t, "cluster", s.Labels["argocd.argoproj.io/secret-type"])
	require.Equal(t, clusterName, string(s.Data["name"]))
	require.Equal(t, server, string(s.Data["server"]))

	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(s.Data["config"], &cfg))
	tls := cfg["tlsClientConfig"].(map[string]interface{})
	require.Equal(t, b64(testCA), tls["caData"])
	require.Equal(t, b64(testCert), tls["certData"])
	require.Equal(t, b64(testKey), tls["keyData"])
	return s
}

func TestReconcileCreate(t *testing.T) {
	createNS(t, "test-create")
	secret := makeKubeconfigSecret("test-create", "cluster-a-kubeconfig", "cluster-a",
		certKubeconfig("https://1.1.1.1:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))
	t.Cleanup(func() { k8sClient.Delete(testCtx, secret) })                                                                                              //nolint:errcheck
	t.Cleanup(func() { k8sClient.Delete(testCtx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "cluster-a", Namespace: "argocd"}}) }) //nolint:errcheck

	assertArgoCDSecretCreated(t, "cluster-a", "https://1.1.1.1:6443")
}

func TestReconcileUpdate(t *testing.T) {
	createNS(t, "test-update")
	secret := makeKubeconfigSecret("test-update", "cluster-b-kubeconfig", "cluster-b",
		certKubeconfig("https://2.2.2.2:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))
	t.Cleanup(func() { k8sClient.Delete(testCtx, secret) })                                                                                              //nolint:errcheck
	t.Cleanup(func() { k8sClient.Delete(testCtx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "cluster-b", Namespace: "argocd"}}) }) //nolint:errcheck

	assertArgoCDSecretCreated(t, "cluster-b", "https://2.2.2.2:6443")

	require.NoError(t, k8sClient.Get(testCtx, types.NamespacedName{Namespace: "test-update", Name: "cluster-b-kubeconfig"}, secret))
	secret.Data["value"] = certKubeconfig("https://3.3.3.3:6443")
	require.NoError(t, k8sClient.Update(testCtx, secret))

	require.True(t, waitFor(t, func() bool {
		s, err := getArgoCDSecret("cluster-b")
		return err == nil && string(s.Data["server"]) == "https://3.3.3.3:6443"
	}), "ArgoCD secret not updated within timeout")
}

func TestReconcileDelete(t *testing.T) {
	createNS(t, "test-delete")
	secret := makeKubeconfigSecret("test-delete", "cluster-c-kubeconfig", "cluster-c",
		certKubeconfig("https://4.4.4.4:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))

	assertArgoCDSecretCreated(t, "cluster-c", "https://4.4.4.4:6443")

	require.NoError(t, k8sClient.Delete(testCtx, secret))

	require.True(t, waitFor(t, func() bool {
		_, err := getArgoCDSecret("cluster-c")
		return err != nil // expect not found
	}), "ArgoCD secret not deleted within timeout")
}

func TestReconcileRestoreExternallyDeleted(t *testing.T) {
	createNS(t, "test-restore-del")
	secret := makeKubeconfigSecret("test-restore-del", "cluster-d-kubeconfig", "cluster-d",
		certKubeconfig("https://5.5.5.5:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))
	t.Cleanup(func() { k8sClient.Delete(testCtx, secret) })                                                                                              //nolint:errcheck
	t.Cleanup(func() { k8sClient.Delete(testCtx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "cluster-d", Namespace: "argocd"}}) }) //nolint:errcheck

	assertArgoCDSecretCreated(t, "cluster-d", "https://5.5.5.5:6443")

	require.NoError(t, k8sClient.Delete(testCtx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "cluster-d", Namespace: "argocd"},
	}))

	require.True(t, waitFor(t, func() bool {
		_, err := getArgoCDSecret("cluster-d")
		return err == nil
	}), "ArgoCD secret not recreated within timeout")
}

func TestReconcileRestoreExternallyModified(t *testing.T) {
	createNS(t, "test-restore-mod")
	secret := makeKubeconfigSecret("test-restore-mod", "cluster-e-kubeconfig", "cluster-e",
		certKubeconfig("https://6.6.6.6:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))
	t.Cleanup(func() { k8sClient.Delete(testCtx, secret) })                                                                                              //nolint:errcheck
	t.Cleanup(func() { k8sClient.Delete(testCtx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "cluster-e", Namespace: "argocd"}}) }) //nolint:errcheck

	assertArgoCDSecretCreated(t, "cluster-e", "https://6.6.6.6:6443")

	argoSecret, err := getArgoCDSecret("cluster-e")
	require.NoError(t, err)
	argoSecret.Data["server"] = []byte("https://tampered:6443")
	require.NoError(t, k8sClient.Update(testCtx, argoSecret))

	require.True(t, waitFor(t, func() bool {
		s, err := getArgoCDSecret("cluster-e")
		return err == nil && string(s.Data["server"]) == "https://6.6.6.6:6443"
	}), "ArgoCD secret not restored within timeout")
}

func TestReconcileUnmanagedConflict(t *testing.T) {
	createNS(t, "test-conflict")

	preExisting := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-f",
			Namespace: "argocd",
			Labels:    map[string]string{"argocd.argoproj.io/secret-type": "cluster"},
		},
		Data: map[string][]byte{"server": []byte("https://original:6443")},
	}
	require.NoError(t, k8sClient.Create(testCtx, preExisting))
	t.Cleanup(func() { k8sClient.Delete(testCtx, preExisting) }) //nolint:errcheck

	secret := makeKubeconfigSecret("test-conflict", "cluster-f-kubeconfig", "cluster-f",
		certKubeconfig("https://7.7.7.7:6443"))
	require.NoError(t, k8sClient.Create(testCtx, secret))
	t.Cleanup(func() { k8sClient.Delete(testCtx, secret) }) //nolint:errcheck

	// Give the reconciler enough time to run
	time.Sleep(2 * time.Second)

	s, err := getArgoCDSecret("cluster-f")
	require.NoError(t, err)
	require.Equal(t, "https://original:6443", string(s.Data["server"]))
	require.Empty(t, s.Labels["capi-argocd-bootstrapper/managed"])
}
