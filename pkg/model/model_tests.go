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
)

func setTestEnv(useDefMandatoryObjects ...bool) {

	useDef := true
	if len(useDefMandatoryObjects) > 0 {
		useDef = useDefMandatoryObjects[0]
	}

	if useDef {
		_ = os.Setenv("LOCALBIN", "./testdata")
	} else {
		_ = os.Setenv("LOCALBIN", ".")
	}

}

func readTestYamlFile(name string) ([]byte, error) {

	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}
	return b, nil
}
