package qust

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/types"
	"strings"
	"testing"
)

const tempPermissionCode os.FileMode = 0777

func setup() (func(), string) {
	dir, _ := ioutil.TempDir("", "testing_path")
	kustFile := `
kind: Kustomization
apiversion: kustomize.config.k8s.io/v1beta1
transformers:
- test-transformer.yaml
patches:
- path: test-patch.yaml
resources:
- mongodb-secret.yaml`
	kustf := filepath.Join(dir, "kustomization.yaml")
	ioutil.WriteFile(kustf, []byte(kustFile), tempPermissionCode)
	tearDown := func() {
		os.RemoveAll(dir)
	}
	return tearDown, dir
}

func setupCr(t *testing.T) io.Reader {
	t.Parallel()
	sampleConfig := `
configProfile: manifests/base
manifestsRoot: "."
configs:
- dataKey: acceptEULA
  values:
    qliksense: "yes"
secrets:
- secretKey: mongoDbUri
  values:
    qliksense: mongo://mongo:3307`
	os.Setenv("YAML_CONF", sampleConfig)
	return strings.NewReader(sampleConfig)
}

func createManifestsStructure(t *testing.T) (func(), string) {
	/*
		manifestsRoot
		|--.operator
		   |--configs
					|--kustomization.yaml
			 |--secrets
			    |--kustomization.yaml
	*/
	dir, _ := ioutil.TempDir("", "test_manifests")
	oprCnfDir := filepath.Join(dir, ".operator", "configs")
	oprSecDir := filepath.Join(dir, ".operator", "secrets")
	os.MkdirAll(oprCnfDir, tempPermissionCode)
	os.MkdirAll(oprSecDir, tempPermissionCode)
	k := `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:`

	err := ioutil.WriteFile(filepath.Join(oprCnfDir, "kustomization.yaml"), []byte(k), tempPermissionCode)
	if err != nil {
		t.Log(err)
		os.Exit(1)
	}
	ioutil.WriteFile(filepath.Join(oprSecDir, "kustomization.yaml"), []byte(k), tempPermissionCode)
	tearDown := func() {
		os.RemoveAll(dir)
	}
	return tearDown, dir
}

func TestAddResourceToKustomization(t *testing.T) {
	td, dir := setup()
	kustFile := filepath.Join(dir, "kustomization.yaml")
	addResourceToKustomization("test-file.yaml", kustFile)
	kust := &types.Kustomization{}
	content, err := ioutil.ReadFile(kustFile)
	if err != nil {
		t.FailNow()
	}
	yaml.Unmarshal(content, kust)
	if kust.Resources[1] != "test-file.yaml" {
		t.Fail()
	}
	addResourceToKustomization("test/test-file.yaml", kustFile)
	kust = &types.Kustomization{}
	content, err = ioutil.ReadFile(kustFile)
	if err != nil {
		t.FailNow()
	}
	yaml.Unmarshal(content, kust)
	if kust.Resources[2] != "test/test-file.yaml" {
		t.Fail()
	}
	td()
}

// func TestCreateSupperConfigSelectivePatch(t *testing.T) {
// 	reader := setupCr(t)
// 	cfg, err := config.ReadCRConfigFromFile(reader)
// 	if err != nil {
// 		t.Fatalf("error reading config from file")
// 	}
// 	spMap := createSupperConfigSelectivePatch(cfg.Configs)
// 	sp := spMap["qliksense"]
// 	if sp.ApiVersion != "qlik.com/v1" {
// 		t.Fail()
// 	}
// 	if sp.Kind != "SelectivePatch" {
// 		t.Fail()
// 	}
// 	if sp.Metadata["name"] != "qliksense-operator-configs" {
// 		t.Fail()
// 	}
// 	if sp.Patches[0].Target.LabelSelector != "app=qliksense" || sp.Patches[0].Target.Kind != "SuperConfigMap" {
// 		t.Fail()
// 	}
// 	scm := &config.SupperConfigMap{
// 		ApiVersion: "qlik.com/v1",
// 		Kind:       "SuperConfigMap",
// 		Metadata: map[string]string{
// 			"name": "qliksense-configs",
// 		},
// 		Data: map[string]string{
// 			"acceptEULA": "yes",
// 		},
// 	}
// 	scm2 := &config.SupperConfigMap{}
// 	yaml.Unmarshal([]byte(sp.Patches[0].Patch), scm2)
// 	if !reflect.DeepEqual(scm, scm2) {
// 		t.Fail()
// 	}
// }

// func TestProcessCrConfigs(t *testing.T) {
// 	reader := setupCr(t)
// 	cfg, err := config.ReadCRConfigFromFile(reader)
// 	if err != nil {
// 		t.Fatalf("error reading config from file")
// 	}

// 	td, dir := createManifestsStructure(t)

// 	cfg.ManifestsRoot = dir
// 	ProcessCrConfigs(cfg)
// 	content, _ := ioutil.ReadFile(filepath.Join(dir, ".operator", "configs", "qliksense.yaml"))

// 	sp := getSuperConfigSPTemplate("qliksense")
// 	scm := getSuperConfigMapTemplate("qliksense")
// 	scm.Data = map[string]string{
// 		"acceptEULA": "yes",
// 	}
// 	phb, _ := yaml.Marshal(scm)
// 	sp.Patches = []types.Patch{
// 		types.Patch{
// 			Patch:  string(phb),
// 			Target: getSelector("SuperConfigMap", "qliksense"),
// 		},
// 	}
// 	spOut := &config.SelectivePatch{}
// 	yaml.Unmarshal(content, spOut)
// 	if !reflect.DeepEqual(sp, spOut) {
// 		t.Fail()
// 	}

// 	td()
// }
