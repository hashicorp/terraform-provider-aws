// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
)

func TestBGPAuthKeySchemasSensitive(t *testing.T) {
	t.Parallel()

	testCases := map[string]func() *schema.Resource{
		"aws_dx_bgp_peer":                         tfdirectconnect.ResourceBGPPeer,
		"aws_dx_private_virtual_interface":        tfdirectconnect.ResourcePrivateVirtualInterface,
		"aws_dx_public_virtual_interface":         tfdirectconnect.ResourcePublicVirtualInterface,
		"aws_dx_transit_virtual_interface":        tfdirectconnect.ResourceTransitVirtualInterface,
		"aws_dx_hosted_private_virtual_interface": tfdirectconnect.ResourceHostedPrivateVirtualInterface,
		"aws_dx_hosted_public_virtual_interface":  tfdirectconnect.ResourceHostedPublicVirtualInterface,
		"aws_dx_hosted_transit_virtual_interface": tfdirectconnect.ResourceHostedTransitVirtualInterface,
	}

	for name, factory := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resource := factory()

			if err := resource.InternalValidate(nil, true); err != nil {
				t.Fatalf("invalid resource schema: %s", err)
			}

			attribute, ok := resource.SchemaFunc()["bgp_auth_key"]
			if !ok {
				t.Fatal("bgp_auth_key schema is missing")
			}

			if attribute.Type != schema.TypeString {
				t.Errorf("Type = %v, want %v", attribute.Type, schema.TypeString)
			}
			if !attribute.Optional {
				t.Error("Optional = false, want true")
			}
			if !attribute.Computed {
				t.Error("Computed = false, want true")
			}
			if !attribute.ForceNew {
				t.Error("ForceNew = false, want true")
			}
			if !attribute.Sensitive {
				t.Error("Sensitive = false, want true")
			}
		})
	}
}
