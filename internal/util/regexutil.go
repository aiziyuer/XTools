package util

import (
	"github.com/dlclark/regexp2"
	"go.uber.org/zap"
	"regexp"
)

func NamedStringSubMatch(r *regexp.Regexp, text string) map[string]string {

	result := map[string]string{}

	match := r.FindStringSubmatch(text)
	if match == nil {
		return result
	}

	for i, name := range r.SubexpNames() {
		if i != 0 {
			result[name] = match[i]
		}
	}

	return result
}

func NamedStringAllMatch(pattern string, text string) []map[string]string {

	resultList := make([]map[string]string, 0)
	r := regexp2.MustCompile(pattern, regexp2.RE2)

	m, err := r.FindStringMatch(text)
	zap.S().Debug(err)
	for m != nil {
		tmpResult := map[string]string{}
		resultList = append(resultList, tmpResult)
		for _, group := range m.Groups() {
			tmpResult[group.Name] = group.Capture.String()
		}
		m, _ = r.FindNextMatch(m)
	}

	return resultList
}
