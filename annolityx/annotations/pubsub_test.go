package annotations

import (
	"fmt"
	"github.com/metrilyx/annolityx/annolityx/config"
	"path/filepath"
	//"sync"
	"testing"
	"time"
)

var testConfigFile, _ = filepath.Abs("../../etc/annolityx/annolityx.toml")

var testCfg *config.Config = &config.Config{}

var (
	testSrvPort    = 34343
	testListenAddr = fmt.Sprintf("tcp://*:%d", testSrvPort)
	testSubConnUri = fmt.Sprintf("tcp://127.0.0.1:%d", testSrvPort)
)

func Test_LoadConfigFromFile(t *testing.T) {
	var err error
	testCfg, err = config.LoadConfigFromFile(testConfigFile)
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

func Test_NewEventAnno_PublisherSubscriber(t *testing.T) {
	//

	testCfg, _ = config.LoadConfigFromFile(testConfigFile)

	pub, err := NewEventAnnoPublisher(testListenAddr, testCfg.Publisher.Type)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	go func() {
		sub, err := NewEventAnnoSubscriber(fmt.Sprintf(testSubConnUri, testCfg.Publisher.Port), "SUB", []string{""})
		defer sub.Close()
		if err != nil {
			t.Errorf("%s", err)
			t.FailNow()
		}
		t.Logf("subscriber done")
		resp, err := sub.Receive()
		if err != nil {
			t.Fatalf("%s", err)
		}
		t.Logf("Received: %#v", resp)
	}()

	time.Sleep(3)
	if err = pub.Publish("", `{"name": "test"}`); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	t.Logf("Published!")
	t.Logf("Waiting...")
}
