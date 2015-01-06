package ess

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	elastigo "github.com/mattbaird/elastigo/lib"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"github.com/metrilyx/annolityx/annolityx/config"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const ESS_DEFAULT_RESULT_SIZE int = 10000000

type EssMapping struct {
	Meta             map[string]interface{} `json:"_meta"`
	DynamicTemplates []interface{}          `json:"dynamic_templates"`
}

type ElasticsearchDatastore struct {
	conn  *elastigo.Conn
	index string
}

func NewElasticsearchDatastore(cfg *config.Config) (*ElasticsearchDatastore, error) {

	c := elastigo.NewConn()
	c.Domain = cfg.Datastore.Host
	c.Port = fmt.Sprintf("%d", cfg.Datastore.Port)

	ed := ElasticsearchDatastore{}
	ed.conn = c
	ed.index = cfg.Datastore.Index

	exists, err := c.ExistsIndex(cfg.Datastore.Index, "", nil)
	if err != nil {
		if err.Error() == "record not found" {
			exists = false
		} else {
			return &ed, err
		}
	}

	if !exists {
		return &ed, ed.initializeIndex(cfg.Datastore.MappingFile)
	}
	return &ed, nil
}

func (e *ElasticsearchDatastore) initializeIndex(mappingFile string) error {
	resp, err := e.conn.CreateIndex(e.index)
	if err != nil {
		return err
	}
	fmt.Printf("Index created: %s %s\n", e.index, resp)

	if _, err := os.Stat(mappingFile); err != nil {
		return fmt.Errorf("Mapping file not found %s: %s", mappingFile, err)
	}

	mappingDataBytes, err := ioutil.ReadFile(mappingFile)
	if err != nil {
		return err
	}
	b, err := e.conn.DoCommand("PUT", fmt.Sprintf("/%s/_mapping/_default_", e.index), nil, mappingDataBytes)
	if err != nil {
		return err
	}
	fmt.Printf("Updated _default_ mapping for %s: %s\n", e.index, b)
	return nil
}

func (e *ElasticsearchDatastore) Get(etype, id string) (*annotations.EventAnnotation, error) {
	var evt annotations.EventAnnotation

	resp, err := e.conn.Get(e.index, etype, id, nil)
	if err != nil {
		return &evt, err
	}

	if err = json.Unmarshal(*resp.Source, &evt); err != nil {
		return &evt, err
	}
	return &evt, nil
}

func (e *ElasticsearchDatastore) generateId(anno *annotations.EventAnnotation) (string, error) {
	if anno.Id == "" {
		b, err := json.Marshal(&anno)
		if err != nil {
			return "", err
		}
		anno.Id = fmt.Sprintf("%x", sha1.Sum(b))
	}
	return anno.Id, nil
}

// type IEventAnnotation interface //
func (e *ElasticsearchDatastore) Annotate(anno *annotations.EventAnnotation) (*annotations.EventAnnotation, error) {

	anno.PostedTimestamp = float64(time.Now().UnixNano()) / 1000000000

	id, err := e.generateId(anno)
	if err != nil {
		return anno, err
	}

	essResp, err := e.conn.Index(e.index, anno.Type, id, nil, anno)
	e.conn.Flush()
	if err != nil {
		return anno, err
	}
	if !essResp.Created {
		return anno, fmt.Errorf("Failed to annotate: %s", essResp)
	}
	return anno, nil
}

// type IEventAnnotation interface //
func (e *ElasticsearchDatastore) Query(annoQuery annotations.EventAnnotationQuery, limit int64) ([]*annotations.EventAnnotation, error) {
	// array //
	essQuery, err := e.getQuery(annoQuery, false)
	if err != nil {
		return nil, err
	}

	var opts map[string]interface{}
	if limit < 1 {
		opts = map[string]interface{}{
			"from": 0,
			"size": ESS_DEFAULT_RESULT_SIZE,
			"sort": "timestamp"}
	} else {
		opts = map[string]interface{}{
			"from": 0,
			"size": int(limit),
			"sort": "timestamp"}
	}

	resp, err := e.conn.Search(e.index, "", opts, essQuery[0]) // temporary until looped //
	if err != nil {
		return nil, err
	}

	out := make([]*annotations.EventAnnotation, len(resp.Hits.Hits))
	for i, hit := range resp.Hits.Hits {
		var ea annotations.EventAnnotation
		err := json.Unmarshal(*hit.Source, &ea)
		if err != nil {
			return out, err
		}
		out[i] = &ea
	}

	return out, nil
}

func (e *ElasticsearchDatastore) getQuery(q annotations.EventAnnotationQuery, splitByType bool) ([]interface{}, error) {
	timeQ, err := e.timeQuery(q.Start, q.End)
	if err != nil {
		return make([]interface{}, 0), err
	}

	tagsQ := e.tagsQuery(q.Tags)

	andQueries := make([]interface{}, 1+len(tagsQ))
	andQueries[0] = timeQ

	for i, v := range tagsQ {
		andQueries[i+1] = v
	}

	typeQueries := make([]map[string]map[string]string, len(q.Types))
	for i, t := range q.Types {
		typeQueries[i] = e.typeQuery(t)
	}

	var queries []interface{}
	if splitByType {
		queries = make([]interface{}, len(typeQueries))
		for i, typeq := range typeQueries {
			queries[i] = map[string]interface{}{
				"query": map[string]map[string]map[string]map[string]interface{}{
					"filtered": {
						"filter": {
							"bool": {
								"must": append(andQueries, typeq),
							},
						},
					},
				},
			}
		}
	} else {
		queries = make([]interface{}, 1)
		queries[0] = map[string]interface{}{
			"query": map[string]map[string]map[string]map[string]interface{}{
				"filtered": {
					"filter": {
						"bool": {
							"must":   andQueries,
							"should": typeQueries,
						},
					},
				},
			},
		}
	}

	return queries, nil
}

func (e *ElasticsearchDatastore) timeQuery(start, end float64) (map[string]map[string]map[string]float64, error) {
	var out map[string]map[string]map[string]float64

	if end < start && end > 0 {
		return out, fmt.Errorf("end > start: %f %f", start, end)
	} else if end <= 0 {
		out = map[string]map[string]map[string]float64{"range": {
			"timestamp": {"gte": start}}}
	} else {
		out = map[string]map[string]map[string]float64{"range": {
			"timestamp": {"gte": start, "lte": end}}}
	}
	return out, nil
}

func (e *ElasticsearchDatastore) typeQuery(eType string) map[string]map[string]string {
	etype := strings.ToLower(eType)
	return map[string]map[string]string{"term": {"type": etype}}
}

func (e *ElasticsearchDatastore) tagsQuery(tagMap map[string]string) []map[string]map[string]string {
	tQuery := make([]map[string]map[string]string, len(tagMap))
	i := 0
	for k, v := range tagMap {
		tQuery[i] = map[string]map[string]string{"term": {fmt.Sprintf("tags.%s", k): v}}
		i++
	}
	return tQuery
}
