// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"slices"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_allowed_images_settings", name="Allowed Images Settings")
func newAllowedImagesSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAllowedImagesSettings{}

	return r, nil
}

const (
	ResNameEC2AllowedImagesSettings = "Allowed Images Settings"
)

type resourceAllowedImagesSettings struct {
	framework.ResourceWithModel[resourceAllowedImagesSettingsModel]
}

func (r *resourceAllowedImagesSettings) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrState: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "enabled", "audit-mode"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"image_criteria": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[imageCriterionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"image_names": schema.SetAttribute{
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
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeAtMost(200),
								setvalidator.ValueStringsAre(
									stringvalidator.Any(
										stringvalidator.OneOf("amazon", "aws-marketplace", "aws-backup-vault", "none"),
										stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9]{12}$`), "must be a valid AWS account ID"),
									),
								),
							},
						},
						"marketplace_product_codes": schema.SetAttribute{
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
						"creation_date_condition": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[imageCriterionCreationDateConditionModel](ctx),
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(
									path.MatchRelative().AtName("maximum_days_since_created"),
								),
							},
							Attributes: map[string]schema.Attribute{
								"maximum_days_since_created": schema.Int32Attribute{
									Optional: true,
									Validators: []validator.Int32{
										int32validator.AtLeast(0),
									},
								},
							},
						},
						"deprecation_time_condition": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[imageCriterionDeprecationTimeConditionModel](ctx),
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(
									path.MatchRelative().AtName("maximum_days_since_deprecated"),
								),
							},
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
	}
}

func (r *resourceAllowedImagesSettings) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceAllowedImagesSettingsModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the allowed images settings state
	r.updateAllowedImagesSettingsState(ctx, conn, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If we enabled the setting, set the image criteria if provided
	if plan.State.ValueString() != "disabled" {
		r.updateImageCriteria(ctx, conn, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAllowedImagesSettings) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceAllowedImagesSettingsModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.GetAllowedImagesSettings(ctx, &ec2.GetAllowedImagesSettingsInput{})
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// If the setting is disabled, treat it as deleted
	if out.State != nil && *out.State == "disabled" {
		resp.State.RemoveResource(ctx)
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceAllowedImagesSettings) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state resourceAllowedImagesSettingsModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if state has changed (enabled/disabled/audit-mode)
	if !plan.State.Equal(state.State) {
		r.updateAllowedImagesSettingsState(ctx, conn, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.State.ValueString() != "disabled" {
		r.updateImageCriteria(ctx, conn, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceAllowedImagesSettings) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceAllowedImagesSettingsModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	clearCriteriaInput := &ec2.ReplaceImageCriteriaInAllowedImagesSettingsInput{}
	criteriaOut, err := conn.ReplaceImageCriteriaInAllowedImagesSettings(ctx, clearCriteriaInput)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
	if criteriaOut == nil || !*criteriaOut.ReturnValue {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("failed to clear image criteria"))
		return
	}

	// Disable the allowed images settings as we can't delete it
	input := &ec2.DisableAllowedImagesSettingsInput{}

	out, err := conn.DisableAllowedImagesSettings(ctx, input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("response from disabling allowed images settings was nil"))
		return
	}
	if out.AllowedImagesSettingsState != "disabled" {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("failed to disable allowed images settings"), "API returned unexpected state: "+string(out.AllowedImagesSettingsState))
		return
	}
}

func (r *resourceAllowedImagesSettings) updateAllowedImagesSettingsState(ctx context.Context, conn *ec2.Client, plan *resourceAllowedImagesSettingsModel, diags *diag.Diagnostics) {
	stateValue := plan.State.ValueString()

	if slices.Contains(awstypes.AllowedImagesSettingsEnabledState.Values(""), awstypes.AllowedImagesSettingsEnabledState(stateValue)) {
		// Enable with the specified state (enabled or audit-mode)
		input := ec2.EnableAllowedImagesSettingsInput{
			AllowedImagesSettingsState: awstypes.AllowedImagesSettingsEnabledState(stateValue),
		}

		out, err := conn.EnableAllowedImagesSettings(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, diags, err)
			return
		}
		if out == nil {
			smerr.AddError(ctx, diags, errors.New("response from enabling allowed images settings was nil"))
			return
		}
		if out.AllowedImagesSettingsState != input.AllowedImagesSettingsState {
			smerr.AddError(ctx, diags, errors.New("returned state setting does not match what was requested"), "API returned unexpected state: "+string(out.AllowedImagesSettingsState))
			return
		}
		smerr.EnrichAppend(ctx, diags, flex.Flatten(ctx, out, plan))
	} else if slices.Contains(awstypes.AllowedImagesSettingsDisabledState.Values(""), awstypes.AllowedImagesSettingsDisabledState(stateValue)) {
		// Disable
		input := &ec2.DisableAllowedImagesSettingsInput{}

		out, err := conn.DisableAllowedImagesSettings(ctx, input)
		if err != nil {
			smerr.AddError(ctx, diags, err)
			return
		}
		if out == nil {
			smerr.AddError(ctx, diags, errors.New("response from disabling allowed images settings was nil"))
			return
		}
		if out.AllowedImagesSettingsState != "disabled" {
			smerr.AddError(ctx, diags, errors.New("failed to disable allowed images settings"), "API returned unexpected state: "+string(out.AllowedImagesSettingsState))
			return
		}
		smerr.EnrichAppend(ctx, diags, flex.Flatten(ctx, out, plan))
	} else {
		smerr.AddError(ctx, diags, errors.New("invalid state requested"))
	}
}

func (r *resourceAllowedImagesSettings) updateImageCriteria(ctx context.Context, conn *ec2.Client, plan *resourceAllowedImagesSettingsModel, diags *diag.Diagnostics) {
	if plan.ImageCriteria.IsUnknown() {
		return
	}

	var input ec2.ReplaceImageCriteriaInAllowedImagesSettingsInput

	// AWS keeps image criteria options, even if set to disabled - set to empty
	if !plan.ImageCriteria.IsNull() {
		smerr.EnrichAppend(ctx, diags, flex.Expand(ctx, plan, &input))
		if diags.HasError() {
			return
		}
	}

	out, err := conn.ReplaceImageCriteriaInAllowedImagesSettings(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, diags, err)
		return
	}
	if out == nil {
		smerr.AddError(ctx, diags, errors.New("response from replacing image criteria was nil"))
		return
	}
	if !*out.ReturnValue {
		smerr.AddError(ctx, diags, errors.New("response from replacing image criteria indicated failure"))
		return
	}
	smerr.EnrichAppend(ctx, diags, flex.Flatten(ctx, out, plan))
}

func (r *resourceAllowedImagesSettings) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrRegion), req, resp)
}

type resourceAllowedImagesSettingsModel struct {
	framework.WithRegionModel
	State         types.String                                         `tfsdk:"state"`
	ImageCriteria fwtypes.ListNestedObjectValueOf[imageCriterionModel] `tfsdk:"image_criteria"`
}

type imageCriterionModel struct {
	ImageNames               fwtypes.SetValueOf[types.String]                                   `tfsdk:"image_names"`
	ImageProviders           fwtypes.SetValueOf[types.String]                                   `tfsdk:"image_providers"`
	MarketplaceProductCodes  fwtypes.SetValueOf[types.String]                                   `tfsdk:"marketplace_product_codes"`
	CreationDateCondition    fwtypes.ObjectValueOf[imageCriterionCreationDateConditionModel]    `tfsdk:"creation_date_condition"`
	DeprecationTimeCondition fwtypes.ObjectValueOf[imageCriterionDeprecationTimeConditionModel] `tfsdk:"deprecation_time_condition"`
}

type imageCriterionCreationDateConditionModel struct {
	MaximumDaysSinceCreated types.Int32 `tfsdk:"maximum_days_since_created"`
}

type imageCriterionDeprecationTimeConditionModel struct {
	MaximumDaysSinceDeprecated types.Int32 `tfsdk:"maximum_days_since_deprecated"`
}
