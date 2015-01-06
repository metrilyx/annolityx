package ess

import (
	"encoding/json"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"github.com/metrilyx/annolityx/annolityx/datastores"
	"testing"
	"time"
)

var testConfigFile string = "/Users/abs/workbench/GoLang/src/github.com/metrilyx/annolityx/conf/annolityx.toml"
var testTypestoreDbfile string = "/Users/abs/workbench/GoLang/src/github.com/metrilyx/annolityx/conf/annotation-types.json"

var testEssDatastore *ElasticsearchDatastore = &ElasticsearchDatastore{}
var testTypestore *datastores.JsonFileTypestore = &datastores.JsonFileTypestore{}

var testConfig *config.Config = &config.Config{}

var testType string = "Deployment"
var testAnnoMsg string = "Test deployment annotations"
var testAnnoData map[string]interface{} = map[string]interface{}{
	"host": "foo.bar.org", "datacenter": "dc1", "contact": "bar@foo.bar.org"}
var testTags map[string]string = map[string]string{"host": "foo", "dc": "dc1"}

var testStart float64 = 1418081663
var testEnd float64 = -1
var testAnnoQuery annotations.EventAnnotationQuery = annotations.EventAnnotationQuery{
	[]string{testType}, testTags, testStart, -1}

var testAnno annotations.EventAnnotation = annotations.EventAnnotation{
	Type:      testType,
	Message:   testAnnoMsg,
	Tags:      testTags,
	Data:      testAnnoData,
	Timestamp: float64(time.Now().UnixNano()) / 1000000000,
}

func Test_NewElasticsearchDatastore(t *testing.T) {
	testConfig, err := config.LoadConfigFromFile(testConfigFile)
	if err != nil {
		t.Fatalf("%s", err)
	}
	testEssDatastore, err = NewElasticsearchDatastore(testConfig)
	if err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("NewElasticsearchDatastore(%s, %d, %s)", testConfig.Datastore.Host, testConfig.Datastore.Port, testConfig.Datastore.Index)
}

func Test_ElasticsearchDatastore_Privates(t *testing.T) {

	etype := testEssDatastore.typeQuery(testType)
	if etype["term"]["type"] != "deployment" {
		t.Errorf("Event type mismatch: %s", etype)
	} else {
		t.Logf("typeQuery('%s')", testType)
	}

	tagsQ := testEssDatastore.tagsQuery(testTags)
	if len(tagsQ) != 2 {
		t.Errorf("Length mismatch: %s\n", tagsQ)
	} else {
		t.Logf("tagsQuery(%s)", testTags)
	}

	timeQ, err := testEssDatastore.timeQuery(testStart, testEnd)
	if err != nil {
		t.Errorf("%s\n", err)
		t.FailNow()
	}
	if timeQ["range"]["timestamp"]["gte"] != testStart {
		t.Errorf("Time mismatch: %v", timeQ["range"]["timestamp"])
	} else {
		t.Logf("timeQuery(%f,%f)", testStart, testEnd)
	}

	_, err = testEssDatastore.getQuery(testAnnoQuery, false)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	t.Logf("getQuery(%#v)", testAnnoQuery)
}

func Test_ElasticsearchDatastore_Annotate_Get(t *testing.T) {
	resp, err := testEssDatastore.Annotate(&testAnno)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	t.Logf("Annotate(%#v)", testAnno)
	j, _ := json.MarshalIndent(&resp, "", "  ")
	t.Logf("%s", j)

	respEvt, err := testEssDatastore.Get(resp.Type, resp.Id)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	t.Logf("%s", respEvt)
}

func Test_ElasticsearchDatastore_Query(t *testing.T) {

	resp, err := testEssDatastore.Query(testAnnoQuery, 0)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	if len(resp) < 1 {
		t.Errorf("No results returned!")
		t.FailNow()
	}
	t.Logf("Query(%#v)", testAnnoQuery)
	t.Logf("Result count: %d", len(resp))
}

func Test_ElasticsearchDatastore_ListTypes(t *testing.T) {
	testTypestore, err := datastores.NewJsonFileTypestore(testTypestoreDbfile)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	types := testTypestore.ListTypes()

	t.Logf("ListTypes() %d %#v\n", len(types), types)
}
