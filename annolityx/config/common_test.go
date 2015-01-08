package config

import (
	"encoding/json"
	"testing"
)

var testConfigFile string = fmt.Sprintf("%s/src/github.com/metrilyx/annolityx/conf/annolityx.toml", os.Getenv("GOPATH"))

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
