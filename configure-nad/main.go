package main

import (
	"context"
	"os"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

var _ fn.Runner = &ConfigureNad{}

type ConfigureNad struct {
        Configs map[string]string
} 

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
