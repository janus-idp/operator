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

// Need Identifier for configuration object
// Used on initialization phase to let initializer know what to do if configuration object
// of the certain type is not found
const (
	// Mandatory for Backstage deployment, initialization fails
	Mandatory needType = "Mandatory"
	// Optional for Backstage deployment (for example config parameters), initialization continues
	Optional needType = "Optional"
	// Mandatory if Local database Enabled, initialization fails if LocalDB enabled, ignored otherwise
	ForLocalDatabase needType = "ForLocalDatabase"
	// Used for Openshift cluster only, ignored otherwise
	ForOpenshift needType = "ForOpenshift"
)

type needType string

// Registered Object configuring Backstage deployment
type ObjectConfig struct {
	// Factory to create the object
	ObjectFactory ObjectFactory
	// Unique key identifying the "kind" of Object which also is the name of config file.
	// For example: "deployment.yaml" containing configuration of Backstage Deployment
	Key string
	// Need identifier
	need needType
}

type ObjectFactory interface {
	newBackstageObject() BackstageObject
}

// Abstraction for the model Backstage object taking part in deployment
type BackstageObject interface {
	// underlying Kubernetes object
	Object() client.Object
	// Inits metadata. Typically used to set/change object name, labels, selectors to ensure integrity
	//setMetaInfo(backstageMeta bsv1alpha1.Backstage, ownsRuntime bool)
	// needed only for check if Object exists to call KubeClient.Get() and it should be garbage collected right away
	EmptyObject() client.Object
	// (For some types Backstage objects), adds it to the model
	addToModel(model *RuntimeModel, backstageMeta bsv1alpha1.Backstage, ownsRuntime bool)
	// at this stage all the information is updated
	// set the final references validates the object at the end of initialization (after 3 phases)
	validate(model *RuntimeModel) error
}

// BackstageObject contributing to Backstage pod. Usually app-config related
type BackstagePodContributor interface {
	BackstageObject
	updateBackstagePod(pod *backstagePod)
}

// BackstageObject contributing to Local DB pod
//type LocalDbPodContributor interface {
//	BackstageObject
//	updateLocalDbPod(model *RuntimeModel)
//}
