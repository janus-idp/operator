package integration_tests

import (
	"context"
	"time"

	"janus-idp.io/backstage-operator/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"janus-idp.io/backstage-operator/pkg/model"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
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

		By("Checking if the custom resource was successfully created")

		Eventually(func() error {
			found := &bsv1alpha1.Backstage{}
			return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
		}, time.Minute, time.Second).Should(Succeed())

		_, err := NewTestBackstageReconciler(ns).ReconcileLocalCluster(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		Eventually(func(g Gomega) {
			By("creating a secret for accessing the Database")
			secret := &corev1.Secret{}
			secretName := model.DbSecretDefaultName(backstageName)
			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: secretName}, secret)
			g.Expect(err).ShouldNot(HaveOccurred())

			By("creating a StatefulSet for the Database")
			ss := &appsv1.StatefulSet{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DbStatefulSetName(backstageName)}, ss)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(getEnvFromSecret(ss.Spec.Template.Spec.Containers[0], model.DbSecretDefaultName(backstageName))).ToNot(BeNil())
			g.Expect(ss.GetOwnerReferences()).To(HaveLen(1))

			By("checking if Deployment was successfully created in the reconciliation")
			deploy := &appsv1.Deployment{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.DeploymentName(backstageName)}, deploy)
			g.Expect(err).ShouldNot(HaveOccurred())
			By("checking the number of replicas")
			Expect(deploy.Spec.Replicas).To(HaveValue(BeEquivalentTo(1)))

			By("creating default app-config")
			appConfig := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: model.AppConfigDefaultName(backstageName)}, appConfig)
			g.Expect(err).ShouldNot(HaveOccurred())
			_, ok := findVolume(deploy.Spec.Template.Spec.Volumes, utils.GenerateVolumeNameFromCmOrSecret(model.AppConfigDefaultName(backstageName)))
			g.Expect(ok).To(BeTrue())
			g.Expect(appConfig.GetOwnerReferences()).To(HaveLen(1))

		}, time.Minute, time.Second).Should(Succeed())

	})
})
