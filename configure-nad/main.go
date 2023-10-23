package main

import (
	"encoding/json"
	"fmt"
	"os"
  	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/xeipuuv/gojsonschema"
)

var schemaValidators *schemaValidator

type dataContainer interface {}

type schemaValidator struct {
	schemas map[string]*gojsonschema.Schema
}

var defaultIdentifierAnnotation string = "configure-nad"

func CreateNad(rl *fn.ResourceList) (bool, error) {
        if err := InitSchemaValidator(filepath.Join(scriptDir(), "cni-schemas")); err != nil {
               rl.Results = append(rl.Results, fn.ErrorResult(err))

               return false, err
        }

        identifierAnnotation, exists, _ := rl.FunctionConfig.NestedString("data", "identifierAnnotation")
        if !exists {
            identifierAnnotation = defaultIdentifierAnnotation
        }

        resourceName, exists, _ := rl.FunctionConfig.NestedString("data", "resourceName")
        if !exists {
            resourceName = ""
        }

        for _, kubeObject := range rl.Items {
                if !kubeObject.IsGVK("", "v1", "ConfigMap") {

                        continue
                }

                if !kubeObject.HasAnnotations(map[string]string{identifierAnnotation: ""}){
                        continue
                }

                success, cniType, CNIConfigs, nadDir := loadConfigMapConfigs(kubeObject, &rl.Results)
                if !success {
                        return false, nil
                }

                if !validateCNIConfigs(cniType, CNIConfigs, &rl.Results) {
                        return false, nil
                }

                nadKubeObject, err := createNADKubeObject(kubeObject.GetName(), CNIConfigs, resourceName, nadDir)
                if err != nil {
                        rl.Results = append(rl.Results, fn.GeneralResult(fmt.Sprintf("failed to create NAD %v", err), fn.Error))

                        return false, err
                }

		        if err := rl.UpsertObjectToItems(nadKubeObject, checkExistence, true); err != nil {
		                return false, err
		        }
        }

        rl.Results = append(rl.Results, fn.GeneralResult("Successfully created all NADs", fn.Info))

	return true, nil
}

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(CreateNad)); err != nil {
		os.Exit(1)
	}
}

func loadConfigMapConfigs(kubeObject *fn.KubeObject, results *fn.Results) (bool, string, string, string) {
        nadDir := filepath.Dir(kubeObject.PathAnnotation())

        var ConfigMapData dataContainer

        err := kubeObject.As(&ConfigMapData)
        if err != nil {
                *results = append(*results, fn.GeneralResult("Error loading configMap", fn.Error))
                *results = append(*results, fn.ErrorResult(err))

                return false, "", "", ""
        }

        cniData, ok := ConfigMapData.(map[string]interface{})["data"]
        if !ok {
                *results = append(*results, fn.GeneralResult("ConfigMap is missing the data", fn.Error))

                return false, "", "", ""
        }

        cniType, ok := cniData.(map[string]interface{})["type"]
        if !ok {
                *results = append(*results, fn.GeneralResult("ConfigMap is missing the type", fn.Error))

                return false, "", "", ""
        }

        jsonData, err := json.Marshal(cniData)
        if err != nil {
                *results = append(*results, fn.ErrorResult(err))

                return false, "", "", ""
        }

        cniConfigs := string(jsonData)

        return true, cniType.(string), cniConfigs, nadDir
}

func validateCNIConfigs(cniType string, cniConfigs string, results *fn.Results) bool {
    sriovCNISchema, err := schemaValidators.GetSchema(cniType)
    if err != nil {
        *results = append(*results, fn.ErrorResult(err))

        return false
    }

    configJSONLoader := gojsonschema.NewStringLoader(cniConfigs)

    result, err := sriovCNISchema.Validate(configJSONLoader)
    if err != nil {
        *results = append(*results, fn.GeneralResult("Error while validating the CNI config", fn.Error))
        *results = append(*results, fn.ErrorResult(err))

        return false

    } else if !result.Valid() {
        *results = append(*results, fn.GeneralResult("Error in CNI config schema!", fn.Error))

        for _, ResultErr := range result.Errors() {
            *results = append(*results, fn.GeneralResult(ResultErr.String(), fn.Error))
        }

        return false
    }

    return true
}

func createNADKubeObject(nadName string, nadConfig string, resourceName string, nadDir string) (*fn.KubeObject, error) {
        nadPath := filepath.Join(nadDir, nadName + ".yaml")

	nadKubeobject := fn.NewEmptyKubeObject()

        if err := nadKubeobject.SetAPIVersion("k8s.cni.cncf.io/v1"); err != nil {
                return fn.NewEmptyKubeObject(), err
        }

        if err := nadKubeobject.SetKind("NetworkAttachmentDefinition"); err != nil {
                return fn.NewEmptyKubeObject(), err
        }

        if err := nadKubeobject.SetName(nadName); err != nil {
                return fn.NewEmptyKubeObject(), err
        }

        if err := nadKubeobject.SetNestedString(nadConfig, "spec", "config"); err != nil {
                return fn.NewEmptyKubeObject(), err
	}

	if err := nadKubeobject.SetAnnotation("internal.config.kubernetes.io/path", nadPath); err != nil {
		return fn.NewEmptyKubeObject(), err
	}

        if resourceName != "" {
                if err := nadKubeobject.SetAnnotation("k8s.v1.cni.cncf.io/resourceName", resourceName); err != nil {
                        return fn.NewEmptyKubeObject(), err
                }
        }

        return nadKubeobject, nil
}

func (sv *schemaValidator) GetSchema(schemaName string) (*gojsonschema.Schema, error) {
	s, ok := sv.schemas[schemaName]
	if !ok {
		return nil, fmt.Errorf("validation schema not found: %s", schemaName)
	}
	return s, nil
}

func InitSchemaValidator(schemaPath string) error {
	sv := &schemaValidator{
		schemas: make(map[string]*gojsonschema.Schema),
	}

	files, err := os.ReadDir(schemaPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		s, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s/%s", schemaPath, f.Name())))
		if err != nil {
		    return err
		}

		sv.schemas[strings.TrimSuffix(f.Name(), ".json")] = s
	}

	schemaValidators = sv

	return nil
}

func scriptDir() string {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(path)
}

func checkExistence(obj, another *fn.KubeObject) bool {
    return obj.GetKind() == another.GetKind() && obj.GetName() == another.GetName()
}

