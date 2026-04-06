package main

import (
	"flag"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/fergalhk-lab/apps/capi-argocd-bootstrapper/internal/controller"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var argoCDNamespace string
	flag.StringVar(&argoCDNamespace, "argocd-namespace", "argocd", "Namespace where ArgoCD cluster secrets are managed")
	flag.Parse()

	ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: "0"},
		HealthProbeBindAddress: ":8081",
	})
	if err != nil {
		log.Fatalf("create manager: %v", err)
	}

	if err := (&controller.KubeconfigReconciler{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		ArgoCDNamespace: argoCDNamespace,
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("setup controller: %v", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Fatalf("set up health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Fatalf("set up ready check: %v", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("manager: %v", err)
	}
}
