// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workspacesweb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspacesweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspacesweb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// @FrameworkResource("aws_workspacesweb_ip_access_settings", name="IP Access Settings")
// @Tags(identifierAttribute="ip_access_settings_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.IpAccessSettings")
// @Testing(importStateIdAttribute="ip_access_settings_arn")
func newIPAccessSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &ipAccessSettingsResource{}, nil
}

type ipAccessSettingsResource struct {
	framework.ResourceWithModel[ipAccessSettingsResourceModel]
}

func (r *ipAccessSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_encryption_context": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"associated_portal_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_managed_key": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"ip_access_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"ip_rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(100),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"ip_range": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *ipAccessSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ipAccessSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.DisplayName)
	var input workspacesweb.CreateIpAccessSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIpAccessSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WorkSpacesWeb IP Access Settings (%s)", name), err.Error())
		return
	}

	data.IPAccessSettingsARN = fwflex.StringToFramework(ctx, output.IpAccessSettingsArn)

	// Get the IP access settings details to populate other fields
	ipAccessSettings, err := findIPAccessSettingsByARN(ctx, conn, data.IPAccessSettingsARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb IP Access Settings (%s)", data.IPAccessSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, ipAccessSettings, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *ipAccessSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ipAccessSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findIPAccessSettingsByARN(ctx, conn, data.IPAccessSettingsARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb IP Access Settings (%s)", data.IPAccessSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ipAccessSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old ipAccessSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.DisplayName.Equal(old.DisplayName) ||
		!new.IPRules.Equal(old.IPRules) {
		var input workspacesweb.UpdateIpAccessSettingsInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := conn.UpdateIpAccessSettings(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb IP Access Settings (%s)", new.IPAccessSettingsARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *ipAccessSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ipAccessSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteIpAccessSettingsInput{
		IpAccessSettingsArn: fwflex.StringFromFramework(ctx, data.IPAccessSettingsARN),
	}
	_, err := conn.DeleteIpAccessSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb IP Access Settings (%s)", data.IPAccessSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *ipAccessSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ip_access_settings_arn"), request, response)
}

func findIPAccessSettingsByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.IpAccessSettings, error) {
	input := workspacesweb.GetIpAccessSettingsInput{
		IpAccessSettingsArn: &arn,
	}
	output, err := conn.GetIpAccessSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IpAccessSettings == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.IpAccessSettings, nil
}

type ipAccessSettingsResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext fwtypes.MapOfString                          `tfsdk:"additional_encryption_context"`
	AssociatedPortalARNs        fwtypes.ListOfString                         `tfsdk:"associated_portal_arns"`
	CustomerManagedKey          fwtypes.ARN                                  `tfsdk:"customer_managed_key"`
	Description                 types.String                                 `tfsdk:"description"`
	DisplayName                 types.String                                 `tfsdk:"display_name"`
	IPAccessSettingsARN         types.String                                 `tfsdk:"ip_access_settings_arn"`
	IPRules                     fwtypes.ListNestedObjectValueOf[ipRuleModel] `tfsdk:"ip_rule"`
	Tags                        tftags.Map                                   `tfsdk:"tags"`
	TagsAll                     tftags.Map                                   `tfsdk:"tags_all"`
}

type ipRuleModel struct {
	Description types.String `tfsdk:"description"`
	IPRange     types.String `tfsdk:"ip_range"`
}
