// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkResource("aws_iam_outbound_web_identity_federation", name="Outbound Web Identity Federation")
func newResourceOutboundWebIdentityFederation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceOutboundWebIdentityFederation{}

	return r, nil
}

const (
	ResNameOutboundWebIdentityFederation = "Outbound Web Identity Federation"
)

type resourceOutboundWebIdentityFederation struct {
	framework.ResourceWithModel[resourceOutboundWebIdentityFederationModel]
	framework.WithNoUpdate
}

func (r *resourceOutboundWebIdentityFederation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"issuer_identifier": schema.StringAttribute{
				Computed: true,
			},
			"jwt_vending_enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceOutboundWebIdentityFederation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan resourceOutboundWebIdentityFederationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.EnableOutboundWebIdentityFederation(ctx, &iam.EnableOutboundWebIdentityFederationInput{})
	if errs.IsA[*awstypes.FeatureEnabledException](err) {
		// Feature is already enabled, adopt existing state
		outAlreadyEnabled, err := getOutboundWebIdentityFederation(ctx, conn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		if outAlreadyEnabled == nil {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("expected non-nil response from GetOutboundWebIdentityFederationInfo"))
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outAlreadyEnabled, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("expected non-nil response from EnableOutboundWebIdentityFederation"))
			return
		}
		plan.JwtVendingEnabled = types.BoolValue(true)
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceOutboundWebIdentityFederation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state resourceOutboundWebIdentityFederationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := getOutboundWebIdentityFederation(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceOutboundWebIdentityFederation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().IAMClient(ctx)

	var state resourceOutboundWebIdentityFederationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DisableOutboundWebIdentityFederation(ctx, &iam.DisableOutboundWebIdentityFederationInput{})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID)
		return
	}

}

type resourceOutboundWebIdentityFederationModel struct {
	JwtVendingEnabled types.Bool   `tfsdk:"jwt_vending_enabled"`
	IssuerIdentifier  types.String `tfsdk:"issuer_identifier"`
}

func getOutboundWebIdentityFederation(ctx context.Context, conn *iam.Client) (*iam.GetOutboundWebIdentityFederationInfoOutput, error) {
	out, err := conn.GetOutboundWebIdentityFederationInfo(ctx, &iam.GetOutboundWebIdentityFederationInfoInput{})
	if err != nil {
		if errs.IsA[*awstypes.FeatureDisabledException](err) {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}
