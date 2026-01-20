// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

// Based on from https://cs.opensource.google/go/go/+/refs/tags/go1.23.0:src/encoding/json/tags.go

package flex

import (
	"strings"
)

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	tag, opt, _ := strings.Cut(tag, ",")
	return tag, tagOptions(opt)
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var name string
		name, s, _ = strings.Cut(s, ",")
		if name == optionName {
			return true
		}
	}
	return false
}

func (o tagOptions) Legacy() bool {
	return o.Contains("legacy")
}

func (o tagOptions) OmitEmpty() bool {
	return o.Contains("omitempty")
}

func (o tagOptions) NoFlatten() bool {
	return o.Contains("noflatten")
}

func (o tagOptions) XMLWrapperField() string {
	if len(o) == 0 {
		return ""
	}
	s := string(o)
	for s != "" {
		var option string
		option, s, _ = strings.Cut(s, ",")
		if name, value, found := strings.Cut(option, "="); found && name == "xmlwrapper" {
			return value
		}
	}
	return ""
}
