package model

import (
	openshift "github.com/openshift/api/route/v1"
	"janus-idp.io/backstage-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectConfig struct {
	Object *metav1.Object
	Key    string
}

func (c *ObjectConfig) isEmpty() bool {
	return c.Object == nil
}

type RuntimeModel struct {
	BackstageDeployment      appsv1.Deployment
	BackstageService         corev1.Service
	AppConfigs               []corev1.ConfigMap
	ExtraConfigMapsToFiles   []corev1.ConfigMap
	ExtraConfigMapsToEnvVars []corev1.ConfigMap
	ExtraSecretsToFiles      []corev1.Secret
	ExtraSecretsToEnvVars    []corev1.Secret
	ExtraEnvVars             map[string]string

	LocalDbStatefulSet appsv1.StatefulSet
	LocalDbService     corev1.Service

	NetworkingRoute   openshift.Route
	NetworkingIngress networkingv1.Ingress
}

func InitObjects(backstage v1alpha1.Backstage, ns string) ([]ObjectConfig, error) {

	// 3 phases of Backstage configuration:
	// 1- load from Operator defaults, modify metadata (labels, selectors..) and namespace as needed
	// 2- overlay some/all objects with Backstage.spec.rawRuntimeConfig CM
	// 3- override some parameters defined in Backstage.spec.application
	// At the end there should be an array of runtime Objects to apply (order optimized)

	objectConfigs = make([]ObjectConfig, 12)
	// Phase 1:

	m.BackstageDeployment = deployment
	m.BackstageService = service
}
