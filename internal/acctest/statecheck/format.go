// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func populateFromResourceState(format string, state *tfjson.StateResource) (string, error) {
	remaining := format

	var buf strings.Builder
	for remaining != "" {
		var (
			stuff string
			found bool
		)
		stuff, remaining, found = strings.Cut(remaining, "{")
		buf.WriteString(stuff)
		if found {
			var param string
			param, remaining, found = strings.Cut(remaining, "}")
			if !found {
				return "", fmt.Errorf("missing closing '}' in format %q", format)
			}

			attr, ok := state.AttributeValues[param]
			if !ok {
				return "", fmt.Errorf("attribute %q not found in resource %q, referenced in format %q", param, state.Address, format)
			}
			fmt.Fprintf(&buf, "%v", attr)
		}
	}

	return buf.String(), nil
}
