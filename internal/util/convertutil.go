package util

import (
	"encoding/json"

	"github.com/PaesslerAG/jsonpath"
	"github.com/ghodss/yaml"
)

func JsonPathQueryInYamlContent(content []byte, query string) (interface{}, error) {

	contentObj, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}

	v := interface{}(nil)
	json.Unmarshal(contentObj, &v)

	resultObj, err := jsonpath.Get(query, v)
	if err != nil {
		return nil, err
	}

	return resultObj, nil
}
