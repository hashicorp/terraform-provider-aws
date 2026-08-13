// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"strings"
)

// Args represents an argument list of the form:
// postional0, keywordA=valueA, positional1, keywordB=valueB
type Args struct {
	Positional []string
	Keyword    map[string]string
}

func ParseArgs(s string) (Args, error) {
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
		if err := checkBalancedQuotes(key); err != nil {
			return args, err
		}
		if err := checkBalancedQuotes(value); err != nil {
			return args, err
		}
		// Unquote.
		key = strings.Trim(key, `"`)
		value = strings.Trim(value, `"`)
		if value == "" {
			args.Positional = append(args.Positional, key)
		} else {
			args.Keyword[key] = value
		}
	}

	return args, nil
}

func checkBalancedQuotes(s string) error {
	count := strings.Count(s, `"`)
	if count%2 != 0 {
		return fmt.Errorf("unclosed quote in %q", s)
	}
	return nil
}
