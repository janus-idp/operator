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
	"crypto/rand"
	"encoding/base64"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	bs "redhat-developer/backstage-operator/api/v1alpha1"
)

const (
	postGresSecret                = "<POSTGRESQL_SECRET>" // #nosec G101.  This is a placeholder for a secret name not an actual secret
	_defaultPsqlMainContainerName = "postgresql"
)

func (r *BackstageReconciler) handlePsqlSecret(ctx context.Context, statefulSet *appsv1.StatefulSet, backstage *bs.Backstage) (*corev1.Secret, error) {
	secretName := getSecretNameForGeneration(statefulSet, backstage)
	if len(secretName) == 0 {
		return nil, nil
	}

	var sec corev1.Secret
	err := r.readConfigMapOrDefault(ctx, backstage.Spec.RawRuntimeConfig.BackstageConfigName, "db-secret.yaml", statefulSet.Namespace, &sec)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %s", err)
	}

	//Generate the PostGresSQL secret if it does not exist
	sec.SetName(secretName)
	sec.SetNamespace(statefulSet.Namespace)
	// Create a secret with a random value
	pwd, pwdErr := generatePassword(24)
	if pwdErr != nil {
		return nil, fmt.Errorf("failed to generate a password for the PostgreSQL database: %w", pwdErr)
	}
	sec.StringData["POSTGRES_PASSWORD"] = pwd
	sec.StringData["POSTGRESQL_ADMIN_PASSWORD"] = pwd
	sec.StringData["POSTGRES_HOST"] = getDefaultDbObjName(*backstage)
	if r.OwnsRuntime {
		// Set the ownerreferences for the secret so that when the backstage CR is deleted,
		// the generated secret is automatically deleted
		if err := controllerutil.SetControllerReference(backstage, &sec, r.Scheme); err != nil {
			return nil, fmt.Errorf(ownerRefFmt, err)
		}
	}

	err = r.Create(ctx, &sec)
	if err != nil && !errors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create secret for local PostgreSQL DB, reason: %s", err)
	}
	return nil, nil // If the secret already exists, simply return
}

func getDefaultPsqlSecretName(backstage *bs.Backstage) string {
	return fmt.Sprintf("backstage-psql-secret-%s", backstage.Name)
}

func generatePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode the password to prevent special characters
	return base64.StdEncoding.EncodeToString(bytes), nil
}

func getSecretNameForGeneration(statefulSet *appsv1.StatefulSet, backstage *bs.Backstage) string {
	// A secret for the Postgres DB will be autogenerated if and only if
	// a) the container has an envFrom entry pointing to the secret reference with special name '<POSTGRESQL_SECRET>', and
	// b) no existingDbSecret is specified in backstage CR.
	for i, c := range statefulSet.Spec.Template.Spec.Containers {
		if c.Name != _defaultPsqlMainContainerName {
			continue
		}
		for k, from := range statefulSet.Spec.Template.Spec.Containers[i].EnvFrom {
			if from.SecretRef.Name == postGresSecret {
				if len(backstage.Spec.Database.AuthSecretName) == 0 {
					from.SecretRef.Name = getDefaultPsqlSecretName(backstage)
					statefulSet.Spec.Template.Spec.Containers[i].EnvFrom[k] = from
					return from.SecretRef.Name
				} else {
					from.SecretRef.Name = backstage.Spec.Database.AuthSecretName
					statefulSet.Spec.Template.Spec.Containers[i].EnvFrom[k] = from
					break
				}
			}
		}
		break
	}
	return ""
}
