package config

import (
	"encoding/json"
	"testing"
)

var testConfigFile string = "/Users/abs/workbench/GoLang/src/github.com/euforia/annolityx/conf/annolityx.toml"

func Test_LoadConfigFromFile(t *testing.T) {
	testConfig, err := LoadConfigFromFile(testConfigFile)
	if err != nil {
		t.Errorf("FAILED: %s", err)
		t.FailNow()
	}
	b, err := json.MarshalIndent(&testConfig, "", "  ")
	t.Logf("%s\n", b)
}
