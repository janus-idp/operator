package model

import (
	"context"
	"fmt"
	"testing"

	"janus-idp.io/backstage-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

// NOTE: to make it work locally env var LOCALBIN should point to the directory where default-config folder located
func TestInitDefaultDeploy(t *testing.T) {

	bs := v1alpha1.Backstage{
		//TypeMeta: metav1.TypeMeta{
		//	APIVersion: "v1alpha1",
		//	Kind:       "Backstage",
		//},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
	}

	model, err := InitObjects(context.TODO(), bs, &DetailedBackstageSpec{}, true)

	assert.NoError(t, err)
	assert.Equal(t, 4, len(model))
	assert.Equal(t, "bs-deployment", model[0].Object().GetName())
	assert.Equal(t, "ns123", model[0].Object().GetNamespace())
	assert.Equal(t, 2, len(model[0].Object().GetLabels()))
	//	assert.Equal(t, 1, len(model[0].Object().GetOwnerReferences()))

	bsDeployment := model[0].(*BackstageDeployment)
	assert.NotNil(t, bsDeployment.pod.container)
	assert.Equal(t, backstageContainerName, bsDeployment.pod.container.Name)
	assert.NotNil(t, bsDeployment.pod.volumes)

	//	assert.Equal(t, "Backstage", bsDeployment.deployment.OwnerReferences[0].Kind)

	bsService := model[1].(*BackstageService)
	assert.Equal(t, "bs-service", bsService.service.Name)
	assert.True(t, len(bsService.service.Spec.Ports) > 0)

	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsDeployment.deployment.Spec.Template.ObjectMeta.Labels[backstageAppLabel])
	assert.Equal(t, fmt.Sprintf("backstage-%s", "bs"), bsService.service.Spec.Selector[backstageAppLabel])

}
