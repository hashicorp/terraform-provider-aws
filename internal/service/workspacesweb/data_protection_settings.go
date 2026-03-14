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

// @FrameworkResource("aws_workspacesweb_data_protection_settings", name="Data Protection Settings")
// @Tags(identifierAttribute="data_protection_settings_arn")
// @Testing(tagsTest=true)
// @Testing(generator=false)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workspacesweb/types;types.DataProtectionSettings")
// @Testing(importStateIdAttribute="data_protection_settings_arn")
func newDataProtectionSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &dataProtectionSettingsResource{}, nil
}

type dataProtectionSettingsResource struct {
	framework.ResourceWithModel[dataProtectionSettingsResourceModel]
}

func (r *dataProtectionSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"data_protection_settings_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"inline_redaction_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inlineRedactionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"global_confidence_level": schema.Int64Attribute{
							Optional: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 3),
							},
						},
						"global_enforced_urls": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"global_exempt_urls": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"inline_redaction_pattern": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[inlineRedactionPatternModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(150),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"built_in_pattern_id": schema.StringAttribute{
										Optional: true,
									},
									"confidence_level": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 3),
										},
									},
									"enforced_urls": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"exempt_urls": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
								},
								Blocks: map[string]schema.Block{
									"custom_pattern": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[customPatternModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"keyword_regex": schema.StringAttribute{
													Optional: true,
												},
												"pattern_description": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 256),
													},
												},
												"pattern_name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 20),
													},
												},
												"pattern_regex": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"redaction_place_holder": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[redactionPlaceHolderModel](ctx),
										Validators: []validator.List{
											listvalidator.ExactlyOneOf(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"redaction_place_holder_text": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthAtMost(20),
													},
												},
												"redaction_place_holder_type": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.RedactionPlaceHolderType](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *dataProtectionSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data dataProtectionSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.DisplayName)
	var input workspacesweb.CreateDataProtectionSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDataProtectionSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WorkSpacesWeb Data Protection Settings (%s)", name), err.Error())
		return
	}

	data.DataProtectionSettingsARN = fwflex.StringToFramework(ctx, output.DataProtectionSettingsArn)

	// Get the data protection settings details to populate other fields
	dataProtectionSettings, err := findDataProtectionSettingsByARN(ctx, conn, data.DataProtectionSettingsARN.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Data Protection Settings (%s)", data.DataProtectionSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, dataProtectionSettings, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *dataProtectionSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data dataProtectionSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	output, err := findDataProtectionSettingsByARN(ctx, conn, data.DataProtectionSettingsARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpacesWeb Data Protection Settings (%s)", data.DataProtectionSettingsARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataProtectionSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old dataProtectionSettingsResourceModel
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
		!new.InlineRedactionConfiguration.Equal(old.InlineRedactionConfiguration) {
		var input workspacesweb.UpdateDataProtectionSettingsInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		_, err := conn.UpdateDataProtectionSettings(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating WorkSpacesWeb Data Protection Settings (%s)", new.DataProtectionSettingsARN.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *dataProtectionSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data dataProtectionSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesWebClient(ctx)

	input := workspacesweb.DeleteDataProtectionSettingsInput{
		DataProtectionSettingsArn: fwflex.StringFromFramework(ctx, data.DataProtectionSettingsARN),
	}
	_, err := conn.DeleteDataProtectionSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpacesWeb Data Protection Settings (%s)", data.DataProtectionSettingsARN.ValueString()), err.Error())
		return
	}
}

func (r *dataProtectionSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("data_protection_settings_arn"), request, response)
}

func findDataProtectionSettingsByARN(ctx context.Context, conn *workspacesweb.Client, arn string) (*awstypes.DataProtectionSettings, error) {
	input := workspacesweb.GetDataProtectionSettingsInput{
		DataProtectionSettingsArn: &arn,
	}
	output, err := conn.GetDataProtectionSettings(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataProtectionSettings == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.DataProtectionSettings, nil
}

type dataProtectionSettingsResourceModel struct {
	framework.WithRegionModel
	AdditionalEncryptionContext  fwtypes.MapOfString                                                `tfsdk:"additional_encryption_context"`
	AssociatedPortalARNs         fwtypes.ListOfString                                               `tfsdk:"associated_portal_arns"`
	CustomerManagedKey           fwtypes.ARN                                                        `tfsdk:"customer_managed_key"`
	DataProtectionSettingsARN    types.String                                                       `tfsdk:"data_protection_settings_arn"`
	Description                  types.String                                                       `tfsdk:"description"`
	DisplayName                  types.String                                                       `tfsdk:"display_name"`
	InlineRedactionConfiguration fwtypes.ListNestedObjectValueOf[inlineRedactionConfigurationModel] `tfsdk:"inline_redaction_configuration"`
	Tags                         tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                      tftags.Map                                                         `tfsdk:"tags_all"`
}

type inlineRedactionConfigurationModel struct {
	GlobalConfidenceLevel   types.Int64                                                  `tfsdk:"global_confidence_level"`
	GlobalEnforcedURLs      fwtypes.ListOfString                                         `tfsdk:"global_enforced_urls"`
	GlobalExemptURLs        fwtypes.ListOfString                                         `tfsdk:"global_exempt_urls"`
	InlineRedactionPatterns fwtypes.ListNestedObjectValueOf[inlineRedactionPatternModel] `tfsdk:"inline_redaction_pattern"`
}

type inlineRedactionPatternModel struct {
	BuiltInPatternID     types.String                                               `tfsdk:"built_in_pattern_id"`
	ConfidenceLevel      types.Int64                                                `tfsdk:"confidence_level"`
	CustomPattern        fwtypes.ListNestedObjectValueOf[customPatternModel]        `tfsdk:"custom_pattern"`
	EnforcedURLs         fwtypes.ListOfString                                       `tfsdk:"enforced_urls"`
	ExemptURLs           fwtypes.ListOfString                                       `tfsdk:"exempt_urls"`
	RedactionPlaceHolder fwtypes.ListNestedObjectValueOf[redactionPlaceHolderModel] `tfsdk:"redaction_place_holder"`
}

type customPatternModel struct {
	KeywordRegex       types.String `tfsdk:"keyword_regex"`
	PatternDescription types.String `tfsdk:"pattern_description"`
	PatternName        types.String `tfsdk:"pattern_name"`
	PatternRegex       types.String `tfsdk:"pattern_regex"`
}

type redactionPlaceHolderModel struct {
	RedactionPlaceHolderText types.String                                          `tfsdk:"redaction_place_holder_text"`
	RedactionPlaceHolderType fwtypes.StringEnum[awstypes.RedactionPlaceHolderType] `tfsdk:"redaction_place_holder_type"`
}
