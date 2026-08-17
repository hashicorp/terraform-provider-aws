// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceHostedPrivateVirtualInterfaceAccepterSchema(t *testing.T) {
	t.Parallel()

	resource := resourceHostedPrivateVirtualInterfaceAccepter()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("invalid resource schema: %s", err)
	}

	attribute, ok := resource.SchemaMap()["sitelink_enabled"]
	if !ok {
		t.Fatal("expected sitelink_enabled schema")
	}
	if attribute.Type != schema.TypeBool {
		t.Errorf("expected sitelink_enabled to be TypeBool, got %s", attribute.Type)
	}
	if !attribute.Optional {
		t.Error("expected sitelink_enabled to be optional")
	}
	if !attribute.Computed {
		t.Error("expected sitelink_enabled to be computed")
	}
	if resource.Timeouts.Update == nil || *resource.Timeouts.Update != 10*time.Minute {
		t.Errorf("expected update timeout to be 10m, got %v", resource.Timeouts.Update)
	}
}

func TestValidateHostedPrivateVirtualInterfaceAccepterSiteLink(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		siteLinkEnabled cty.Value
		vpnGatewayID    cty.Value
		wantErr         bool
	}{
		"disabled with virtual private gateway": {
			siteLinkEnabled: cty.BoolVal(false),
			vpnGatewayID:    cty.StringVal("vgw-12345678"),
		},
		"enabled with direct connect gateway": {
			siteLinkEnabled: cty.BoolVal(true),
			vpnGatewayID:    cty.NullVal(cty.String),
		},
		"enabled with virtual private gateway": {
			siteLinkEnabled: cty.BoolVal(true),
			vpnGatewayID:    cty.StringVal("vgw-12345678"),
			wantErr:         true,
		},
		"unknown SiteLink value": {
			siteLinkEnabled: cty.UnknownVal(cty.Bool),
			vpnGatewayID:    cty.StringVal("vgw-12345678"),
		},
		"enabled with unknown virtual private gateway": {
			siteLinkEnabled: cty.BoolVal(true),
			vpnGatewayID:    cty.UnknownVal(cty.String),
			wantErr:         true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := validateHostedPrivateVirtualInterfaceAccepterSiteLink(testCase.siteLinkEnabled, testCase.vpnGatewayID)
			if gotErr := err != nil; gotErr != testCase.wantErr {
				t.Errorf("expected error: %t, got: %v", testCase.wantErr, err)
			}
		})
	}
}
