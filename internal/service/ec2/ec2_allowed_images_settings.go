// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_allowed_images_settings", name="Allowed Images Settings")
func newAllowedImagesSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &allowedImagesSettingsResource{}

	return r, nil
}

type allowedImagesSettingsResource struct {
	framework.ResourceWithModel[allowedImagesSettingsResourceModel]
}

func (r *allowedImagesSettingsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AllowedImagesSettingsEnabledState](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"image_criterion": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[imageCriterionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"image_names": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtMost(50),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(1, 128),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9\-\._/\?\[\]@'\(\)\*\w]+$`), "can only contain valid characters"),
								),
							},
						},
						"image_providers": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtMost(200),
								setvalidator.ValueStringsAre(
									stringvalidator.Any(
										stringvalidator.OneOf("amazon", "aws-marketplace", "aws-backup-vault", "none"),
										fwvalidators.AWSAccountID(),
									),
								),
							},
						},
						"marketplace_product_codes": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtMost(50),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(1, 25),
									stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9]+$`), "must be a valid marketplace product code"),
								),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"creation_date_condition": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[creationDateConditionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_days_since_created": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(0),
										},
									},
								},
							},
						},
						"deprecation_time_condition": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[deprecationTimeConditionModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"maximum_days_since_deprecated": schema.Int32Attribute{
										Optional: true,
										Validators: []validator.Int32{
											int32validator.AtLeast(0),
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

func (r *allowedImagesSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data allowedImagesSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	inputE := ec2.EnableAllowedImagesSettingsInput{
		AllowedImagesSettingsState: data.State.ValueEnum(),
	}
	_, err := conn.EnableAllowedImagesSettings(ctx, &inputE)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	var inputR ec2.ReplaceImageCriteriaInAllowedImagesSettingsInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &inputR))
	if response.Diagnostics.HasError() {
		return
	}

	_, err = conn.ReplaceImageCriteriaInAllowedImagesSettings(ctx, &inputR)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *allowedImagesSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data allowedImagesSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	out, err := findAllowedImagesSettings(ctx, conn)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *allowedImagesSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old allowedImagesSettingsResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	if !new.State.Equal(old.State) {
		input := ec2.EnableAllowedImagesSettingsInput{
			AllowedImagesSettingsState: new.State.ValueEnum(),
		}
		_, err := conn.EnableAllowedImagesSettings(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
	}

	if !new.ImageCriteria.Equal(old.ImageCriteria) {
		var input ec2.ReplaceImageCriteriaInAllowedImagesSettingsInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.ReplaceImageCriteriaInAllowedImagesSettings(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *allowedImagesSettingsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var inputR ec2.ReplaceImageCriteriaInAllowedImagesSettingsInput
	_, err := conn.ReplaceImageCriteriaInAllowedImagesSettings(ctx, &inputR)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	var inputD ec2.DisableAllowedImagesSettingsInput
	_, err = conn.DisableAllowedImagesSettings(ctx, &inputD)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}
}

func (r *allowedImagesSettingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrRegion), request, response)
}

type allowedImagesSettingsResourceModel struct {
	framework.WithRegionModel
	ImageCriteria fwtypes.ListNestedObjectValueOf[imageCriterionModel]           `tfsdk:"image_criterion"`
	State         fwtypes.StringEnum[awstypes.AllowedImagesSettingsEnabledState] `tfsdk:"state"`
}

type imageCriterionModel struct {
	CreationDateCondition    fwtypes.ListNestedObjectValueOf[creationDateConditionModel]    `tfsdk:"creation_date_condition"`
	DeprecationTimeCondition fwtypes.ListNestedObjectValueOf[deprecationTimeConditionModel] `tfsdk:"deprecation_time_condition"`
	ImageNames               fwtypes.SetOfString                                            `tfsdk:"image_names"`
	ImageProviders           fwtypes.SetOfString                                            `tfsdk:"image_providers"`
	MarketplaceProductCodes  fwtypes.SetOfString                                            `tfsdk:"marketplace_product_codes"`
}

type creationDateConditionModel struct {
	MaximumDaysSinceCreated types.Int32 `tfsdk:"maximum_days_since_created"`
}

type deprecationTimeConditionModel struct {
	MaximumDaysSinceDeprecated types.Int32 `tfsdk:"maximum_days_since_deprecated"`
}
