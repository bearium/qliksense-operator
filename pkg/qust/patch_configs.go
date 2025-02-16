package qust

import (
	"github.com/qlik-oss/qliksense-operator/pkg/config"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/types"
)

func ProcessCrConfigs(cr *config.CRConfig) {
	pm := createSupperConfigSelectivePatch(cr.Configs)
	baseConfigDir := filepath.Join(cr.ManifestsRoot, operatorPatchBaseFolder, "configs")
	for svc, sps := range pm {
		fpath := filepath.Join(baseConfigDir, svc+".yaml")
		fileHand, _ := os.Create(fpath)
		YamlToWriter(fileHand, sps)
		err := addResourceToKustomization(svc+".yaml", filepath.Join(baseConfigDir, "kustomization.yaml"))
		if err != nil {
			log.Println("Cannot process configs", err)
		}
	}
}

// create a selectivepatch map for each service for a dataKey
func createSupperConfigSelectivePatch(confg []config.Config) map[string]*config.SelectivePatch {
	spMap := make(map[string]*config.SelectivePatch)
	for _, conf := range confg {
		for svc, v := range conf.Values {
			p := getConfigMapPatchBody(conf.DataKey, svc, v)
			sp := getSuperConfigSPTemplate(svc)
			sp.Patches = []types.Patch{p}
			mergeSelectivePatches(sp, spMap[svc])
			spMap[svc] = sp

		}
	}
	return spMap
}

// create a patch section to be added to the selective patch
func getConfigMapPatchBody(dataKey, svc, value string) types.Patch {
	ph := getSuperConfigMapTemplate(svc)
	ph.Data = map[string]string{
		dataKey: value,
	}
	// ph := `
	// 	apiVersion: qlik.com/v1
	// 	kind: SuperConfigMap
	// 	metadata:
	// 		name: ` + svc + `-configs
	// 	data:
	// 		` + dataKey + `: ` + value

	// target:
	//   kind: SuperConfigMap
	//   labelSelector: "app=" + svc,
	phb, _ := yaml.Marshal(ph)
	p1 := types.Patch{
		Patch:  string(phb),
		Target: getSelector("SuperConfigMap", svc),
	}
	return p1
}

// a SelectivePatch object with service name in it
func getSuperConfigSPTemplate(svc string) *config.SelectivePatch {
	su := &config.SelectivePatch{
		ApiVersion: "qlik.com/v1",
		Kind:       "SelectivePatch",
		Metadata: map[string]string{
			"name": svc + "-operator-configs",
		},
		Enabled: true,
	}
	return su
}

func getSuperConfigMapTemplate(svc string) *config.SupperConfigMap {
	return &config.SupperConfigMap{
		ApiVersion: "qlik.com/v1",
		Kind:       "SuperConfigMap",
		Metadata: map[string]string{
			"name": svc + "-configs",
		},
	}
}
