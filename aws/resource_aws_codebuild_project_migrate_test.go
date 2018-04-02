package aws

import (
	"reflect"
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
				"description": "some description",
				"timeout":     "5",
			},
			Expected: map[string]string{
				"description":   "some description",
				"timeout":       "5",
				"build_timeout": "5",
			},
		},
		"v0_2": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"description":   "some description",
				"build_timeout": "5",
			},
			Expected: map[string]string{
				"description":   "some description",
				"build_timeout": "5",
			},
		},
		"v1_1": {
			StateVersion: 1,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"description":                                "some description",
				"source.1060593600.auth.2706882902.type":     "OAUTH",
				"source.1060593600.type":                     "GITHUB",
				"source.1060593600.buildspec":                "",
				"source.#":                                   "1",
				"source.1060593600.location":                 "https://github.com/hashicorp/packer.git",
				"source.1060593600.auth.2706882902.resource": "FAKERESOURCE1",
				"source.1060593600.auth.#":                   "1",
			},
			// added defaults
			Expected: map[string]string{
				"description":                                "some description",
				"source.#":                                   "1",
				"source.3680370172.type":                     "GITHUB",
				"source.3680370172.buildspec":                "",
				"source.3680370172.location":                 "https://github.com/hashicorp/packer.git",
				"source.3680370172.git_clone_depth":          "0",
				"source.3680370172.insecure_ssl":             "false",
				"source.3680370172.auth.#":                   "1",
				"source.3680370172.auth.2706882902.resource": "FAKERESOURCE1",
				"source.3680370172.auth.2706882902.type":     "OAUTH",
			},
		},
		"v1_noop": {
			StateVersion: 1,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"description":                                "some description",
				"source.#":                                   "1",
				"source.3680370172.type":                     "GITHUB",
				"source.3680370172.buildspec":                "",
				"source.3680370172.location":                 "https://github.com/hashicorp/packer.git",
				"source.3680370172.git_clone_depth":          "0",
				"source.3680370172.insecure_ssl":             "false",
				"source.3680370172.auth.#":                   "1",
				"source.3680370172.auth.2706882902.resource": "FAKERESOURCE1",
				"source.3680370172.auth.2706882902.type":     "OAUTH",
			},
			Expected: map[string]string{
				"description":                                "some description",
				"source.#":                                   "1",
				"source.3680370172.type":                     "GITHUB",
				"source.3680370172.buildspec":                "",
				"source.3680370172.location":                 "https://github.com/hashicorp/packer.git",
				"source.3680370172.git_clone_depth":          "0",
				"source.3680370172.insecure_ssl":             "false",
				"source.3680370172.auth.#":                   "1",
				"source.3680370172.auth.2706882902.resource": "FAKERESOURCE1",
				"source.3680370172.auth.2706882902.type":     "OAUTH",
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

		if !reflect.DeepEqual(is.Attributes, tc.Expected) {
			t.Fatalf("Bad migration: %s\n\n expected: %s", is.Attributes, tc.Expected)
		}
	}
}
