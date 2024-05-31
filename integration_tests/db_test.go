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

	"redhat-developer/red-hat-developer-hub-operator/pkg/model"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"

	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	bsv1alpha1 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = When("create backstage with CR configured", func() {

	var (
		ctx context.Context
		ns  string
	)

	BeforeEach(func() {
		ctx = context.Background()
		ns = createNamespace(ctx)
	})

	AfterEach(func() {
		deleteNamespace(ctx, ns)
	})

	It("creates default Backstage and then update CR to not to use local DB", func() {
		backstageName := createAndReconcileBackstage(ctx, ns, bsv1alpha1.BackstageSpec{}, "")

		Eventually(func(g Gomega) {
			By("creating Deployment with database.enableLocalDb=true by default")

			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)}, &appsv1.StatefulSet{})
			g.Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbServiceName(backstageName)}, &corev1.Service{})
			g.Expect(err).To(Not(HaveOccurred()))

			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbSecretDefaultName(backstageName)}, &corev1.Secret{})
			g.Expect(err).To(Not(HaveOccurred()))

		}, time.Minute, time.Second).Should(Succeed())

		By("updating Backstage")
		update := &bsv1alpha1.Backstage{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, update)
		Expect(err).To(Not(HaveOccurred()))
		update.Spec.Database = &bsv1alpha1.Database{}
		update.Spec.Database.EnableLocalDb = ptr.To(false)
		err = k8sClient.Update(ctx, update)
		Expect(err).To(Not(HaveOccurred()))
		_, err = NewTestBackstageReconciler(ns).ReconcileAny(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		Eventually(func(g Gomega) {
			By("deleting Local Db StatefulSet, Service and Secret")
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)}, &appsv1.StatefulSet{})
			g.Expect(err).To(HaveOccurred())
			g.Expect(errors.IsNotFound(err))

			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbServiceName(backstageName)}, &corev1.Service{})
			g.Expect(err).To(HaveOccurred())
			g.Expect(errors.IsNotFound(err))

			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbSecretDefaultName(backstageName)}, &corev1.Secret{})
			g.Expect(err).To(HaveOccurred())
			g.Expect(errors.IsNotFound(err))
		}, time.Minute, time.Second).Should(Succeed())

	})

	It("creates Backstage with disabled local DB and secret", func() {
		backstageName := createAndReconcileBackstage(ctx, ns, bsv1alpha1.BackstageSpec{
			Database: &bsv1alpha1.Database{
				EnableLocalDb:  ptr.To(false),
				AuthSecretName: "existing-secret",
			},
		}, "")

		Eventually(func(g Gomega) {
			By("not creating a StatefulSet for the Database")
			err := k8sClient.Get(ctx,
				types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)},
				&appsv1.StatefulSet{})
			g.Expect(err).Should(HaveOccurred())
			g.Expect(errors.IsNotFound(err))

			By("Checking if Deployment was successfully created in the reconciliation")
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, &appsv1.Deployment{})
			g.Expect(err).Should(Not(HaveOccurred()))
		}, time.Minute, time.Second).Should(Succeed())
	})

	It("creates Backstage with disabled local DB no secret", func() {
		backstageName := createAndReconcileBackstage(ctx, ns, bsv1alpha1.BackstageSpec{
			Database: &bsv1alpha1.Database{
				EnableLocalDb: ptr.To(false),
			},
		}, "")

		Eventually(func(g Gomega) {
			By("not creating a StatefulSet for the Database")
			err := k8sClient.Get(ctx,
				types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)},
				&appsv1.StatefulSet{})
			g.Expect(err).Should(HaveOccurred())
			g.Expect(errors.IsNotFound(err))

			By("Checking if Deployment was successfully created in the reconciliation")
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, &appsv1.Deployment{})
			g.Expect(err).Should(Not(HaveOccurred()))
		}, time.Minute, time.Second).Should(Succeed())
	})

	When("reconciling with already existing DB resources", func() {
		var backstageName string
		var dbSecretName string
		var dbServiceName, dbServiceHLName string
		var dbStatefulSetName string

		BeforeEach(func() {
			if !*testEnv.UseExistingCluster {
				Skip("Real cluster required to assert actual deletion and replacement of resources")
			}

			/*
				Simulates DB resources created by the Operator in 1.1. List of resources from 1.1.x for reference:

				NAME                                                                      NAMESPACE  AGE
				backstage.rhdh.redhat.com/bs1                                             my-ns      6m49s
				configmap/bs1-auth-app-config                                             my-ns      6m47s
				configmap/bs1-dynamic-plugins                                             my-ns      6m47s
				configmap/kube-root-ca.crt                                                my-ns      52m
				configmap/openshift-service-ca.crt                                        my-ns      52m
				controllerrevision.apps/backstage-psql-bs1-7d85d479f5                     my-ns      6m48s
				deployment.apps/backstage-bs1                                             my-ns      6m47s
				endpoints/backstage-bs1                                                   my-ns      6m47s
				endpoints/backstage-psql-bs1-hl                                           my-ns      6m48s
				endpoints/backstage-psql-bs1                                              my-ns      6m48s
				endpointslice.discovery.k8s.io/backstage-bs1-9b7zd                        my-ns      6m47s
				endpointslice.discovery.k8s.io/backstage-psql-bs1-85h8j                   my-ns      6m48s
				endpointslice.discovery.k8s.io/backstage-psql-bs1-hl-82tfx                my-ns      6m48s
				persistentvolumeclaim/backstage-bs1-fb7c547df-mqbtv-dynamic-plugins-root  my-ns      6m47s
				persistentvolumeclaim/data-backstage-psql-bs1-0                           my-ns      6m48s
				pod/backstage-bs1-fb7c547df-mqbtv                                         my-ns      6m47s
				pod/backstage-psql-bs1-0                                                  my-ns      6m48s
				replicaset.apps/backstage-bs1-fb7c547df                                   my-ns      6m47s
				rolebinding.authorization.openshift.io/admin                              my-ns      52m
				rolebinding.authorization.openshift.io/system:deployers                   my-ns      52m
				rolebinding.authorization.openshift.io/system:image-builders              my-ns      52m
				rolebinding.authorization.openshift.io/system:image-pullers               my-ns      52m
				rolebinding.rbac.authorization.k8s.io/admin                               my-ns      52m
				rolebinding.rbac.authorization.k8s.io/system:deployers                    my-ns      52m
				rolebinding.rbac.authorization.k8s.io/system:image-builders               my-ns      52m
				rolebinding.rbac.authorization.k8s.io/system:image-pullers                my-ns      52m
				route.route.openshift.io/backstage-bs1                                    my-ns      6m46s
				secret/backstage-psql-secret-bs1                                          my-ns      6m48s
				secret/builder-dockercfg-mqj7w                                            my-ns      52m
				secret/builder-token-86wp4                                                my-ns      52m
				secret/default-dockercfg-ln6nj                                            my-ns      52m
				secret/default-token-fv8m5                                                my-ns      52m
				secret/deployer-dockercfg-dv8f6                                           my-ns      52m
				secret/deployer-token-wsk85                                               my-ns      52m
				serviceaccount/builder                                                    my-ns      52m
				serviceaccount/default                                                    my-ns      52m
				serviceaccount/deployer                                                   my-ns      52m
				service/backstage-bs1                                                     my-ns      6m47s
				service/backstage-psql-bs1-hl                                             my-ns      6m48s
				service/backstage-psql-bs1                                                my-ns      6m48s
				statefulset.apps/backstage-psql-bs1                                       my-ns      6m48s
			*/

			backstageName = createBackstage(ctx, bsv1alpha1.BackstageSpec{}, ns, "")
			dbServiceName = fmt.Sprintf("backstage-psql-%s", backstageName)
			dbServiceHLName = fmt.Sprintf("backstage-psql-%s-hl", backstageName)
			dbSecretName = fmt.Sprintf("backstage-psql-secret-%s", backstageName)
			dbStatefulSetName = fmt.Sprintf("backstage-psql-%s", backstageName)

			err := k8sClient.Create(ctx, &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbServiceName,
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"rhdh.redhat.com/app": backstageName,
					},
					Type: corev1.ServiceTypeClusterIP,
					Ports: []corev1.ServicePort{
						{Port: 5432},
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbServiceHLName,
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"rhdh.redhat.com/app": backstageName,
					},
					ClusterIP: corev1.ClusterIPNone,
					Ports: []corev1.ServicePort{
						{Port: 5432},
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbSecretName,
					Namespace: ns,
				},
				StringData: map[string]string{
					"POSTGRES_HOST":             dbServiceName,
					"POSTGRES_PORT":             "5432",
					"POSTGRES_USER":             "postgresql",
					"POSTGRES_PASSWORD":         "my-awesome-password",       //notsecret
					"POSTGRESQL_ADMIN_PASSWORD": "my-super-awesome-password", //notsecret
				},
			})
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dbStatefulSetName,
					Namespace: ns,
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"rhdh.redhat.com/app": dbStatefulSetName,
						},
					},
					ServiceName: dbServiceHLName,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"rhdh.redhat.com/app": dbStatefulSetName,
							},
							Name: dbStatefulSetName,
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "postgresql",
									Image: "quay.io/fedora/postgresql-15:latest",
									EnvFrom: []corev1.EnvFromSource{
										{
											SecretRef: &corev1.SecretEnvSource{
												LocalObjectReference: corev1.LocalObjectReference{
													Name: dbSecretName,
												},
											},
										},
									},
									Ports: []corev1.ContainerPort{
										{ContainerPort: 5432},
									},
								},
							},
						},
					},
					VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "data",
							},
							Spec: corev1.PersistentVolumeClaimSpec{
								AccessModes: []corev1.PersistentVolumeAccessMode{
									corev1.ReadWriteOnce,
								},
								Resources: corev1.VolumeResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceStorage: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should reuse existing DB resources when reconciling", func() {
			reconcileBackstage(ctx, ns, backstageName)

			Eventually(func(g Gomega) {
				By("Checking if Deployment was successfully created in the reconciliation")
				err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, &appsv1.Deployment{})
				g.Expect(err).Should(Not(HaveOccurred()))
			}, time.Minute, time.Second).Should(Succeed())

			// Second run should recreate it
			reconcileBackstage(ctx, ns, backstageName)

			Eventually(func(g Gomega) {
				By("recreating the StatefulSet with the existing name")
				var statefulSetList appsv1.StatefulSetList
				err := k8sClient.List(ctx, &statefulSetList, &ctrlruntimeclient.ListOptions{Namespace: ns})
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(statefulSetList.Items).To(HaveLen(1))
				g.Expect(statefulSetList.Items[0].Name).To(Equal(dbStatefulSetName))
				// TODO(rm3l): this should be the name of the headless service (which should also be created), but currently this is hardcoded
				g.Expect(statefulSetList.Items[0].Spec.ServiceName).To(Equal("backstage-psql-cr1-hl"))
				g.Expect(statefulSetList.Items[0].Spec.Template.Spec.Containers).To(HaveLen(1))
				g.Expect(dbSecretName).To(BeEnvFromForContainer(statefulSetList.Items[0].Spec.Template.Spec.Containers[0]))
				g.Expect(statefulSetList.Items[0].Spec.Template.Spec.Containers[0].SecurityContext).ToNot(BeNil())
				g.Expect(statefulSetList.Items[0].Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup).ToNot(BeNil())
				g.Expect(*statefulSetList.Items[0].Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup).To(BeEquivalentTo(0))
				g.Expect(statefulSetList.Items[0].Spec.VolumeClaimTemplates).To(HaveLen(1))
				g.Expect(statefulSetList.Items[0].Spec.VolumeClaimTemplates[0].Name).To(Equal("data"))
			}, time.Minute, 2*time.Second).Should(Succeed())
		})
	})
})
