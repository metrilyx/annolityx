package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type EssDatastoreConfig struct {
	Host        string `toml:"host"`
	Port        int64  `toml:"port"`
	Index       string `toml:"index"`
	MappingFile string `toml:"mapping_file"`
}

type TypestoreConfig struct {
	DBFile string `toml:"dbfile"`
}

type HttpConfig struct {
	Port              int64  `toml:"port"`
	Webroot           string `toml:"webroot"`
	AnnoEndpoint      string `toml:"anno_endpoint"`
	TypesEndpoint     string `toml:"types_endpoint"`
	WebsocketEndpoint string `toml:"websocket_endpoint"`
	WebsocketHostname string `toml:"websocket_hostname"`
}

type PublisherConfig struct {
	Port int64  `toml:"port"`
	Type string `toml:"type"`
}

type Config struct {
	Typestore TypestoreConfig
	Datastore EssDatastoreConfig
	Http      HttpConfig
	Publisher PublisherConfig
}

func LoadConfigFromFile(filepath string) (*Config, error) {
	var config Config
	d, err := ioutil.ReadFile(filepath)
	if err != nil {
		return &config, err
	}
	_, err = toml.Decode(string(d), &config)
	if err != nil {
		return &config, err
	}
	if config.Http.Webroot[0] != '/' {
		currDir, err := os.Getwd()
		if err != nil {
			return &config, err
		}
		config.Http.Webroot = fmt.Sprintf("%s/%s", currDir, config.Http.Webroot)
	}
	if config.Typestore.DBFile[0] != '/' {
		currDir, err := os.Getwd()
		if err != nil {
			return &config, err
		}
		config.Typestore.DBFile = fmt.Sprintf("%s/%s", currDir, config.Typestore.DBFile)
	}
	return &config, nil
}
