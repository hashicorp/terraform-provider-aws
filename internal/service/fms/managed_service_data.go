// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// suppressEquivalentManagedServiceDataJSON provides custom difference suppression
// for strings that are equivalent once empty fields have been removed.
func suppressEquivalentManagedServiceDataJSON(k, old, new string, d *schema.ResourceData) bool {
	if !json.Valid([]byte(old)) || !json.Valid([]byte(new)) {
		return old == new
	}

	old, new = removeEmptyFieldsFromJSON(old), removeEmptyFieldsFromJSON(new)

	return verify.JSONStringsEqual(old, new)
}

// removeEmptyFieldsFromJSON removes null and empty array values from a valid JSON string.
func removeEmptyFieldsFromJSON(in string) string {
	out, err := tfjson.RemoveJSONValues(in, tfjson.RemoveOptions{
		Null:        true,
		EmptyArrays: true,
	})
	if err != nil {
		return ""
	}

	return out
}
