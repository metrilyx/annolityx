package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

var testConfigFile, _ = filepath.Abs("../../etc/annolityx/annolityx.toml")

func Test_LoadConfigFromFile(t *testing.T) {
	testConfig, err := LoadConfigFromFile(testConfigFile)
	if err != nil {
		t.Errorf("FAILED: %s", err)
		t.FailNow()
	}
	t.Logf("Config loaded: %s", testConfigFile)
	b, err := json.MarshalIndent(&testConfig, "", "  ")
	t.Logf("%s\n", b)
}
