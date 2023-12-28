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

package model

import (
	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Mandatory        needType = "Mandatory"
	Optional         needType = "Optional"
	ForLocalDatabase needType = "ForLocalDatabase"
	ForOpenshift     needType = "ForOpenshift"
)

type needType string

type ObjectConfig struct {
	ObjectFactory ObjectFactory
	Key           string
	need          needType
}

type ObjectFactory interface {
	newBackstageObject() BackstageObject
}

type BackstageObject interface {
	Object() client.Object
	initMetainfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool)
	// needed only for check if Object exists to call KubeClient.Get() and it should be garbage collected right away
	EmptyObject() client.Object
	addToModel(model *runtimeModel)
}

type BackstagePodContributor interface {
	BackstageObject
	updateBackstagePod(pod *backstagePod)
}

type LocalDbPodContributor interface {
	BackstageObject
	updateLocalDbPod(model *runtimeModel)
}
