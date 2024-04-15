/*
MIT License

Copyright (c) 2024 Publieke Dienstverlening op de Kaart

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controller

import (
	"context"

	"github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/provider"
	traefikcontainous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	traefikio "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IngressRouteReconciler reconciles Traefik IngressRoutes with an uptime monitoring (SaaS) provider
type IngressRouteReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	UptimeProvider provider.UptimeProvider
}

//+kubebuilder:rbac:groups=traefik.containo.us,resources=ingressroutes,verbs=get;list;watch
//+kubebuilder:rbac:groups=traefik.containo.us,resources=ingressroutes/finalizers,verbs=update
//+kubebuilder:rbac:groups=traefik.io,resources=ingressroutes,verbs=get;list;watch
//+kubebuilder:rbac:groups=traefik.io,resources=ingressroutes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *IngressRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	annotations, err := r.getAnnotations(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	r.syncWithUptimeProvider(ctx, annotations)
	return ctrl.Result{}, nil
}

func (r *IngressRouteReconciler) getAnnotations(ctx context.Context, req ctrl.Request) (map[string]string, error) {
	// first reconcile on "traefik.containo.us/v1alpha1" ingress
	ingressContainous := &traefikcontainous.IngressRoute{}
	if err := r.Get(ctx, req.NamespacedName, ingressContainous); err != nil {
		// not found, now reconcile on "traefik.io/v1alpha1" ingress
		ingressIo := &traefikio.IngressRoute{}
		if err = r.Get(ctx, req.NamespacedName, ingressIo); err != nil {
			// still not found, handle error
			logger := log.FromContext(ctx)
			if apierrors.IsNotFound(err) {
				logger.Info("IngressRoute resource not found", "name", req.NamespacedName)
			} else {
				logger.Error(err, "unable to fetch IngressRoute resource", "error", err)
			}
			return nil, err
		}
		return ingressIo.Annotations, nil
	}
	return ingressContainous.Annotations, nil
}

func (r *IngressRouteReconciler) syncWithUptimeProvider(ctx context.Context, annotations map[string]string) {
	logger := log.FromContext(ctx)
	check := model.NewUptimeCheck(annotations)
	if check != nil {
		logger.Info("syncing uptime check with id", "id", check.ID)
		err := r.UptimeProvider.CreateOrUpdateCheck(*check)
		if err != nil {
			logger.Error(err, "failed syncing uptime check", "error", err)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	preCondition := predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})

	return ctrl.NewControllerManagedBy(mgr).
		Named("uptime-operator").
		Watches(
			&traefikcontainous.IngressRoute{}, // watch "traefik.containo.us/v1alpha1" ingresses
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(preCondition)).
		Watches(
			&traefikio.IngressRoute{}, // watch "traefik.io/v1alpha1" ingresses
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(preCondition)).
		Complete(r)
}
