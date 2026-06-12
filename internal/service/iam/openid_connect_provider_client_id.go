// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_iam_openid_connect_provider_client_id", name="Open ID Connect Provider Client ID")
// @IdentityAttribute("openid_connect_provider_arn")
// @IdentityAttribute("client_id")
// @ImportIDHandler("openIDConnectProviderClientIDImportID")
// @Testing(hasNoPreExistingResource=true)
func newOpenIDConnectProviderClientIDResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &openIDConnectProviderClientIDResource{}, nil
}

const (
	ResNameOpenIDConnectProviderClientID = "Open ID Connect Provider Client ID"
)

type openIDConnectProviderClientIDResource struct {
	framework.ResourceWithModel[openIDConnectProviderClientIDResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *openIDConnectProviderClientIDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"openid_connect_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrClientID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *openIDConnectProviderClientIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().IAMClient(ctx)

	var plan openIDConnectProviderClientIDResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := iam.AddClientIDToOpenIDConnectProviderInput{
		ClientID:                 plan.ClientID.ValueStringPointer(),
		OpenIDConnectProviderArn: plan.OpenIDConnectProviderARN.ValueStringPointer(),
	}

	_, err := conn.AddClientIDToOpenIDConnectProvider(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			"adding client ID to IAM OIDC Provider",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *openIDConnectProviderClientIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state openIDConnectProviderClientIDResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := findOpenIDConnectProviderClientID(ctx, conn, state.OpenIDConnectProviderARN.ValueString(), state.ClientID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"reading IAM OIDC Provider client ID",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *openIDConnectProviderClientIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IAMClient(ctx)

	var state openIDConnectProviderClientIDResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := iam.RemoveClientIDFromOpenIDConnectProviderInput{
		ClientID:                 state.ClientID.ValueStringPointer(),
		OpenIDConnectProviderArn: state.OpenIDConnectProviderARN.ValueStringPointer(),
	}

	_, err := conn.RemoveClientIDFromOpenIDConnectProvider(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return
		}

		resp.Diagnostics.AddError(
			"removing client ID from IAM OIDC Provider",
			err.Error(),
		)
		return
	}
}

// findOpenIDConnectProviderClientID checks whether a client ID exists in the
// list associated with the given OIDC provider ARN. It returns a NotFoundError
// if the provider itself is not found or if the client ID is not in the list.
func findOpenIDConnectProviderClientID(ctx context.Context, conn *iam.Client, providerARN, clientID string) error {
	output, err := findOpenIDConnectProviderByARN(ctx, conn, providerARN)
	if err != nil {
		return err
	}

	if !slices.Contains(output.ClientIDList, clientID) {
		return tfresource.NewEmptyResultError()
	}

	return nil
}

type openIDConnectProviderClientIDResourceModel struct {
	ClientID                 types.String `tfsdk:"client_id"`
	OpenIDConnectProviderARN fwtypes.ARN  `tfsdk:"openid_connect_provider_arn"`
}

var (
	_ inttypes.ImportIDParser = openIDConnectProviderClientIDImportID{}
)

type openIDConnectProviderClientIDImportID struct{}

func (openIDConnectProviderClientIDImportID) Parse(id string) (string, map[string]any, error) {
	providerARN, clientID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("import ID %q should be in the format <provider-arn>%s<client-id>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"openid_connect_provider_arn": providerARN,
		names.AttrClientID:            clientID,
	}

	return id, result, nil
}
