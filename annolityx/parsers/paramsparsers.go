package parsers

import (
	"encoding/json"
	"fmt"
	"github.com/euforia/annolityx/annolityx/annotations"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
)

type AnnoQueryParamsParser struct {
	Params      url.Values
	RequestBody io.ReadCloser
}

func (p *AnnoQueryParamsParser) parseTagsParam() (map[string]string, error) {
	tags := make(map[string]string)
	if len(p.Params["tags"]) > 0 {
		kvpairs := strings.Split(p.Params["tags"][0], ",")
		for _, kvpair := range kvpairs {
			kv := strings.Split(kvpair, ":")
			if len(kv) != 2 {
				return tags, fmt.Errorf(`{"error": "Invalid tags: %s"}`, p.Params["tags"][0])
			} else if kv[0] == "" || kv[1] == "" {
				return tags, fmt.Errorf(`{"error": "Invalid tags: %s"}`, p.Params["tags"][0])
			}
			tags[kv[0]] = kv[1]
		}
	}
	return tags, nil
}

func (p *AnnoQueryParamsParser) parseTypesParam() []string {
	typesArr := make([]string, 0)
	if len(p.Params["types"]) > 0 {
		uTypes := strings.Split(p.Params["types"][0], ",")
		for _, ut := range uTypes {
			if ut != "" {
				typesArr = append(typesArr, ut)
			}
		}
	}
	return typesArr
}

func (p *AnnoQueryParamsParser) ParseTime() (float64, float64, error) {

	if len(p.Params["start"]) < 1 {
		return -1, -1, fmt.Errorf(`{"error":"Missing parameter: 'start'!"}`)
	}

	startTime, err := ParseTimeToEpoch(p.Params["start"][0])
	if err != nil {
		return -1, -1, fmt.Errorf(`{"error": "%s"}`, strings.Replace(err.Error(), `"`, "'", -1))
	}

	var endTime float64
	if len(p.Params["end"]) > 0 {
		endTime, err = ParseTimeToEpoch(p.Params["end"][0])
		if err != nil {
			return -1, -1, fmt.Errorf(`{"error": "%s"}`, strings.Replace(err.Error(), `"`, "'", -1))
		}
	} else {
		endTime = -1
	}
	return startTime, endTime, nil
}

func (p *AnnoQueryParamsParser) readGetBody(a annotations.EventAnnotationQuery) error {

	body, err := ioutil.ReadAll(p.RequestBody)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &a)
}

func (p *AnnoQueryParamsParser) ParseGetParams() (*annotations.EventAnnotationQuery, error) {

	var err error
	eaq := annotations.EventAnnotationQuery{}

	if err = p.readGetBody(eaq); err == nil {
		return &eaq, nil
	}

	if eaq.Start, eaq.End, err = p.ParseTime(); err != nil {
		return &eaq, err
	}

	if eaq.Tags, err = p.parseTagsParam(); err != nil {
		return &eaq, err
	}

	eaq.Types = p.parseTypesParam()

	return &eaq, nil
}
