// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestKeyPairMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"fingerprint":       "1d:cd:46:31:a9:4a:e0:06:8a:a1:22:cb:3b:bf:8e:42",
				"key_name":          "tf-testing-file",
				names.AttrPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA4LBtwcFsQAYWw1cnOwRTZCJCzPSzq0dl3== comment", // nosemgrep:ci.ssh-key
			},
			Expected: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA4LBtwcFsQAYWw1cnOwRTZCJCzPSzq0dl3== comment", // nosemgrep:ci.ssh-key
		},
		"v0_2": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"fingerprint":       "1d:cd:46:31:a9:4a:e0:06:8a:a1:22:cb:3b:bf:8e:42",
				"key_name":          "tf-testing-file",
				names.AttrPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA4LBtwcFsQAYWw1cnOwRTZCJCzPSzq0dl3== comment\n", // nosemgrep:ci.ssh-key
			},
			Expected: "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA4LBtwcFsQAYWw1cnOwRTZCJCzPSzq0dl3== comment", // nosemgrep:ci.ssh-key
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := tfec2.KeyPairMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.Attributes[names.AttrPublicKey] != tc.Expected {
			t.Fatalf("Bad public_key migration: %s\n\n expected: %s", is.Attributes[names.AttrPublicKey], tc.Expected)
		}
	}
}
