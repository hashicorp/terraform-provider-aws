package wafregional

import (
	"testing"
)

func TestValidMetricName(t *testing.T) {
	validNames := []string{
		"testrule",
		"testRule",
		"testRule123",
	}
	for _, v := range validNames {
		_, errors := validMetricName(v, "name")
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
		_, errors := validMetricName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid WAF metric name", v)
		}
	}
}
