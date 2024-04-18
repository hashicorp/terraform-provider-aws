// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"strings"
)

// Args represents an argument list of the form:
// postional0, keywordA=valueA, positional1, keywordB=valueB
type Args struct {
	Positional []string
	Keyword    map[string]string
}

func ParseArgs(s string) Args {
	args := Args{
		Keyword: make(map[string]string),
	}
	var key string

	for s != "" {
		key, s, _ = strings.Cut(s, ",")
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")
		// Unquote.
		key = strings.Trim(key, `"`)
		value = strings.Trim(value, `"`)
		if value == "" {
			args.Positional = append(args.Positional, key)
		} else {
			args.Keyword[key] = value
		}
	}

	return args
}
