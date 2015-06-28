package datastores

import (
	"encoding/json"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"path/filepath"
	"testing"
	"time"
)

/* test data */
var (
	testConfigFile, _  = filepath.Abs("../../etc/annolityx/annolityx.toml")
	testTypesDbFile, _ = filepath.Abs("../../etc/annolityx/annotation-types.json")
	testMappingFile, _ = filepath.Abs("../../etc/annolityx/eventannotations-mapping.json")

	testEssDatastore = &ElasticsearchDatastore{}
	testTypestore    = &JsonFileTypestore{}
	testConfig       = &config.Config{}
)

/* mock data */
var (
	testType     = "deployment"
	testAnnoMsg  = "Test deployment annotations"
	testAnnoData = map[string]interface{}{
		"host":       "foo.bar.org",
		"datacenter": "dc1",
		"contact":    "bar@foo.bar.org",
	}
	testQueryTags = map[string]string{
		"class": "met",
		"dc":    "dc1|dc2",
	}
	testTags = map[string]string{"dc": "dc1", "class": "met"}

	testCartesianTags = map[string][]string{
		"dc":    []string{"dc1", "dc2"},
		"class": []string{"app", "met"},
	}

	testStart float64 = 1418081663
	testEnd   float64 = -1

	testAnnoQuery = annotations.EventAnnotationQuery{
		Types: []string{testType},
		Tags:  testQueryTags,
		Start: testStart,
		End:   testEnd,
	}

	testAnno = annotations.EventAnnotation{
		Type:      testType,
		Message:   testAnnoMsg,
		Tags:      testTags,
		Data:      testAnnoData,
		Timestamp: float64(time.Now().UnixNano()) / 1000000000,
	}
)

func Test_NewElasticsearchDatastore(t *testing.T) {
	testConfig, err := config.LoadConfigFromFile(testConfigFile)
	if err != nil {
		t.Fatalf("%s", err)
	}

	testConfig.Datastore.MappingFile = testMappingFile

	if testEssDatastore, err = NewElasticsearchDatastore(testConfig); err != nil {
		t.Fatalf("%s", err)
	}
	t.Logf("NewElasticsearchDatastore(%s, %d, %s)", testConfig.Datastore.Host,
		testConfig.Datastore.Port, testConfig.Datastore.Index)
}

func Test_ElasticsearchDatastore_Privates(t *testing.T) {

	etype := testEssDatastore.typeQuery(testType)
	if etype["term"]["type"] != "deployment" {
		t.Errorf("Event type mismatch: %s", etype)
	} else {
		t.Logf("typeQuery('%s')", testType)
	}

	tagsQ := testEssDatastore.tagsQuery(testTags)
	//if len(tagsQ) != 2 {
	//	t.Errorf("Length mismatch: %s\n", tagsQ)
	//} else {
	t.Logf("tagsQuery(%s)", tagsQ)
	//}

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

func Test_ElasticsearchDatastore_tagsCartesianProduct(t *testing.T) {
	tagsArr := testEssDatastore.tagsCartesianProduct(testCartesianTags)
	if len(tagsArr) != 4 {
		t.Errorf("Tags mismatch: %d", len(tagsArr))
		t.FailNow()
	}
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
	t.Logf("%#v", resp[0])
}

func Test_ElasticsearchDatastore_ListTypes(t *testing.T) {
	testTypestore, err := NewJsonFileTypestore(testTypesDbFile)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	types := testTypestore.ListTypes()

	t.Logf("ListTypes() %d %#v\n", len(types), types)
}
