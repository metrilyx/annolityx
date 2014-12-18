package datastores

import (
	"encoding/json"
	"fmt"
	"github.com/metrilyx/annolityx/annolityx/annotations"
	"io/ioutil"
)

type JsonFileTypestore struct {
	dbfile    string
	AnnoTypes map[string]annotations.EventAnnoType
}

func NewJsonFileTypestore(dbfile string) (*JsonFileTypestore, error) {
	jfts := JsonFileTypestore{}
	jfts.dbfile = dbfile

	b, err := ioutil.ReadFile(dbfile)
	if err != nil {
		return &jfts, err
	}

	if err = json.Unmarshal(b, &jfts.AnnoTypes); err != nil {
		return &jfts, err
	}
	return &jfts, nil
}

func (e *JsonFileTypestore) GetType(id string) (annotations.EventAnnoType, error) {
	if val, ok := e.AnnoTypes[id]; ok {
		return val, nil
	}
	return annotations.EventAnnoType{}, fmt.Errorf("Type not found: %s", id)
}

func (e *JsonFileTypestore) UpsertType(eat annotations.EventAnnoType) error {
	e.AnnoTypes[eat.Id] = eat
	if err := e.writeToFile(); err != nil {
		return err
	}
	return nil
}
func (e *JsonFileTypestore) RemoveType(id string) error {
	if _, ok := e.AnnoTypes[id]; ok {
		delete(e.AnnoTypes, id)
		if err := e.writeToFile(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Type not found: %s", id)
	}
	return nil
}
func (e *JsonFileTypestore) ListTypes() []annotations.EventAnnoType {
	out := make([]annotations.EventAnnoType, len(e.AnnoTypes))
	i := 0
	for _, v := range e.AnnoTypes {
		out[i] = v
		i++
	}
	return out
}

func (e *JsonFileTypestore) writeToFile() error {
	if b, err := json.MarshalIndent(&e.AnnoTypes, "", "  "); err == nil {
		if err = ioutil.WriteFile(e.dbfile, b, 0777); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}
