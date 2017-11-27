package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAWSCodebuildMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"artifacts.#":                        "1",
				"artifacts.537882814.location":       "tf-codebuild-93892171401399123",
				"artifacts.537882814.name":           "packaging_test",
				"artifacts.537882814.namespace_type": "NONE",
				"artifacts.537882814.packaging":      "ZIP",
				"artifacts.537882814.path":           "",
				"artifacts.537882814.type":           "S3",
				"build_timeout":                      "60",
				"source.#":                           "1",
				"source.3360844398.auth.#":           "0",
				"source.3360844398.buildspec":        "buildspec_build.yml",
				"source.3360844398.location":         "fugafuga",
				"source.3360844398.type":             "CODECOMMIT",
			},
			Expected: map[string]string{
				"artifacts.#":                         "1",
				"artifacts.2191994982.location":       "tf-codebuild-93892171401399123",
				"artifacts.2191994982.name":           "packaging_test",
				"artifacts.2191994982.namespace_type": "NONE",
				"artifacts.2191994982.packaging":      "ZIP",
				"artifacts.2191994982.path":           "",
				"artifacts.2191994982.type":           "S3",
				"build_timeout":                       "60",
				"source.#":                            "1",
				"source.3360844398.auth.#":            "0",
				"source.3360844398.buildspec":         "buildspec_build.yml",
				"source.3360844398.location":          "fugafuga",
				"source.3360844398.type":              "CODECOMMIT",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsCodebuildMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}
