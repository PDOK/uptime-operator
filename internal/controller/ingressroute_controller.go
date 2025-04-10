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

	m "github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service"
	traefikio "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IngressRouteReconciler reconciles Traefik IngressRoutes with an uptime monitoring (SaaS) provider
type IngressRouteReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	UptimeCheckService *service.UptimeCheckService
}

//+kubebuilder:rbac:groups=traefik.io,resources=ingressroutes,verbs=get;list;watch
//+kubebuilder:rbac:groups=traefik.io,resources=ingressroutes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *IngressRouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ingressRoute, err := r.getIngressRoute(ctx, req)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	shouldContinue, err := finalizeIfNecessary(ctx, r.Client, ingressRoute, m.AnnotationFinalizer, func() error {
		r.UptimeCheckService.Mutate(ctx, m.Delete, ingressRoute.GetName(), ingressRoute.GetAnnotations())
		return nil
	})
	if !shouldContinue || err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	r.UptimeCheckService.Mutate(ctx, m.CreateOrUpdate, ingressRoute.GetName(), ingressRoute.GetAnnotations())
	return ctrl.Result{}, nil
}

func (r *IngressRouteReconciler) getIngressRoute(ctx context.Context, req ctrl.Request) (client.Object, error) {
	// try getting "traefik.io/v1alpha1" ingress
	ingressIo := &traefikio.IngressRoute{}
	if err := r.Get(ctx, req.NamespacedName, ingressIo); err != nil {
		// still not found, handle error
		logger := log.FromContext(ctx)
		if apierrors.IsNotFound(err) {
			logger.Info("IngressRoute resource not found", "name", req.NamespacedName)
		} else {
			logger.Error(err, "unable to fetch IngressRoute resource", "error", err)
		}
		return nil, err
	}
	return ingressIo, nil
}

func finalizeIfNecessary(ctx context.Context, c client.Client, obj client.Object, finalizerName string, finalizer func() error) (shouldContinue bool, err error) {
	// not under deletion, ensure finalizer annotation
	if obj.GetDeletionTimestamp().IsZero() {
		if !controllerutil.ContainsFinalizer(obj, finalizerName) {
			controllerutil.AddFinalizer(obj, finalizerName)
			err = c.Update(ctx, obj)
			return true, err
		}
		return true, nil
	}

	// under deletion but not our finalizer annotation, do nothing
	if !controllerutil.ContainsFinalizer(obj, finalizerName) {
		return false, nil
	}

	// run finalizer and remove annotation
	if err = finalizer(); err != nil {
		return false, err
	}
	controllerutil.RemoveFinalizer(obj, finalizerName)
	err = c.Update(ctx, obj)
	return false, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	preCondition := predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})

	return ctrl.NewControllerManagedBy(mgr).
		Named(m.OperatorName).
		Watches(
			&traefikio.IngressRoute{}, // watch "traefik.io/v1alpha1" ingresses
			&handler.EnqueueRequestForObject{},
			builder.WithPredicates(preCondition)).
		Complete(r)
}
