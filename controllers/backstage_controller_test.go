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

package controller

import (
	"context"
	"fmt"
	"time"

	"janus-idp.io/backstage-operator/pkg/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
)

var _ = Describe("Backstage controller", func() {
	var (
		ctx                 context.Context
		ns                  string
		backstageName       string
		backstageReconciler *BackstageReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()
		ns = fmt.Sprintf("ns-%d-%s", GinkgoParallelProcess(), randString(5))
		backstageName = "test-backstage-" + randString(5)

		By("Creating the Namespace to perform the tests")
		err := k8sClient.Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ns,
				Namespace: ns,
			},
		})
		Expect(err).To(Not(HaveOccurred()))

		backstageReconciler = &BackstageReconciler{
			Client:      k8sClient,
			Scheme:      k8sClient.Scheme(),
			Namespace:   ns,
			OwnsRuntime: true,
		}
	})

	AfterEach(func() {
		// NOTE: Be aware of the current delete namespace limitations.
		// More info: https://book.kubebuilder.io/reference/envtest.html#testing-considerations
		By("Deleting the Namespace to perform the tests")
		_ = k8sClient.Delete(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ns,
				Namespace: ns,
			},
		})
	})

	buildBackstageCR := func(spec bsv1alpha1.BackstageSpec) *bsv1alpha1.Backstage {
		return &bsv1alpha1.Backstage{
			ObjectMeta: metav1.ObjectMeta{
				Name:      backstageName,
				Namespace: ns,
			},
			Spec: spec,
		}
	}

	//buildConfigMap := func(name string, data map[string]string) *corev1.ConfigMap {
	//	return &corev1.ConfigMap{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: "v1",
	//			Kind:       "ConfigMap",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      name,
	//			Namespace: ns,
	//		},
	//		Data: data,
	//	}
	//}

	//buildSecret := func(name string, data map[string][]byte) *corev1.Secret {
	//	return &corev1.Secret{
	//		TypeMeta: metav1.TypeMeta{
	//			APIVersion: "v1",
	//			Kind:       "Secret",
	//		},
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name:      name,
	//			Namespace: ns,
	//		},
	//		Data: data,
	//	}
	//}

	//verifyBackstageInstance := func(ctx context.Context) {
	//	Eventually(func(g Gomega) {
	//		var backstage bsv1alpha1.Backstage
	//		err := k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, &backstage)
	//		g.Expect(err).NotTo(HaveOccurred())
	//		//TODO the status is under construction
	//		g.Expect(len(backstage.Status.Conditions)).To(Equal(2))
	//	}, time.Minute, time.Second).Should(Succeed())
	//}

	//findEnvVar := func(envVars []corev1.EnvVar, key string) (corev1.EnvVar, bool) {
	//	return findElementByPredicate(envVars, func(envVar corev1.EnvVar) bool {
	//		return envVar.Name == key
	//	})
	//}

	//findVolume := func(vols []corev1.Volume, name string) (corev1.Volume, bool) {
	//	return findElementByPredicate(vols, func(vol corev1.Volume) bool {
	//		return vol.Name == name
	//	})
	//}
	//
	//findVolumeMount := func(mounts []corev1.VolumeMount, name string) (corev1.VolumeMount, bool) {
	//	return findElementByPredicate(mounts, func(mount corev1.VolumeMount) bool {
	//		return mount.Name == name
	//	})
	//}

	When("creating default CR with no spec", func() {
		var backstage *bsv1alpha1.Backstage
		BeforeEach(func() {
			backstage = buildBackstageCR(bsv1alpha1.BackstageSpec{})
			err := k8sClient.Create(ctx, backstage)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should successfully reconcile a custom resource for default Backstage", func() {
			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &bsv1alpha1.Backstage{}
				return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			_, err := backstageReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
			})
			Expect(err).To(Not(HaveOccurred()))

			By("Checking if Deployment was successfully created in the reconciliation")
			found := &appsv1.Deployment{}
			Eventually(func() error {
				// TODO to get name from default
				return k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: utils.GenerateRuntimeObjectName(backstageName, "deployment")}, found)
			}, time.Minute, time.Second).Should(Succeed())

			//By("Checking that the Deployment is configured with a random backend auth secret")
			//backendSecretEnvVar, ok := findEnvVar(found.Spec.Template.Spec.Containers[0].Env, "BACKEND_SECRET")
			//Expect(ok).To(BeTrue(), "env var BACKEND_SECRET not found in main container")
			//Expect(backendSecretEnvVar.ValueFrom.SecretKeyRef.Name).To(
			//	Not(BeEmpty()), "'name' for backend auth secret ref should not be empty")
			//Expect(backendSecretEnvVar.ValueFrom.SecretKeyRef.Key).To(
			//	Equal("backend-secret"), "Unexpected secret key ref for backend secret")
			//Expect(backendSecretEnvVar.ValueFrom.SecretKeyRef.Optional).To(HaveValue(BeFalse()),
			//	"'optional' for backend auth secret ref should be 'false'")
			//
			//backendAuthAppConfigEnvVar, ok := findEnvVar(found.Spec.Template.Spec.Containers[0].Env, "APP_CONFIG_backend_auth_keys")
			//Expect(ok).To(BeTrue(), "env var APP_CONFIG_backend_auth_keys not found in main container")
			//Expect(backendAuthAppConfigEnvVar.Value).To(Equal(`[{"secret": "$(BACKEND_SECRET)"}]`))
		})
	})
})

func findElementByPredicate[T any](l []T, predicate func(t T) bool) (T, bool) {
	for _, v := range l {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}
