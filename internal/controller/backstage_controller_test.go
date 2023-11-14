/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	bsv1alphav1 "backstage.io/backstage-deploy-operator/api/v1alpha1"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

func TestCreateBackstage(t *testing.T) {
	var _ = Describe("Backstage controller", func() {
		Context("Backstage controller test", func() {

			const BackstageName = "test-backstage"

			ctx := context.Background()

			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      BackstageName,
					Namespace: BackstageName,
				},
			}

			typeNamespaceName := types.NamespacedName{Name: BackstageName, Namespace: BackstageName}

			BeforeEach(func() {
				By("Creating the Namespace to perform the tests")
				err := k8sClient.Create(ctx, namespace)
				Expect(err).To(Not(HaveOccurred()))

				By("Setting the Image ENV VAR which stores the Operand image")
				err = os.Setenv("MEMCACHED_IMAGE", "example.com/image:test")
				Expect(err).To(Not(HaveOccurred()))
			})

			AfterEach(func() {
				// TODO(user): Attention if you improve this code by adding other context test you MUST
				// be aware of the current delete namespace limitations. More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
				By("Deleting the Namespace to perform the tests")
				_ = k8sClient.Delete(ctx, namespace)

				By("Removing the Image ENV VAR which stores the Operand image")
				//				_ = os.Unsetenv("BACKSTAGE_IMAGE")
			})

			It("should successfully reconcile a custom resource for default Backstage", func() {
				By("Creating the custom resource for the Kind Backstage")
				backstage := &bsv1alphav1.Backstage{}
				err := k8sClient.Get(ctx, typeNamespaceName, backstage)
				if err != nil && errors.IsNotFound(err) {
					backstage := &bsv1alphav1.Backstage{
						ObjectMeta: metav1.ObjectMeta{
							Name:      BackstageName,
							Namespace: namespace.Name,
						},
						Spec: bsv1alphav1.BackstageSpec{},
					}

					err = k8sClient.Create(ctx, backstage)
					Expect(err).To(Not(HaveOccurred()))
				}

				By("Checking if the custom resource was successfully created")
				Eventually(func() error {
					found := &bsv1alphav1.Backstage{}
					return k8sClient.Get(ctx, typeNamespaceName, found)
				}, time.Minute, time.Second).Should(Succeed())

				By("Reconciling the custom resource created")
				backstageReconciler := &BackstageReconciler{
					Client: k8sClient,
					Scheme: k8sClient.Scheme(),
				}

				_, err = backstageReconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: typeNamespaceName,
				})
				Expect(err).To(Not(HaveOccurred()))

				By("Checking if Deployment was successfully created in the reconciliation")
				Eventually(func() error {
					found := &appsv1.Deployment{}
					// TODO to get name from default
					return k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace.Name, Name: "backstage"}, found)
				}, time.Minute, time.Second).Should(Succeed())

				By("Checking the latest Status added to the Backstage instance")
				Eventually(func() error {
					//TODO the status is under construction
					err = k8sClient.Get(ctx, typeNamespaceName, backstage)
					if backstage.Status.BackstageState != "deployed" {
						return fmt.Errorf("The status is not 'deployed' '%s'", backstage.Status)
					}
					return nil
				}, time.Minute, time.Second).Should(Succeed())
			})
		})
	})
}
