// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

var _ fn.Runner = &ConfigureNad{}

// TODO: Change to your functionConfig "Kind" name.
type ConfigureNad struct {
        Configs map[string]string
} 

// Run is the main function logic.
// `items` is parsed from the STDIN "ResourceList.Items".
// `functionConfig` is from the STDIN "ResourceList.FunctionConfig". The value has been assigned to the r attributes
// `results` is the "ResourceList.Results" that you can write result info to.
func (cn *ConfigureNad) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
        if _, ok := cn.Configs["config"]; !ok {
            *results = append(*results, fn.GeneralResult("config is missing!", fn.Error))

            return false
        }

        for _, kubeObject := range items {
            if kubeObject.IsGVK("k8s.cni.cncf.io", "v1", "NetworkAttachmentDefinition") {
                if nadResource, ok := cn.Configs["resourceName"]; ok {
                    kubeObject.SetAnnotation("k8s.v1.cni.cncf.io/resourceName", nadResource)
                }

                if err := kubeObject.SetNestedString(cn.Configs["config"], "spec", "config"); err != nil {
                    *results = append(*results, fn.ErrorResult(err))

                    return false
                }
            }
         }

        *results = append(*results, fn.GeneralResult("Successfully configured NADs", fn.Info))

        return true
}

func main() {
	runner := fn.WithContext(context.Background(), &ConfigureNad{})
	if err := fn.AsMain(runner); err != nil {
		os.Exit(1)
	}
}
