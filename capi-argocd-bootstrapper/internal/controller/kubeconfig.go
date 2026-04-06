package controller

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/fergalhk-lab/apps/capi-argocd-bootstrapper/internal/kubeconfig"
)

const (
	labelCAPHEnvironment  = "caph.environment"
	labelClusterName      = "cluster.x-k8s.io/cluster-name"
	labelManagedBy        = "capi-argocd-bootstrapper/managed"
	labelArgoCDSecretType = "argocd.argoproj.io/secret-type"
	annotationSourceNS    = "capi-argocd-bootstrapper/source-namespace"
	annotationSourceName  = "capi-argocd-bootstrapper/source-name"
	argoCDNamespace       = "argocd"
)

// KubeconfigReconciler watches CAPI kubeconfig secrets and reconciles ArgoCD cluster secrets.
type KubeconfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *KubeconfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var secret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return r.handleDeletion(ctx, req.NamespacedName)
		}
		return ctrl.Result{}, err
	}

	parsed, err := kubeconfig.Parse(secret.Data["value"])
	if err != nil {
		logger.Error(err, "permanent parse error, skipping", "secret", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	clusterName := secret.Labels[labelClusterName]
	if clusterName == "" {
		logger.Error(nil, "cluster name label is empty, skipping", "secret", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	configJSON, err := parsed.ToArgoCDConfigJSON()
	if err != nil {
		logger.Error(err, "permanent config serialisation error, skipping")
		return ctrl.Result{}, nil
	}

	desired := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: argoCDNamespace,
			Labels: map[string]string{
				labelArgoCDSecretType: "cluster",
				labelManagedBy:        "true",
			},
			Annotations: map[string]string{
				annotationSourceNS:   req.Namespace,
				annotationSourceName: req.Name,
			},
		},
		Data: map[string][]byte{
			"name":   []byte(clusterName),
			"server": []byte(parsed.Server),
			"config": configJSON,
		},
	}

	var existing corev1.Secret
	err = r.Get(ctx, types.NamespacedName{Namespace: argoCDNamespace, Name: clusterName}, &existing)
	if apierrors.IsNotFound(err) {
		if err := r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, fmt.Errorf("create ArgoCD secret: %w", err)
		}
		logger.Info("created ArgoCD cluster secret", "cluster", clusterName)
		return ctrl.Result{}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	if existing.Labels[labelManagedBy] != "true" {
		logger.Info("ArgoCD secret not managed by us, skipping", "cluster", clusterName)
		return ctrl.Result{}, nil
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Data = desired.Data
	if err := r.Update(ctx, &existing); err != nil {
		return ctrl.Result{}, fmt.Errorf("update ArgoCD secret: %w", err)
	}
	logger.Info("updated ArgoCD cluster secret", "cluster", clusterName)
	return ctrl.Result{}, nil
}

func (r *KubeconfigReconciler) handleDeletion(ctx context.Context, source types.NamespacedName) (ctrl.Result, error) {
	var list corev1.SecretList
	if err := r.List(ctx, &list,
		client.InNamespace(argoCDNamespace),
		client.MatchingLabels{labelManagedBy: "true"},
	); err != nil {
		return ctrl.Result{}, err
	}
	for i := range list.Items {
		s := &list.Items[i]
		if s.Annotations[annotationSourceNS] == source.Namespace &&
			s.Annotations[annotationSourceName] == source.Name {
			if err := r.Delete(ctx, s); err != nil && !apierrors.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("delete ArgoCD secret: %w", err)
			}
			log.FromContext(ctx).Info("deleted ArgoCD cluster secret", "cluster", s.Name)
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *KubeconfigReconciler) mapArgoCDSecretToSource(_ context.Context, obj client.Object) []reconcile.Request {
	ann := obj.GetAnnotations()
	if ann == nil {
		return nil
	}
	ns := ann[annotationSourceNS]
	name := ann[annotationSourceName]
	if ns == "" || name == "" {
		return nil
	}
	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{Namespace: ns, Name: name}},
	}
}

func isKubeconfigSecret(obj client.Object) bool {
	if !strings.HasSuffix(obj.GetName(), "-kubeconfig") {
		return false
	}
	labels := obj.GetLabels()
	if labels[labelCAPHEnvironment] != "owned" {
		return false
	}
	_, ok := labels[labelClusterName]
	return ok
}

func kubeconfigPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return isKubeconfigSecret(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return isKubeconfigSecret(e.ObjectNew) || isKubeconfigSecret(e.ObjectOld) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return isKubeconfigSecret(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return isKubeconfigSecret(e.Object) },
	}
}

func argoCDManagedPredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(obj client.Object) bool {
		return obj.GetNamespace() == argoCDNamespace && obj.GetLabels()[labelManagedBy] == "true"
	})
}

// SetupWithManager registers the controller with the manager.
func (r *KubeconfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(kubeconfigPredicate())).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.mapArgoCDSecretToSource),
			builder.WithPredicates(argoCDManagedPredicate()),
		).
		Complete(r)
}
