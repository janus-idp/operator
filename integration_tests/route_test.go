//
// Copyright (c) 2023 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_tests

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"redhat-developer/red-hat-developer-hub-operator/pkg/model"

	bsv1alpha1 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = When("create default backstage", func() {

	var (
		ctx context.Context
		ns  string
		//backstageName string
	)

	BeforeEach(func() {
		ctx = context.Background()
		ns = createNamespace(ctx)
	})

	AfterEach(func() {
		// NOTE: Be aware of the current delete namespace limitations.
		// More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
		_ = k8sClient.Delete(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
	})

	It("creates Backstage object (on Openshift)", func() {

		backstageName := createBackstage(ctx, bsv1alpha1.BackstageSpec{
			Application: &bsv1alpha1.Application{
				Route: &bsv1alpha1.Route{},
			},
		}, ns)

		Eventually(func() error {
			found := &bsv1alpha1.Backstage{}
			return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
		}, time.Minute, time.Second).Should(Succeed())

		_, err := NewTestBackstageReconciler(ns).ReconcileLocalCluster(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		Eventually(func(g Gomega) {

			By("creating Deployment")
			deploy := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, deploy)
			g.Expect(err).ShouldNot(HaveOccurred())
			Expect(deploy.Spec.Replicas).To(HaveValue(BeEquivalentTo(1)))

			By("setting Backstage status")
			bs := &bsv1alpha1.Backstage{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: backstageName}, bs)
			g.Expect(err).ShouldNot(HaveOccurred())
			fmt.Printf(">>>>>>>>>>>>>>>>>>>> %v\n", bs.Spec.Application.Route)

			// TODO better matcher for Conditions
			g.Expect(bs.Status.Conditions[0].Reason).To(Equal("Deployed"))

			for _, cond := range deploy.Status.Conditions {
				if cond.Type == "Available" {
					g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				}
			}

		}, time.Minute, time.Second).Should(Succeed())

	})
})
