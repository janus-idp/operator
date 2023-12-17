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

// Add additional details to the Backstage Spec helping in making Bakstage Objects Model
// Validates Backstage Spec and fails fast if something not correct
func (r *BackstageReconciler) preprocessSpec(ctx context.Context, bsSpec bs.BackstageSpec) (*model.DetailedBackstageSpec, error) {
	//lg := log.FromContext(ctx)

	result := &model.DetailedBackstageSpec{
		BackstageSpec: bsSpec,
	}

	// Process RawRuntimeConfig
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

	// Process AppConfigs
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

	// TODO extra objects

	return result, nil
}
