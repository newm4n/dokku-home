package configuration

import (
	"encoding/json"
	"fmt"
)

func ConfigurationFromBytes(source []byte) (*ConfigurationFile, error) {
	confFile := &ConfigurationFile{}
	err := json.Unmarshal(source, &confFile)
	if err != nil {
		return nil, err
	}
	if confFile.Version != "1.0" {
		return nil, fmt.Errorf("only accept version 1.0. got %s", confFile.Version)
	}
	if len(confFile.EncPoints) < 1 {
		return nil, fmt.Errorf("this configuration require minimum 1 endpoint")
	}
	for idx, cnf := range confFile.EncPoints {
		if len(cnf.URLPathPrefix) == 0 || len(cnf.PathPrefix) == 0 {
			return nil, fmt.Errorf("on config %d, only accept valid URL Prefix and Path Prefix. got %s and %s", idx, cnf.URLPathPrefix, cnf.PathPrefix)
		}
	}
	return confFile, nil
}

type ConfigurationFile struct {
	Version   string
	EncPoints []*BackendEnd
}

type BackendEnd struct {
	PathPrefix    string
	URLHost       string
	URLPathPrefix string
}
