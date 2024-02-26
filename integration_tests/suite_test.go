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
	"os"
	"strconv"

	"janus-idp.io/backstage-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"

	"path/filepath"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/utils/pointer"

	controller "janus-idp.io/backstage-operator/controllers"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/rand"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var testOnExistingCluster = false

type TestBackstageReconciler struct {
	controller.BackstageReconciler
	namespace string
}

func init() {
	rand.Seed(time.Now().UnixNano())
	testOnExistingCluster, _ = strconv.ParseBool(os.Getenv("TEST_ON_EXISTING_CLUSTER"))
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {
	//logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	testEnv.UseExistingCluster = pointer.Bool(testOnExistingCluster)

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = bsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createBackstage(ctx context.Context, spec bsv1alpha1.BackstageSpec, ns string) string {
	backstageName := "test-backstage-" + randString(5)
	err := k8sClient.Create(ctx, &bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backstageName,
			Namespace: ns,
		},
		Spec: spec,
	})
	Expect(err).To(Not(HaveOccurred()))
	return backstageName
}

func createNamespace(ctx context.Context) string {
	ns := fmt.Sprintf("ns-%d-%s", GinkgoParallelProcess(), randString(5))
	err := k8sClient.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	})
	Expect(err).To(Not(HaveOccurred()))
	return ns
}

func NewTestBackstageReconciler(namespace string) *TestBackstageReconciler {

	var (
		isOpenshift bool
		err         error
	)
	if *testEnv.UseExistingCluster {
		isOpenshift, err = utils.IsOpenshift()
		Expect(err).To(Not(HaveOccurred()))
	} else {
		isOpenshift = false
	}

	return &TestBackstageReconciler{BackstageReconciler: controller.BackstageReconciler{
		Client:      k8sClient,
		Scheme:      k8sClient.Scheme(),
		OwnsRuntime: true,
		// let's set it explicitly to avoid misunderstanding
		IsOpenShift: isOpenshift,
	}, namespace: namespace}
}

func (t *TestBackstageReconciler) ReconcileLocalCluster(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if !*testEnv.UseExistingCluster {
		// Ignore requests for other namespaces, if specified.
		// To overcome a limitation of EnvTest about namespace deletion.
		// More details on https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
		if t.namespace != "" && req.Namespace != t.namespace {
			return ctrl.Result{}, nil
		}
		return t.Reconcile(ctx, req)
	} else {
		return ctrl.Result{}, nil
	}
}

func (t *TestBackstageReconciler) ReconcileAny(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Ignore requests for other namespaces, if specified.
	// To overcome a limitation of EnvTest about namespace deletion.
	// More details on https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
	if !*testEnv.UseExistingCluster && (t.namespace != "" && req.Namespace != t.namespace) {
		return ctrl.Result{}, nil
	}
	return t.Reconcile(ctx, req)
}
