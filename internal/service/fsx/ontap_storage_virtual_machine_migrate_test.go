package fsx_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
)

func TestOntapStorageVirtualMachineMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1-notDistinguidshed": {
			StateVersion: 0,
			Attributes: map[string]string{
				"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name": "",
			},
			Expected: map[string]string{
				"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name": "",
			},
		},
		"v0_1-bitDistinguidshed": {
			StateVersion: 0,
			Attributes: map[string]string{
				"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name": "MeArrugoDerrito",
			},
			Expected: map[string]string{
				"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name": "MeArrugoDerrito",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		is, err := tffsx.ResourceOntapStorageVirtualMachine().MigrateState(tc.StateVersion, is, tc.Meta)
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
