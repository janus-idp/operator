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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	bsv1alpha1 "janus-idp.io/backstage-operator/api/v1alpha1"
)

type testBackstageObject struct {
	backstage    bsv1alpha1.Backstage
	detailedSpec *DetailedBackstageSpec
}

var simpleTestBackstage = bsv1alpha1.Backstage{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "bs",
		Namespace: "ns123",
	},
	Spec: bsv1alpha1.BackstageSpec{
		EnableLocalDb: pointer.Bool(false),
	},
}

func createBackstageTest(bs bsv1alpha1.Backstage) *testBackstageObject {
	b := &testBackstageObject{backstage: bs, detailedSpec: &DetailedBackstageSpec{BackstageSpec: bs.Spec}}
	b.detailedSpec.Details.RawConfig = map[string]string{}
	return b
}

func (b *testBackstageObject) withDefaultConfig(useDef bool) *testBackstageObject {
	if useDef {
		// here we have default-config folder
		_ = os.Setenv("LOCALBIN", "./testdata")
	} else {
		_ = os.Setenv("LOCALBIN", ".")
	}
	return b
}

func (b *testBackstageObject) addToDefaultConfig(key string, fileName string) *testBackstageObject {

	yaml, err := readTestYamlFile(fileName)
	if err != nil {
		panic(err)
	}
	b.detailedSpec.Details.RawConfig[key] = string(yaml)

	return b
}

func readTestYamlFile(name string) ([]byte, error) {

	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	return b, nil
}
