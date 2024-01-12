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

	bs "janus-idp.io/backstage-operator/api/v1alpha1"
	"janus-idp.io/backstage-operator/pkg/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Add additional details to the Backstage Spec helping in making Bakstage Objects Model
// Validates Backstage Spec and fails fast if something not correct
func (r *BackstageReconciler) preprocessSpec(ctx context.Context, bsSpec bs.BackstageSpec, ns string) (*model.DetailedBackstageSpec, error) {
	//lg := log.FromContext(ctx)

	result := &model.DetailedBackstageSpec{
		BackstageSpec:    bsSpec,
		RawConfigContent: map[string]string{},
	}

	// Process RawRuntimeConfig
	if bsSpec.RawRuntimeConfig != "" {
		cm := corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Name: bsSpec.RawRuntimeConfig, Namespace: ns}, &cm); err != nil {
			return nil, fmt.Errorf("failed to load rawConfig %s: %w", bsSpec.RawRuntimeConfig, err)
		}
		for key, value := range cm.Data {
			result.RawConfigContent[key] = value
		}
	} else {
		result.RawConfigContent = map[string]string{}
	}

	// Process AppConfigs
	if bsSpec.Application != nil && bsSpec.Application.AppConfig != nil {
		mountPath := bsSpec.Application.AppConfig.MountPath
		for _, ac := range bsSpec.Application.AppConfig.ConfigMaps {
			cm := corev1.ConfigMap{}
			if err := r.Get(ctx, types.NamespacedName{Name: ac.Name, Namespace: ns}, &cm); err != nil {
				return nil, fmt.Errorf("failed to get configMap %s: %w", ac.Name, err)
			}
			result.AddConfigObject(&model.AppConfig{ConfigMap: &cm, MountPath: mountPath, Key: ac.Key})
		}
	}

	// Process ConfigMapFiles
	if bsSpec.Application != nil && bsSpec.Application.ExtraFiles != nil && bsSpec.Application.ExtraFiles.ConfigMaps != nil {
		mountPath := bsSpec.Application.ExtraFiles.MountPath
		for _, ef := range bsSpec.Application.ExtraFiles.ConfigMaps {
			cm := corev1.ConfigMap{}
			if err := r.Get(ctx, types.NamespacedName{Name: ef.Name, Namespace: ns}, &cm); err != nil {
				return nil, fmt.Errorf("failed to get ConfigMap %s: %w", ef.Name, err)
			}
			result.AddConfigObject(&model.ConfigMapFiles{ConfigMap: &cm, MountPath: mountPath, Key: ef.Key})
		}
	}

	// Process SecretFiles
	if bsSpec.Application != nil && bsSpec.Application.ExtraFiles != nil && bsSpec.Application.ExtraFiles.Secrets != nil {
		mountPath := bsSpec.Application.ExtraFiles.MountPath
		for _, ef := range bsSpec.Application.ExtraFiles.Secrets {
			sec := corev1.Secret{}
			if err := r.Get(ctx, types.NamespacedName{Name: ef.Name, Namespace: ns}, &sec); err != nil {
				return nil, fmt.Errorf("failed to get Secret %s: %w", ef.Name, err)
			}
			result.AddConfigObject(&model.SecretFiles{Secret: &sec, MountPath: mountPath, Key: ef.Key})
		}
	}

	// Process ConfigMapEnvs
	if bsSpec.Application != nil && bsSpec.Application.ExtraEnvs != nil && bsSpec.Application.ExtraEnvs.ConfigMaps != nil {
		for _, ee := range bsSpec.Application.ExtraEnvs.ConfigMaps {
			cm := corev1.ConfigMap{}
			if err := r.Get(ctx, types.NamespacedName{Name: ee.Name, Namespace: ns}, &cm); err != nil {
				return nil, fmt.Errorf("failed to get configMap %s: %w", ee.Name, err)
			}
			result.AddConfigObject(&model.ConfigMapEnvs{ConfigMap: &cm, Key: ee.Key})
		}
	}

	// Process SecretEnvs
	if bsSpec.Application != nil && bsSpec.Application.ExtraEnvs != nil && bsSpec.Application.ExtraEnvs.Secrets != nil {
		for _, ee := range bsSpec.Application.ExtraEnvs.Secrets {
			sec := corev1.Secret{}
			if err := r.Get(ctx, types.NamespacedName{Name: ee.Name, Namespace: ns}, &sec); err != nil {
				return nil, fmt.Errorf("failed to get Secret %s: %w", ee.Name, err)
			}
			result.AddConfigObject(&model.SecretEnvs{Secret: &sec, Key: ee.Key})
		}
	}

	// Process DynamicPlugins
	if bsSpec.Application != nil {
		dynaPluginsConfig := bsSpec.Application.DynamicPluginsConfigMapName
		cm := corev1.ConfigMap{}
		if dynaPluginsConfig != "" {
			if err := r.Get(ctx, types.NamespacedName{Name: dynaPluginsConfig, Namespace: ns}, &cm); err != nil {
				return nil, fmt.Errorf("failed to get ConfigMap %s: %w", dynaPluginsConfig, err)
			}
			result.AddConfigObject(&model.DynamicPlugins{ConfigMap: &cm})
		}

	}

	// PreProcess Database
	//if bsSpec.Database != nil {
	//
	//	if authSecret := bsSpec.Database.AuthSecretName; authSecret != "" {
	//		//TODO do we need this kind of check?
	//		//sec := corev1.Secret{}
	//		//if err := r.Get(ctx, types.NamespacedName{Name: authSecret, Namespace: ns}, &sec); err != nil {
	//		//	return nil, fmt.Errorf("failed to get DB AuthSecret %s: %w", authSecret, err)
	//		//}
	//	}
	//
	//}

	// TODO PreProcess Network
	return result, nil
}
