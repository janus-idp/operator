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
	"time"

	"redhat-developer/red-hat-developer-hub-operator/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"

	"redhat-developer/red-hat-developer-hub-operator/pkg/model"

	bsv1alpha1 "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = When("create default backstage", func() {

	It("creates runtime objects", func() {

		ctx := context.Background()
		ns := createNamespace(ctx)
		backstageName := createAndReconcileBackstage(ctx, ns, bsv1alpha1.BackstageSpec{}, "")

		Eventually(func(g Gomega) {
			deploy := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, deploy)
			g.Expect(err).ShouldNot(HaveOccurred(), controllerMessage())

			By("creating /opt/app-root/src/dynamic-plugins.xml ")
			appConfig := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DynamicPluginsDefaultName(backstageName)}, appConfig)
			g.Expect(err).ShouldNot(HaveOccurred())

			g.Expect(deploy.Spec.Template.Spec.InitContainers).To(HaveLen(1))
			_, initCont := model.DynamicPluginsInitContainer(deploy.Spec.Template.Spec.InitContainers)
			//deploy.Spec.Template.Spec.InitContainers[0]
			g.Expect(initCont.VolumeMounts).To(HaveLen(3))
			g.Expect(initCont.VolumeMounts[0].MountPath).To(Equal("/dynamic-plugins-root"))
			g.Expect(initCont.VolumeMounts[0].SubPath).To(BeEmpty())
			g.Expect(initCont.VolumeMounts[1].MountPath).To(Equal("/opt/app-root/src/.npmrc.dynamic-plugins"))
			g.Expect(initCont.VolumeMounts[1].SubPath).To(Equal(".npmrc"))
			g.Expect(initCont.VolumeMounts[2].MountPath).To(Equal("/opt/app-root/src/dynamic-plugins.yaml"))
			g.Expect(initCont.VolumeMounts[2].SubPath).To(Equal("dynamic-plugins.yaml"))
			g.Expect(initCont.VolumeMounts[2].Name).
				To(Equal(utils.GenerateVolumeNameFromCmOrSecret(model.DynamicPluginsDefaultName(backstageName))))
			g.Expect(initCont.VolumeMounts[2].SubPath).To(Equal(model.DynamicPluginsFile))

			g.Expect(initCont.Env[0].Name).To(Equal("NPM_CONFIG_USERCONFIG"))
			g.Expect(initCont.Env[0].Value).To(Equal("/opt/app-root/src/.npmrc.dynamic-plugins"))

			g.Expect(deploy.Spec.Template.Spec.Volumes).To(HaveLen(4))
			g.Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			mainCont := deploy.Spec.Template.Spec.Containers[0]
			g.Expect(mainCont.Args).To(HaveLen(4))
			g.Expect(mainCont.Args[0]).To(Equal("--config"))
			g.Expect(mainCont.Args[1]).To(Equal("dynamic-plugins-root/app-config.dynamic-plugins.yaml"))
			g.Expect(mainCont.Args[2]).To(Equal("--config"))
			g.Expect(mainCont.Args[3]).To(Equal("/opt/app-root/src/default.app-config.yaml"))

			g.Expect(mainCont.VolumeMounts).To(HaveLen(2))
			g.Expect(mainCont.VolumeMounts[0].MountPath).To(Equal("/opt/app-root/src/dynamic-plugins-root"))
			g.Expect(mainCont.VolumeMounts[0].SubPath).To(BeEmpty())
			g.Expect(mainCont.VolumeMounts[1].MountPath).To(Equal("/opt/app-root/src/default.app-config.yaml"))
			g.Expect(mainCont.VolumeMounts[1].SubPath).To(Equal("default.app-config.yaml"))

		}, 10*time.Second, time.Second).Should(Succeed())

		deleteNamespace(ctx, ns)
	})
})
