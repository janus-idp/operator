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
	"fmt"
	"os"
	"path/filepath"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
)

// testBackstageObject it is a helper object to simplify testing model component allowing to customize and isolate testing configuration
// usual sequence of creating testBackstageObject contains such a steps:
// createBackstageTest(bsv1alpha1.Backstage).
// withDefaultConfig(useDef bool)
// addToDefaultConfig(key, fileName)
type testBackstageObject struct {
	backstage    bsv1alpha1.Backstage
	detailedSpec *DetailedBackstageSpec
	scheme       *runtime.Scheme
}

// simple bsv1alpha1.Backstage
func simpleTestBackstage() bsv1alpha1.Backstage {
	return bsv1alpha1.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bs",
			Namespace: "ns123",
		},
		Spec: bsv1alpha1.BackstageSpec{
			Database: &bsv1alpha1.Database{
				EnableLocalDb: pointer.Bool(false),
			},
		},
	}

}

// initialises testBackstageObject object
func createBackstageTest(bs bsv1alpha1.Backstage) *testBackstageObject {
	b := &testBackstageObject{backstage: bs, detailedSpec: &DetailedBackstageSpec{BackstageSpec: bs.Spec}, scheme: runtime.NewScheme()}
	utilruntime.Must(bsv1alpha1.AddToScheme(b.scheme))
	b.detailedSpec.RawConfigContent = map[string]string{}
	return b
}

// enables LocalDB
func (b *testBackstageObject) withLocalDb(secretName string) *testBackstageObject {
	b.detailedSpec.Database.EnableLocalDb = pointer.Bool(true)
	//if secretName == "" {
	//	secretName =
	//}

	//if dbSecret == nil {
	//	if name == "" {
	//		b.detailedSpec.LocalDbSecret = GenerateDbSecret()
	//	} else {
	//		b.detailedSpec.LocalDbSecret = NewDbSecretFromSpec(name)
	//	}
	//	return b
	//}
	//
	//b.detailedSpec.LocalDbSecret = *dbSecret
	return b
}

// tells if object should use default Backstage Deployment/Service configuration from ./testdata/default-config or not
func (b *testBackstageObject) withDefaultConfig(useDef bool) *testBackstageObject {
	if useDef {
		// here we have default-config folder
		_ = os.Setenv("LOCALBIN", "./testdata")
	} else {
		_ = os.Setenv("LOCALBIN", ".")
	}
	return b
}

// adds particular part of configuration pointing to configuration key
// where key is configuration key (such as "deployment.yaml" and fileName is a name of additional conf file in ./testdata
func (b *testBackstageObject) addToDefaultConfig(key string, fileName string) *testBackstageObject {

	yaml, err := readTestYamlFile(fileName)
	if err != nil {
		panic(err)
	}
	b.detailedSpec.RawConfigContent[key] = string(yaml)

	return b
}

// reads file from ./testdata
func readTestYamlFile(name string) ([]byte, error) {

	b, err := os.ReadFile(filepath.Join("testdata", name)) // #nosec G304, path is constructed internally
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	return b, nil
}
