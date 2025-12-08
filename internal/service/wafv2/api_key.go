// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_wafv2_api_key", name="API Key")
func newAPIKeyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &apiKeyResource{}, nil
}

type apiKeyResource struct {
	framework.ResourceWithModel[apiKeyResourceModel]
	framework.WithNoUpdate
}

func (r *apiKeyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a WAFv2 API Key resource.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The API key value. This is sensitive and not included in responses.",
			},
			names.AttrScope: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Scope](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are CLOUDFRONT or REGIONAL.",
			},
			"token_domains": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
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
							// Based on RFC 1034, RFC 1123, and RFC 5890
							// - Domain labels must start and end with alphanumeric characters
							// - Domain labels can contain hyphens but not at start or end
							// - Each domain label can be up to 63 characters
							// - Must contain at least one period (separating domain and TLD)
							regexache.MustCompile(`^([0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?\.)+[0-9A-Za-z]([0-9A-Za-z-]{0,61}[0-9A-Za-z])?$`),
							"domain names must follow DNS format with valid characters: A-Z, a-z, 0-9, -(hyphen) or . (period)",
						),
						stringvalidator.LengthAtMost(253),
					),
				},
				Description: "The domains that you want to be able to use the API key with, for example example.com. Maximum of 5 domains.",
			},
		},
	}
}

func (r *apiKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data apiKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := wafv2.CreateAPIKeyInput{
		Scope:        data.Scope.ValueEnum(),
		TokenDomains: fwflex.ExpandFrameworkStringValueSet(ctx, data.TokenDomains),
	}

	output, err := conn.CreateAPIKey(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WAFv2 API Key", err.Error())

		return
	}

	// Set values for unknowns.
	data.APIKey = fwflex.StringToFramework(ctx, output.APIKey)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *apiKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data apiKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	output, err := findAPIKeyByTwoPartKey(ctx, conn, data.APIKey.ValueString(), data.Scope.ValueEnum())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading WAFv2 API Key", err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data apiKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := wafv2.DeleteAPIKeyInput{
		APIKey: fwflex.StringFromFramework(ctx, data.APIKey),
		Scope:  data.Scope.ValueEnum(),
	}

	_, err := conn.DeleteAPIKey(ctx, &input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("deleting WAFv2 API Key", err.Error())

		return
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		apiKeyIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, apiKeyIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("api_key"), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrScope), parts[1])...)
}

func findAPIKeyByTwoPartKey(ctx context.Context, conn *wafv2.Client, key string, scope awstypes.Scope) (*awstypes.APIKeySummary, error) {
	input := wafv2.ListAPIKeysInput{
		Scope: scope,
	}

	return findAPIKey(ctx, conn, &input, func(v *awstypes.APIKeySummary) bool {
		return aws.ToString(v.APIKey) == key
	})
}

func findAPIKey(ctx context.Context, conn *wafv2.Client, input *wafv2.ListAPIKeysInput, filter tfslices.Predicate[*awstypes.APIKeySummary]) (*awstypes.APIKeySummary, error) {
	output, err := findAPIKeys(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAPIKeys(ctx context.Context, conn *wafv2.Client, input *wafv2.ListAPIKeysInput, filter tfslices.Predicate[*awstypes.APIKeySummary]) ([]awstypes.APIKeySummary, error) {
	var output []awstypes.APIKeySummary

	err := listAPIKeysPages(ctx, conn, input, func(page *wafv2.ListAPIKeysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.APIKeySummaries {
			if filter(&v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	return output, err
}

type apiKeyResourceModel struct {
	framework.WithRegionModel
	APIKey       types.String                       `tfsdk:"api_key"`
	Scope        fwtypes.StringEnum[awstypes.Scope] `tfsdk:"scope"`
	TokenDomains fwtypes.SetOfString                `tfsdk:"token_domains"`
}
