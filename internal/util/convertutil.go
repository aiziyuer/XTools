package util

import (
	"encoding/json"

	"github.com/PaesslerAG/jsonpath"
	"github.com/ghodss/yaml"
	"github.com/gogf/gf/os/gfile"
	"go.uber.org/zap"
)

func YamlFileToObject(yamlFilePath string) (interface{}, error) {
	return YamlContentToObject(gfile.GetBytes(yamlFilePath))
}

func YamlContentToObject(content []byte) (interface{}, error) {
	contentObj, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}

	inputObject := interface{}(nil)
	json.Unmarshal(contentObj, &inputObject)

	return inputObject, nil
}

func JsonPathQueryInYamlContent(content []byte, query string) (interface{}, error) {

	contentObj, err := yaml.YAMLToJSON(content)
	if err != nil {
		return nil, err
	}

	inputObject := interface{}(nil)
	json.Unmarshal(contentObj, &inputObject)

	return JsonPathQueryInObject(inputObject, query)
}

func JsonPathQueryInObject(inputObject interface{}, query string) (interface{}, error) {

	resultObj, err := jsonpath.Get(query, inputObject)
	if err != nil {
		return nil, err
	}

	return resultObj, nil
}

func MustJsonPathQueryInObject(inputObject interface{}, query string) interface{} {

	resultObj, err := JsonPathQueryInObject(inputObject, query)
	if err != nil {
		zap.S().Fatal(err)
	}

	return resultObj

}
