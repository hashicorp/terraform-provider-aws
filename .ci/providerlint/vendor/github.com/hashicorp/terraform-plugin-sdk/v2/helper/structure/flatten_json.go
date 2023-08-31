// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package structure

import "encoding/json"

func FlattenJsonToString(input map[string]interface{}) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	result, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
