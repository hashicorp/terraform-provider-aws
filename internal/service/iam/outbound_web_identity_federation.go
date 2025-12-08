// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_iam_outbound_web_identity_federation", name="Outbound Web Identity Federation")
// @SingletonIdentity
// @WrappedImport(false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(generator=false)
func newOutboundWebIdentityFederationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &outboundWebIdentityFederationResource{}

	return r, nil
}

type outboundWebIdentityFederationResource struct {
	framework.ResourceWithModel[outboundWebIdentityFederationResourceModel]
	framework.WithNoUpdate
}

func (r *outboundWebIdentityFederationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"issuer_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *outboundWebIdentityFederationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data outboundWebIdentityFederationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	var input iam.EnableOutboundWebIdentityFederationInput
	out, err := conn.EnableOutboundWebIdentityFederation(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("enabling IAM Outbound Web Identity Federation", err.Error())
		return
	}

	// Set values for unknowns.
	data.IssuerIdentifier = fwflex.StringToFramework(ctx, out.IssuerIdentifier)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *outboundWebIdentityFederationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data outboundWebIdentityFederationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IAMClient(ctx)

	out, err := findOutboundWebIdentityFederation(ctx, conn)
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading IAM Outbound Web Identity Federation", err.Error())
		return
	}

	// Set attributes for import.
	data.IssuerIdentifier = fwflex.StringToFramework(ctx, out.IssuerIdentifier)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *outboundWebIdentityFederationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IAMClient(ctx)

	var input iam.DisableOutboundWebIdentityFederationInput
	_, err := conn.DisableOutboundWebIdentityFederation(ctx, &input)
	if errs.IsA[*awstypes.FeatureDisabledException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("disabling IAM Outbound Web Identity Federation", err.Error())
		return
	}
}

func (r *outboundWebIdentityFederationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// Something must be set in state to prevent "Missing Resource Import State" errors.
	resource.ImportStatePassthroughID(ctx, path.Root("issuer_identifier"), request, response)
}

func findOutboundWebIdentityFederation(ctx context.Context, conn *iam.Client) (*iam.GetOutboundWebIdentityFederationInfoOutput, error) {
	var input iam.GetOutboundWebIdentityFederationInfoInput
	out, err := conn.GetOutboundWebIdentityFederationInfo(ctx, &input)

	if errs.IsA[*awstypes.FeatureDisabledException](err) {
		return nil, &sdkretry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type outboundWebIdentityFederationResourceModel struct {
	IssuerIdentifier types.String `tfsdk:"issuer_identifier"`
}
