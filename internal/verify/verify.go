// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	tfyaml "github.com/hashicorp/terraform-provider-aws/internal/yaml"
)

const UUIDRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[ab89][0-9a-f]{3}-[0-9a-f]{12}`

// Takes a value containing YAML string and passes it through
// the YAML parser. Returns either a parsing
// error or original YAML string.
func checkYAMLString(yamlString any) (string, error) {
	if yamlString == nil || yamlString.(string) == "" {
		return "", nil
	}

	var y any

	s := yamlString.(string)

	err := tfyaml.DecodeFromString(s, &y)

	return s, err
}
