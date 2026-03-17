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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

// @FrameworkResource("aws_workspacesweb_user_access_logging_settings", name="User Access Logging Settings")
// @Tags(identifierAttribute="user_access_logging_settings_arn")
// @Testing(tagsTest=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.UserAccessLoggingSettings")
// @Testing(importStateIdAttribute="user_access_logging_settings_arn")
func newUserAccessLoggingSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userAccessLoggingSettingsResource{}, nil
}

type userAccessLoggingSettingsResource struct {
	framework.ResourceWithModel[userAccessLoggingSettingsResourceModel]
}

func (r *userAccessLoggingSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"associated_portal_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"kinesis_stream_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"user_access_logging_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userAccessLoggingSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userAccessLoggingSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	var input workspacesweb.CreateUserAccessLoggingSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateUserAccessLoggingSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb User Access Logging Settings", err.Error())
		return
	}

	data.UserAccessLoggingSettingsARN = fwflex.StringToFramework(ctx, output.UserAccessLoggingSettingsArn)

	// Get the user access logging settings details to populate other fields
	userAccessLoggingSettings, err := findUserAccessLoggingSettingsByARN(ctx, conn, data.UserAccessLoggingSettingsARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb User Access Logging Settings (%s)", data.UserAccessLoggingSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, userAccessLoggingSettings, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userAccessLoggingSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userAccessLoggingSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findUserAccessLoggingSettingsByARN(ctx, conn, data.UserAccessLoggingSettingsARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb User Access Logging Settings (%s)", data.UserAccessLoggingSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userAccessLoggingSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old userAccessLoggingSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.KinesisStreamARN.Equal(old.KinesisStreamARN) {
		var input workspacesweb.UpdateUserAccessLoggingSettingsInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := conn.UpdateUserAccessLoggingSettings(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb User Access Logging Settings (%s)", new.UserAccessLoggingSettingsARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *userAccessLoggingSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userAccessLoggingSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteUserAccessLoggingSettingsInput{
		UserAccessLoggingSettingsArn: fwflex.StringFromFramework(ctx, data.UserAccessLoggingSettingsARN),
	}
	_, err := conn.DeleteUserAccessLoggingSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb User Access Logging Settings (%s)", data.UserAccessLoggingSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *userAccessLoggingSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_access_logging_settings_arn"), request, response)
}

func findUserAccessLoggingSettingsByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.UserAccessLoggingSettings, error) {
	input := workspacesweb.GetUserAccessLoggingSettingsInput{
		UserAccessLoggingSettingsArn: &arn,
	}
	output, err := conn.GetUserAccessLoggingSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.UserAccessLoggingSettings == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.UserAccessLoggingSettings, nil
}

type userAccessLoggingSettingsResourceModel struct {
	framework.WithRegionModel
	AssociatedPortalARNs         fwtypes.ListOfString `tfsdk:"associated_portal_arns"`
	KinesisStreamARN             fwtypes.ARN          `tfsdk:"kinesis_stream_arn"`
	Tags                         tftags.Map           `tfsdk:"tags"`
	TagsAll                      tftags.Map           `tfsdk:"tags_all"`
	UserAccessLoggingSettingsARN types.String         `tfsdk:"user_access_logging_settings_arn"`
}
