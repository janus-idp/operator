package integration_tests

import (
	"context"
	"time"

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

	//Context("and the book is available", func() {
	//	It("lends it to the reader", func(ctx SpecContext) {
	//
	//	}, SpecTimeout(time.Second * 5))
	//})

	It("also creates runtime objects", func() {

		backstageName := createBackstage(ctx, bsv1alpha1.BackstageSpec{}, ns)

		By("Checking if the custom resource was successfully created")
		found := &bsv1alpha1.Backstage{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{Name: backstageName, Namespace: ns}, found)
		}, time.Minute, time.Second).Should(Succeed())

		_, err := NewTestBackstageReconciler(ns).ReconcileLocalCluster(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: backstageName, Namespace: ns},
		})
		Expect(err).To(Not(HaveOccurred()))

		By("creating a secret for accessing the Database")
		found1 := &corev1.Secret{}
		Eventually(func(g Gomega) {
			name := model.DbSecretDefaultName(backstageName)
			err := k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, found1)
			g.Expect(err).ShouldNot(HaveOccurred())

		}, time.Minute, time.Second).Should(Succeed())

	})
})
