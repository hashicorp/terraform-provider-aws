// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"errors"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_wafv2_api_key", name="API Key")
func newResourceAPIKey(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAPIKey{}, nil
}

const (
	ResNameAPIKey = "API Key"
	// Based on RFC 1034, RFC 1123, and RFC 5890
	// - Domain labels must start and end with alphanumeric characters
	// - Domain labels can contain hyphens but not at start or end
	// - Each domain label can be up to 63 characters
	// - Must contain at least one period (separating domain and TLD)
	RegexDNSName  = `^([0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?\.)+[0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?$`
	apiKeyIDParts = 2
)

type resourceAPIKey struct {
	framework.ResourceWithConfigure
}

func (r *resourceAPIKey) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a WAFv2 API Key resource.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The API key value. This is sensitive and not included in responses.",
			},
			"token_domains": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(5),
					setvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexache.MustCompile(RegexDNSName),
							"domain names must follow DNS format with valid characters: A-Z, a-z, 0-9, -(hyphen) or . (period)",
						),
						stringvalidator.LengthAtMost(253),
					),
				},
				Description: "The domains that you want to be able to use the API key with, for example example.com. Maximum of 5 domains.",
			},
			names.AttrScope: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{string(awstypes.ScopeCloudfront), string(awstypes.ScopeRegional)}...),
				},
				Description: "Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are CLOUDFRONT or REGIONAL.",
			},
		},
	}
}

func (r *resourceAPIKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAPIKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := wafv2.CreateAPIKeyInput{
		Scope:        awstypes.Scope(plan.Scope.ValueString()),
		TokenDomains: flex.ExpandFrameworkStringValueSet(ctx, plan.TokenDomains),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("APIKey"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateAPIKey(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameAPIKey, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.APIKey == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionCreating, ResNameAPIKey, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.APIKey = types.StringValue(*out.APIKey)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAPIKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var state resourceAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAPIKeyByKey(ctx, conn, state.APIKey.ValueString(), state.Scope.ValueString())
	if errs.IsA[*tfresource.EmptyResultError](err) || tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionReading, ResNameAPIKey, state.APIKey.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAPIKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// WAFv2 APIKey cannot be updated - any change requires replacement so we reuse existing data
	var state resourceAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAPIKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WAFV2Client(ctx)

	var state resourceAPIKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := wafv2.DeleteAPIKeyInput{
		APIKey: state.APIKey.ValueStringPointer(),
		Scope:  awstypes.Scope(state.Scope.ValueString()),
	}

	_, err := conn.DeleteAPIKey(ctx, &input)
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WAFV2, create.ErrActionDeleting, ResNameAPIKey, state.APIKey.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAPIKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, apiKeyIDParts, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: api_key,scope. Valid scope values are CLOUDFRONT or REGIONAL. Got: %q", req.ID),
		)
		return
	}

	scope := parts[1]
	if scope != string(awstypes.ScopeCloudfront) && scope != string(awstypes.ScopeRegional) {
		resp.Diagnostics.AddError(
			"Invalid Scope Value",
			fmt.Sprintf("Expected scope to be one of %q or %q. Got: %q",
				string(awstypes.ScopeCloudfront), string(awstypes.ScopeRegional), scope),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("api_key"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrScope), parts[1])...)
}

func findAPIKeyByKey(ctx context.Context, conn *wafv2.Client, key, scope string) (*awstypes.APIKeySummary, error) {
	input := &wafv2.ListAPIKeysInput{
		Scope: awstypes.Scope(scope),
	}

	for {
		output, err := conn.ListAPIKeys(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("listing API Keys: %w", err)
		}

		for _, apiKey := range output.APIKeySummaries {
			if aws.ToString(apiKey.APIKey) == key {
				return &apiKey, nil
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	return nil, &tfresource.EmptyResultError{
		LastRequest: input,
	}
}

type resourceAPIKeyModel struct {
	APIKey       types.String `tfsdk:"api_key"`
	Scope        types.String `tfsdk:"scope"`
	TokenDomains types.Set    `tfsdk:"token_domains"`
}
