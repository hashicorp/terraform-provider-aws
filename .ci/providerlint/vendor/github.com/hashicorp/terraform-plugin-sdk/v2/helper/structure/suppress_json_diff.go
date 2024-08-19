// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package structure

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func SuppressJsonDiff(k, oldValue, newValue string, d *schema.ResourceData) bool {
	oldMap, err := ExpandJsonFromString(oldValue)
	if err != nil {
		return false
	}

	newMap, err := ExpandJsonFromString(newValue)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(oldMap, newMap)
}
