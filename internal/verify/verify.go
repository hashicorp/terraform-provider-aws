package verify

import (
	"gopkg.in/yaml.v2"
)

const UUIDRegexPattern = `[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[ab89][a-f0-9]{3}-[a-f0-9]{12}`

func SliceContainsString(slice []interface{}, s string) (int, bool) {
	for idx, value := range slice {
		v := value.(string)
		if v == s {
			return idx, true
		}
	}
	return -1, false
}

// Takes a value containing YAML string and passes it through
// the YAML parser. Returns either a parsing
// error or original YAML string.
func checkYAMLString(yamlString interface{}) (string, error) {
	var y interface{}

	if yamlString == nil || yamlString.(string) == "" {
		return "", nil
	}

	s := yamlString.(string)

	err := yaml.Unmarshal([]byte(s), &y)

	return s, err
}
