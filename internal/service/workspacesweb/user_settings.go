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
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

// @FrameworkResource("aws_workspacesweb_user_settings", name="User Settings")
// @Tags(identifierAttribute="user_settings_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.UserSettings")
// @Testing(importStateIdAttribute="user_settings_arn")
func newUserSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &userSettingsResource{}, nil
}

type userSettingsResource struct {
	framework.ResourceWithModel[userSettingsResourceModel]
}

func (r *userSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"copy_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Required:   true,
			},
			"customer_managed_key": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deep_link_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disconnect_timeout_in_minutes": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 600),
				},
			},
			"download_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Required:   true,
			},
			"idle_disconnect_timeout_in_minutes": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 60),
				},
			},
			"paste_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Required:   true,
			},
			"print_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"upload_allowed": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EnabledType](),
				Required:   true,
			},
			"user_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"cookie_synchronization_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cookieSynchronizationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"allowlist": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cookieSpecificationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDomain: schema.StringAttribute{
										Required: true,
									},
									names.AttrName: schema.StringAttribute{
										Optional: true,
									},
									names.AttrPath: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"blocklist": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cookieSpecificationModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrDomain: schema.StringAttribute{
										Required: true,
									},
									names.AttrName: schema.StringAttribute{
										Optional: true,
									},
									names.AttrPath: schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"toolbar_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[toolbarConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"hidden_toolbar_items": schema.ListAttribute{
							ElementType: fwtypes.StringEnumType[awstypes.ToolbarItem](),
							Optional:    true,
						},
						"max_display_resolution": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MaxDisplayResolution](),
							Optional:   true,
						},
						"toolbar_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ToolbarType](),
							Optional:   true,
						},
						"visual_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.VisualMode](),
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *userSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	var input workspacesweb.CreateUserSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateUserSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpacesWeb User Settings", err.Error())
		return
	}

	data.UserSettingsARN = fwflex.StringToFramework(ctx, output.UserSettingsArn)

	// Get the user settings details to populate other fields
	userSettings, err := findUserSettingsByARN(ctx, conn, data.UserSettingsARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb User Settings (%s)", data.UserSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, userSettings, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findUserSettingsByARN(ctx, conn, data.UserSettingsARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb User Settings (%s)", data.UserSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old userSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	if !new.AdditionalEncryptionContext.Equal(old.AdditionalEncryptionContext) ||
		!new.CookieSynchronizationConfiguration.Equal(old.CookieSynchronizationConfiguration) ||
		!new.CopyAllowed.Equal(old.CopyAllowed) ||
		!new.DeepLinkAllowed.Equal(old.DeepLinkAllowed) ||
		!new.DisconnectTimeoutInMinutes.Equal(old.DisconnectTimeoutInMinutes) ||
		!new.DownloadAllowed.Equal(old.DownloadAllowed) ||
		!new.IdleDisconnectTimeoutInMinutes.Equal(old.IdleDisconnectTimeoutInMinutes) ||
		!new.PasteAllowed.Equal(old.PasteAllowed) ||
		!new.PrintAllowed.Equal(old.PrintAllowed) ||
		!new.ToolbarConfiguration.Equal(old.ToolbarConfiguration) ||
		!new.UploadAllowed.Equal(old.UploadAllowed) {
		var input workspacesweb.UpdateUserSettingsInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := conn.UpdateUserSettings(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb User Settings (%s)", new.UserSettingsARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *userSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteUserSettingsInput{
		UserSettingsArn: fwflex.StringFromFramework(ctx, data.UserSettingsARN),
	}
	_, err := conn.DeleteUserSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb User Settings (%s)", data.UserSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *userSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("user_settings_arn"), request, response)
}

func findUserSettingsByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.UserSettings, error) {
	input := workspacesweb.GetUserSettingsInput{
		UserSettingsArn: &arn,
	}
	output, err := conn.GetUserSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.UserSettings == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.UserSettings, nil
}

type userSettingsResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext        fwtypes.MapOfString                                                      `tfsdk:"additional_encryption_context"`
	AssociatedPortalARNs               fwtypes.ListOfString                                                     `tfsdk:"associated_portal_arns"`
	CookieSynchronizationConfiguration fwtypes.ListNestedObjectValueOf[cookieSynchronizationConfigurationModel] `tfsdk:"cookie_synchronization_configuration"`
	CopyAllowed                        fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"copy_allowed"`
	CustomerManagedKey                 fwtypes.ARN                                                              `tfsdk:"customer_managed_key"`
	DeepLinkAllowed                    fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"deep_link_allowed"`
	DisconnectTimeoutInMinutes         types.Int64                                                              `tfsdk:"disconnect_timeout_in_minutes"`
	DownloadAllowed                    fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"download_allowed"`
	IdleDisconnectTimeoutInMinutes     types.Int64                                                              `tfsdk:"idle_disconnect_timeout_in_minutes"`
	PasteAllowed                       fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"paste_allowed"`
	PrintAllowed                       fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"print_allowed"`
	Tags                               tftags.Map                                                               `tfsdk:"tags"`
	TagsAll                            tftags.Map                                                               `tfsdk:"tags_all"`
	ToolbarConfiguration               fwtypes.ListNestedObjectValueOf[toolbarConfigurationModel]               `tfsdk:"toolbar_configuration"`
	UploadAllowed                      fwtypes.StringEnum[awstypes.EnabledType]                                 `tfsdk:"upload_allowed"`
	UserSettingsARN                    types.String                                                             `tfsdk:"user_settings_arn"`
}

type cookieSynchronizationConfigurationModel struct {
	Allowlist fwtypes.ListNestedObjectValueOf[cookieSpecificationModel] `tfsdk:"allowlist"`
	Blocklist fwtypes.ListNestedObjectValueOf[cookieSpecificationModel] `tfsdk:"blocklist"`
}

type cookieSpecificationModel struct {
	Domain types.String `tfsdk:"domain"`
	Name   types.String `tfsdk:"name"`
	Path   types.String `tfsdk:"path"`
}

type toolbarConfigurationModel struct {
	HiddenToolbarItems   fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.ToolbarItem]] `tfsdk:"hidden_toolbar_items"`
	MaxDisplayResolution fwtypes.StringEnum[awstypes.MaxDisplayResolution]             `tfsdk:"max_display_resolution"`
	ToolbarType          fwtypes.StringEnum[awstypes.ToolbarType]                      `tfsdk:"toolbar_type"`
	VisualMode           fwtypes.StringEnum[awstypes.VisualMode]                       `tfsdk:"visual_mode"`
}
