package controller

import (
	"context"
	"fmt"
	"path/filepath"

	bs "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *BackstageReconciler) preprocessSpec(ctx context.Context, bsSpec bs.BackstageSpec) (*model.DetailedBackstageSpec, error) {
	result := &model.DetailedBackstageSpec{
		BackstageSpec: bsSpec,
	}

	if bsSpec.RawRuntimeConfig != "" {
		cm := corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Name: bsSpec.RawRuntimeConfig, Namespace: r.Namespace}, &cm); err != nil {
			return nil, fmt.Errorf("failed to load rawConfig %s: %w", bsSpec.RawRuntimeConfig, err)
		}
		for key, value := range cm.Data {
			result.Details.RawConfig[key] = value
		}
	} else {
		result.Details.RawConfig = map[string]string{}
	}

	if bsSpec.Application != nil && bsSpec.Application.AppConfig != nil {
		mountPath := bsSpec.Application.AppConfig.MountPath
		for _, ac := range bsSpec.Application.AppConfig.ConfigMaps {
			cm := corev1.ConfigMap{}
			if err := r.Get(ctx, types.NamespacedName{Name: ac.Name, Namespace: r.Namespace}, &cm); err != nil {
				return nil, fmt.Errorf("failed to load configMap %s: %w", ac.Name, err)
			}

			for key := range cm.Data {
				// first key added
				result.Details.AppConfigs = append(result.Details.AppConfigs, model.AppConfigDetails{
					ConfigMapName: cm.Name,
					FilePath:      filepath.Join(mountPath, key),
				})
			}
		}
	}

	return result, nil
}
