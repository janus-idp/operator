package model

import (
	"context"
	"testing"

	"k8s.io/utils/pointer"

	"github.com/stretchr/testify/assert"
	"janus-idp.io/backstage-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSpecifiedAppConfig(t *testing.T) {

	bs := v1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},

		Spec: v1alpha1.BackstageSpec{
			Application: &v1alpha1.Application{
				AppConfig: &v1alpha1.AppConfig{
					MountPath: "/test",
					ConfigMaps: []v1alpha1.ObjectKeyRef{
						{
							Name: "test-app-config",
						},
					},
				},
			},
			EnableLocalDb: pointer.Bool(true),
		},
	}

	model, err := InitObjects(context.TODO(), bs, &DetailedBackstageSpec{BackstageSpec: bs.Spec}, true, false)

	assert.NoError(t, err)
	assert.True(t, len(model) > 0)

	deployment := getBackstageDeployment(model)
	assert.NotNil(t, deployment)

	assert.Equal(t, 2, len(deployment.deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
	assert.Equal(t, 4, len(deployment.deployment.Spec.Template.Spec.Containers[0].Args))
	assert.Equal(t, 3, len(deployment.deployment.Spec.Template.Spec.Volumes))

}
