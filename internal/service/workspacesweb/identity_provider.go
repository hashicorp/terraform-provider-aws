// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workspacesweb

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspacesweb_identity_provider", name="Identity Provider")
// @Tags(identifierAttribute="identity_provider_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.IdentityProvider")
// @Testing(importStateIdAttribute="identity_provider_arn")
func newIdentityProviderResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &identityProviderResource{}, nil
}

type identityProviderResource struct {
	framework.ResourceWithModel[identityProviderResourceModel]
}

func (r *identityProviderResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_provider_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identity_provider_details": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			"identity_provider_name": schema.StringAttribute{
				Required: true,
			},
			"identity_provider_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IdentityProviderType](),
				Required:   true,
			},
			"portal_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *identityProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data identityProviderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.IdentityProviderName)
	var input workspacesweb.CreateIdentityProviderInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIdentityProvider(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WorkSpacesWeb Identity Provider (%s)", name), err.Error())
		return
	}

	data.IdentityProviderARN = fwflex.StringToFramework(ctx, output.IdentityProviderArn)

	// Get the identity provider details to populate other fields
	identityProvider, portalARN, err := findIdentityProviderByARN(ctx, conn, data.IdentityProviderARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Identity Provider (%s)", data.IdentityProviderARN.ValueString()), err.Error())
		return
	}

	data.PortalARN = fwtypes.ARNValue(portalARN)

	response.Diagnostics.Append(fwflex.Flatten(ctx, identityProvider, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *identityProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data identityProviderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, portalARN, err := findIdentityProviderByARN(ctx, conn, data.IdentityProviderARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Identity Provider (%s)", data.IdentityProviderARN.ValueString()), err.Error())
		return
	}

	data.PortalARN = fwtypes.ARNValue(portalARN)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *identityProviderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old identityProviderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.IdentityProviderDetails.Equal(old.IdentityProviderDetails) ||
		!new.IdentityProviderName.Equal(old.IdentityProviderName) ||
		!new.IdentityProviderType.Equal(old.IdentityProviderType) {
		var input workspacesweb.UpdateIdentityProviderInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		output, err := conn.UpdateIdentityProvider(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Identity Provider (%s)", new.IdentityProviderARN.ValueString()), err.Error())
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, output.IdentityProvider, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *identityProviderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data identityProviderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteIdentityProviderInput{
		IdentityProviderArn: data.IdentityProviderARN.ValueStringPointer(),
	}
	_, err := conn.DeleteIdentityProvider(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Identity Provider (%s)", data.IdentityProviderARN.ValueString()), err.Error())
		return
	}
}

func (r *identityProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identity_provider_arn"), request, response)
}

const (
	arnResourceSeparator = "/"
	arnService           = "workspaces-web"
)

func portalARNFromIdentityProviderARN(identityProviderARN string) (string, error) {
	// Identity Provider ARN format: arn:{PARTITION}:workspaces-web:{REGION}:{ACCOUNT_ID}:identityProvider/{PORTAL_ID}/{IDP_RESOURCE_ID}
	// Portal ARN format: arn:{PARTITION}:workspaces-web:{REGION}:{ACCOUNT_ID}:portal/{PORTAL_ID}
	parsedARN, err := arn.Parse(identityProviderARN)

	if err != nil {
		return "", fmt.Errorf("parsing ARN (%s): %w", identityProviderARN, err)
	}

	if actual, expected := parsedARN.Service, arnService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, identityProviderARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, arnResourceSeparator)

	if actual, expected := len(resourceParts), 3; actual != expected {
		return "", fmt.Errorf("expected %d resource parts in ARN (%s), got: %d", expected, identityProviderARN, actual)
	}

	if actual, expected := resourceParts[0], "identityProvider"; actual != expected {
		return "", fmt.Errorf("expected %s in ARN (%s), got: %s", expected, identityProviderARN, actual)
	}

	portalARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  "portal" + arnResourceSeparator + resourceParts[1],
	}.String()

	return portalARN, nil
}

func findIdentityProviderByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.IdentityProvider, string, error) {
	input := workspacesweb.GetIdentityProviderInput{
		IdentityProviderArn: &arn,
	}
	output, err := conn.GetIdentityProvider(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, "", &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, "", err
	}

	if output == nil || output.IdentityProvider == nil {
		return nil, "", tfresource.NewEmptyResultError()
	}

	portalARN, err := portalARNFromIdentityProviderARN(arn)
	if err != nil {
		return nil, "", err
	}

	return output.IdentityProvider, portalARN, nil
}

type identityProviderResourceModel struct {
	framework.WithRegionModel
	IdentityProviderARN     types.String                                      `tfsdk:"identity_provider_arn"`
	IdentityProviderDetails fwtypes.MapOfString                               `tfsdk:"identity_provider_details"`
	IdentityProviderName    types.String                                      `tfsdk:"identity_provider_name"`
	IdentityProviderType    fwtypes.StringEnum[awstypes.IdentityProviderType] `tfsdk:"identity_provider_type"`
	PortalARN               fwtypes.ARN                                       `tfsdk:"portal_arn"`
	Tags                    tftags.Map                                        `tfsdk:"tags"`
	TagsAll                 tftags.Map                                        `tfsdk:"tags_all"`
}
