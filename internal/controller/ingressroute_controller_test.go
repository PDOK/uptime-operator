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
	"fmt"

	m "github.com/PDOK/uptime-operator/internal/model"
	"github.com/PDOK/uptime-operator/internal/service"
	. "github.com/onsi/ginkgo/v2" //nolint:revive // ginkgo bdd
	. "github.com/onsi/gomega"    //nolint:revive // gingko bdd
	traefikcontainous "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	testIngress   = "test-ingress-resource"
	testNamespace = "default"
)

type testUptimeProvider struct {
	checks map[string]m.UptimeCheck
}

func newTestUptimeProvider() *testUptimeProvider {
	return &testUptimeProvider{
		checks: make(map[string]m.UptimeCheck),
	}
}

func (m *testUptimeProvider) HasCheck(check m.UptimeCheck) bool {
	_, ok := m.checks[check.ID]
	return ok
}

func (m *testUptimeProvider) CreateOrUpdateCheck(check m.UptimeCheck) error {
	m.checks[check.ID] = check
	return nil
}

func (m *testUptimeProvider) DeleteCheck(check m.UptimeCheck) error {
	delete(m.checks, check.ID)
	return nil
}

var ingressRouteWithUptimeCheck = &traefikcontainous.IngressRoute{
	TypeMeta: v1.TypeMeta{},
	ObjectMeta: v1.ObjectMeta{
		Name:      testIngress,
		Namespace: testNamespace,
		Annotations: map[string]string{
			// with uptime check annotations
			m.AnnotationID:   "y45735y375",
			m.AnnotationURL:  "https://test.example",
			m.AnnotationName: "Test uptime check",
		},
	},
	Spec: traefikcontainous.IngressRouteSpec{
		Routes: []traefikcontainous.Route{
			{
				Kind:  "Rule",
				Match: "Host(`localhost`)",
				Services: []traefikcontainous.Service{
					{
						LoadBalancerSpec: traefikcontainous.LoadBalancerSpec{
							Name: "test",
						},
					},
				},
			},
		},
	},
}

var _ = Describe("IngressRoute Controller", func() {
	Context("When reconciling IngressRoutes", func() {
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      testIngress,
			Namespace: testNamespace,
		}

		It("Should successfully create + update an uptime check for an ingress route", func() {
			testProvider := newTestUptimeProvider()
			controllerReconciler := &IngressRouteReconciler{
				Client:             k8sClient,
				Scheme:             k8sClient.Scheme(),
				UptimeCheckService: service.New(service.WithProvider(testProvider)),
			}

			By("Creating an IngressRoute")
			newIngressRoute := &traefikcontainous.IngressRoute{}
			err := k8sClient.Get(ctx, typeNamespacedName, newIngressRoute)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					resource := ingressRouteWithUptimeCheck.DeepCopy()
					Expect(k8sClient.Create(ctx, resource)).To(Succeed())
					Expect(k8sClient.Get(ctx, typeNamespacedName, newIngressRoute)).To(Succeed())
				} else {
					Fail(fmt.Sprintf("%v", err))
				}
			}

			By("Reconciling the IngressRoute (thus creating an uptime check)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(testProvider.checks).To(ContainElement(m.UptimeCheck{
				ID:   "y45735y375",
				URL:  "https://test.example",
				Name: "Test uptime check",
				Tags: []string{"managed-by-uptime-operator"},
			}))

			By("Fetching and updating IngressRoute (adding extra uptime annotation)")
			fetchedIngressRoute := &traefikcontainous.IngressRoute{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, fetchedIngressRoute)
				return err == nil
			}, "10s", "1s").Should(BeTrue())
			fetchedIngressRoute.Annotations[m.AnnotationStringContains] = "OK"
			Expect(k8sClient.Update(ctx, fetchedIngressRoute)).Should(Succeed())

			By("Reconciling the IngressRoute again (to make sure uptime check is updated)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(testProvider.checks).To(ContainElement(m.UptimeCheck{
				ID:             "y45735y375",
				URL:            "https://test.example",
				Name:           "Test uptime check",
				Tags:           []string{"managed-by-uptime-operator"},
				StringContains: "OK",
			}))

			By("Reconciling the IngressRoute again to make sure it doesn't cause any side effects")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(testProvider.checks).To(HaveLen(1))
		})

		It("Should delete uptime check for an existing ingress route", func() {
			testProvider := newTestUptimeProvider()
			controllerReconciler := &IngressRouteReconciler{
				Client:             k8sClient,
				Scheme:             k8sClient.Scheme(),
				UptimeCheckService: service.New(service.WithProvider(testProvider)),
			}

			By("Reconciling the IngressRoute (expecting on is available from previous test)")
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(testProvider.checks).To(ContainElement(m.UptimeCheck{
				ID:             "y45735y375",
				URL:            "https://test.example",
				Name:           "Test uptime check",
				Tags:           []string{"managed-by-uptime-operator"},
				StringContains: "OK",
			}))

			By("Delete IngressRoute")
			fetchedIngressRoute := &traefikcontainous.IngressRoute{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, typeNamespacedName, fetchedIngressRoute)
				return err == nil
			}, "10s", "1s").Should(BeTrue())
			Expect(k8sClient.Delete(ctx, fetchedIngressRoute)).To(Succeed())

			By("Reconciling the IngressRoute again (to make sure uptime check is deleted)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(testProvider.checks).To(BeEmpty())
		})
	})
})
