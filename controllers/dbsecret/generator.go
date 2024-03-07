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

package dbsecret

import (
	"context"
	"fmt"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/log"

	bs "redhat-developer/red-hat-developer-hub-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"redhat-developer/red-hat-developer-hub-operator/pkg/utils"

	"redhat-developer/red-hat-developer-hub-operator/pkg/model"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Generate(ctx context.Context, client client.Client, backstage bs.Backstage, dbservice *model.DbService, scheme *runtime.Scheme) error {

	pswd, _ := utils.GeneratePassword(24)
	service := dbservice.Object().(*corev1.Service)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.DbSecretDefaultName(backstage.Name),
			Namespace: backstage.Namespace,
		},
		StringData: map[string]string{
			"POSTGRES_PASSWORD":         pswd,
			"POSTGRESQL_ADMIN_PASSWORD": pswd,
			"POSTGRES_USER":             "postgres",
			"POSTGRES_HOST":             service.GetName(),
			"POSTGRES_PORT":             strconv.FormatInt(int64(service.Spec.Ports[0].Port), 10),
		},
	}
	if err := controllerutil.SetControllerReference(&backstage, secret, scheme); err != nil {
		//error should never have happened,
		//otherwise the Operator has invalid (not a runtime.Object) or non-registered type.
		//In both cases it will fail before this place
		panic(err)
	}
	if err := client.Create(ctx, secret); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create object %w", err)
	}

	log.FromContext(ctx).V(1).Info("DBSecret created", "", secret.Name, "ownerref", len(secret.OwnerReferences))

	return nil
}
