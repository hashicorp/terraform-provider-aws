// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidMetricName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"testrule",
		"testRule",
		"testRule123",
	}
	for _, v := range validNames {
		_, errors := validMetricName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid WAF metric name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"!",
		"/",
		" ",
		":",
		";",
		"white space",
		"/slash-at-the-beginning",
		"slash-at-the-end/",
	}
	for _, v := range invalidNames {
		_, errors := validMetricName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid WAF metric name", v)
		}
	}
}
