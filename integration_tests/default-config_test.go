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
	"redhat-developer/red-hat-developer-hub-operator/pkg/utils"
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

	It("creates runtime objects", func() {

		backstageName := createBackstage(ctx, bsv1alpha1.BackstageSpec{}, ns)

		Eventually(func() error {
			found := &bsv1alpha1.Backstage{}
			return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
		}, time.Minute, time.Second).Should(Succeed())

		_, err := NewTestBackstageReconciler(ns).ReconcileAny(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		Eventually(func(g Gomega) {
			By("creating a secret for accessing the Database")
			secret := &corev1.Secret{}
			secretName := model.DbSecretDefaultName(backstageName)
			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: secretName}, secret)
			g.Expect(err).ShouldNot(HaveOccurred(), controllerMessage())
			g.Expect(len(secret.Data)).To(Equal(5))
			g.Expect(secret.Data).To(HaveKeyWithValue("POSTGRES_USER", []uint8("postgres")))

			By("creating a StatefulSet for the Database")
			ss := &appsv1.StatefulSet{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)}, ss)
			g.Expect(err).ShouldNot(HaveOccurred())

			By("injecting default DB Secret as an env var for Db container")
			g.Expect(model.DbSecretDefaultName(backstageName)).To(BeEnvFromForContainer(ss.Spec.Template.Spec.Containers[0]))
			g.Expect(ss.GetOwnerReferences()).To(HaveLen(1))

			By("creating a Service for the Database")
			err = k8sClient.Get(ctx, types.NamespacedName{Name: model.DbServiceName(backstageName), Namespace: ns}, &corev1.Service{})
			g.Expect(err).To(Not(HaveOccurred()))

			By("creating Deployment")
			deploy := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, deploy)
			g.Expect(err).ShouldNot(HaveOccurred())
			Expect(deploy.Spec.Replicas).To(HaveValue(BeEquivalentTo(1)))

			By("creating OwnerReference to all the runtime objects")
			or := deploy.GetOwnerReferences()
			g.Expect(or).To(HaveLen(1))
			g.Expect(or[0].Name).To(Equal(backstageName))

			By("creating default app-config")
			appConfig := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.AppConfigDefaultName(backstageName)}, appConfig)
			g.Expect(err).ShouldNot(HaveOccurred())

			By("mounting Volume defined in default app-config")
			g.Expect(utils.GenerateVolumeNameFromCmOrSecret(model.AppConfigDefaultName(backstageName))).
				To(BeAddedAsVolumeToPodSpec(deploy.Spec.Template.Spec))

			By("setting Backstage status")
			bs := &bsv1alpha1.Backstage{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: backstageName}, bs)
			g.Expect(err).ShouldNot(HaveOccurred())
			// TODO better matcher for Conditions
			g.Expect(bs.Status.Conditions[0].Reason).To(Equal("Deployed"))

			for _, cond := range deploy.Status.Conditions {
				if cond.Type == "Available" {
					g.Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				}
			}

		}, 5*time.Minute, time.Second).Should(Succeed())

	})

	It("creates runtime object using raw configuration ", func() {

		bsConf := map[string]string{
			"deployment.yaml": `
apiVersion: apps/v1
kind: Deployment
metadata:
 name: bs1-deployment
 labels:
   app: bs1
spec:
 replicas: 1
 selector:
   matchLabels:
     app: bs1
 template:
   metadata:
     labels:
       app: bs1
   spec:
     containers:
       - name: bs1
         image: busybox
`,
		}

		dbConf := map[string]string{
			"db-statefulset.yaml": `
apiVersion: apps/v1
kind: StatefulSet
metadata:
 name: db-statefulset
spec:
 replicas: 1
 selector:
   matchLabels:
     app: db
 template:
   metadata:
     labels:
       app: db
   spec:
     containers:
       - name: db
         image: busybox
`,
		}
		generateConfigMap(ctx, k8sClient, "bsraw", ns, bsConf)
		generateConfigMap(ctx, k8sClient, "dbraw", ns, dbConf)

		backstageName := createBackstage(ctx, bsv1alpha1.BackstageSpec{
			RawRuntimeConfig: &bsv1alpha1.RuntimeConfig{
				BackstageConfigName: "bsraw",
				LocalDbConfigName:   "dbraw",
			},
		}, ns)

		Eventually(func() error {
			found := &bsv1alpha1.Backstage{}
			return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
		}, time.Minute, time.Second).Should(Succeed())

		_, err := NewTestBackstageReconciler(ns).ReconcileAny(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		Eventually(func(g Gomega) {
			By("creating Deployment")
			deploy := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, deploy)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(deploy.Spec.Replicas).To(HaveValue(BeEquivalentTo(1)))
			g.Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			g.Expect(deploy.Spec.Template.Spec.Containers[0].Image).To(Equal("busybox"))

			By("creating StatefulSet")
			ss := &appsv1.StatefulSet{}
			name := model.DbStatefulSetName(backstageName)
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, ss)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(ss.Spec.Template.Spec.Containers).To(HaveLen(1))
			g.Expect(ss.Spec.Template.Spec.Containers[0].Image).To(Equal("busybox"))
		}, time.Minute, time.Second).Should(Succeed())

	})
})
