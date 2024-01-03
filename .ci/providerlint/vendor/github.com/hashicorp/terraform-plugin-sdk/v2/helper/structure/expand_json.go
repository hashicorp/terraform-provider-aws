// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package structure

import "encoding/json"

func ExpandJsonFromString(jsonString string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := json.Unmarshal([]byte(jsonString), &result)

	return result, err
}
