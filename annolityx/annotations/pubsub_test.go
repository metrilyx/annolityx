package annotations

import (
	"fmt"
	"github.com/metrilyx/annolityx/annolityx/config"
	"testing"
	"time"
)

var testConfigFile string = "/Users/abs/workbench/GoLang/src/github.com/metrilyx/annolityx/conf/annolityx.toml"
var testCfg *config.Config = &config.Config{}
var testSubConnUri string = "tcp://localhost:%d"

func Test_LoadConfigFromFile(t *testing.T) {
	var err error
	testCfg, err = config.LoadConfigFromFile(testConfigFile)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
}

func Test_NewEventAnno_PublisherSubscriber(t *testing.T) {

	sub, err := NewEventAnnoSubscriber(fmt.Sprintf(testSubConnUri, testCfg.Publisher.Port), "SUB", []string{""})
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	go func() {
		resp, err := sub.Recieve()
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		t.Logf("%#v", resp)
	}()
	//t.Logf("%#v", sub)
	listenAddr := fmt.Sprintf("tcp://*:%d", testCfg.Publisher.Port)
	pub, err := NewEventAnnoPublisher(listenAddr, "PUB")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	//t.Logf("%#v", pub)

	err = pub.Publish("", `{"name": "test"}`)
	time.Sleep(2)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

}

func Test_SubscriptionHash(t *testing.T) {
	subHash, err := SubscriptionHash(
		[]string{"deployment", "alarm"},
		map[string]string{
			"host": "foo.bar.org",
			"dc":   "dc0",
		})
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	t.Logf("%s", subHash)
}

func Test_SubscriptionHash_NoTypes(t *testing.T) {
	subHash, err := SubscriptionHash(make([]string, 0),
		map[string]string{
			"host": "foo.bar.org",
			"dc":   "dc0",
		})
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	if subHash != "{dc:dc0,host:foo.bar.org}" {
		t.Errorf("invalid hash: %s", subHash)
		t.FailNow()
	}
	t.Logf("%s", subHash)
}

func Test_SubscriptionHash_NoTags(t *testing.T) {
	subHash, err := SubscriptionHash(
		[]string{"deployment", "alarm"},
		make(map[string]string))

	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if subHash != "alarm,deployment{}" {
		t.Errorf("invalid hash: %s", subHash)
		t.FailNow()
	}
	t.Logf("%s", subHash)
}

func Test_SubscriptionHash_NoTypesTags(t *testing.T) {
	subHash, err := SubscriptionHash(make([]string, 0), make(map[string]string))
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if subHash != "" {
		t.Errorf("invalid hash: %s\n", subHash)
	}
}
