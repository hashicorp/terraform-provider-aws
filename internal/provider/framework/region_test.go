// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/framework/resourceattribute"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceSetRegionInStateInterceptor_Read(t *testing.T) {
	t.Parallel()

	const name = "example"

	region := "a_region"

	ctx := context.Background()
	client := mockClient{region: region}
	icpt := resourceSetRegionInStateInterceptor{}

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName:   schema.StringAttribute{Required: true},
			names.AttrRegion: resourceattribute.Region(),
		},
	}

	tests := map[string]struct {
		startState tfsdk.State
		expectSet  bool
	}{
		"when state is present then region is set": {
			startState: stateFromSchema(ctx, s, map[string]string{"name": name}),
			expectSet:  true,
		},
		"when state is null then it remains null": {
			startState: tfsdk.State{
				Raw:    tftypes.NewValue(s.Type().TerraformType(ctx), nil),
				Schema: s,
			},
			expectSet: false,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			req := resource.ReadRequest{State: tc.startState}
			resp := resource.ReadResponse{State: tc.startState}

			icpt.read(ctx, interceptorOptions[resource.ReadRequest, resource.ReadResponse]{
				c:        client,
				request:  &req,
				response: &resp,
				when:     After,
			})
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diags: %s", resp.Diagnostics)
			}

			if tc.expectSet {
				got := getStateAttributeValue(ctx, t, resp.State, path.Root("region"))
				if got != region {
					t.Errorf("expected region %q, got %q", region, got)
				}
			} else {
				if !resp.State.Raw.IsNull() {
					t.Errorf("expected State.Raw to stay null, got %#v", resp.State.Raw)
				}
			}
		})
	}
}

func getStateAttributeValue(ctx context.Context, t *testing.T, st tfsdk.State, p path.Path) string {
	t.Helper()

	var v types.String
	if diags := st.GetAttribute(ctx, p, &v); diags.HasError() {
		t.Fatalf("unexpected error getting State attribute %q: %s", p, fwdiag.DiagnosticsError(diags))
	}
	return v.ValueString()
}
