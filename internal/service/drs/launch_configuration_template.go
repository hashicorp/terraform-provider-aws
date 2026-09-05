// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package drs

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/drs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_drs_launch_configuration_template", name="Launch Configuration Template")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/drs/types;awstypes;awstypes.LaunchConfigurationTemplate")
func newLaunchConfigurationTemplateResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &launchConfigurationTemplateResource{}, nil
}

const (
	ResNameLaunchConfigurationTemplate = "Launch Configuration Template"
)

type launchConfigurationTemplateResource struct {
	framework.ResourceWithModel[launchConfigurationTemplateResourceModel]
	framework.WithImportByID
}

func (r *launchConfigurationTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"copy_private_ip": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"copy_tags": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"export_bucket_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"launch_disposition": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LaunchDisposition](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"launch_into_source_instance": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"post_launch_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"recovery_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecoveryMode](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_instance_type_right_sizing_method": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TargetInstanceTypeRightSizingMethod](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"licensing": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[licensingModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"os_byol": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

var flexOptLaunchConfigurationTemplate = fwflex.WithFieldNamePrefix(ResNameLaunchConfigurationTemplate)

func (r *launchConfigurationTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan launchConfigurationTemplateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	var input drs.CreateLaunchConfigurationTemplateInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, flexOptLaunchConfigurationTemplate))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateLaunchConfigurationTemplate(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
	if output == nil || output.LaunchConfigurationTemplate == nil {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.LaunchConfigurationTemplate, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *launchConfigurationTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state launchConfigurationTemplateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	id := state.ID.ValueString()
	out, err := findLaunchConfigurationTemplateByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *launchConfigurationTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state launchConfigurationTemplateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input drs.UpdateLaunchConfigurationTemplateInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, flexOptLaunchConfigurationTemplate))
		if resp.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdateLaunchConfigurationTemplate(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
			return
		}
		if output == nil || output.LaunchConfigurationTemplate == nil {
			smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError(), smerr.ID, state.ID.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.LaunchConfigurationTemplate, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *launchConfigurationTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state launchConfigurationTemplateResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	id := state.ID.ValueString()
	input := drs.DeleteLaunchConfigurationTemplateInput{
		LaunchConfigurationTemplateID: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteLaunchConfigurationTemplate(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
	}
}

func (r *launchConfigurationTemplateResource) flatten(ctx context.Context, out *awstypes.LaunchConfigurationTemplate, data *launchConfigurationTemplateResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, out, data, flexOptLaunchConfigurationTemplate)...)
	if diags.HasError() {
		return diags
	}

	setTagsOut(ctx, out.Tags)

	return diags
}

func findLaunchConfigurationTemplateByID(ctx context.Context, conn *drs.Client, id string) (*awstypes.LaunchConfigurationTemplate, error) {
	input := drs.DescribeLaunchConfigurationTemplatesInput{
		LaunchConfigurationTemplateIDs: []string{id},
	}

	return findLaunchConfigurationTemplate(ctx, conn, &input)
}

func findLaunchConfigurationTemplate(ctx context.Context, conn *drs.Client, input *drs.DescribeLaunchConfigurationTemplatesInput) (*awstypes.LaunchConfigurationTemplate, error) {
	output, err := findLaunchConfigurationTemplates(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLaunchConfigurationTemplates(ctx context.Context, conn *drs.Client, input *drs.DescribeLaunchConfigurationTemplatesInput) ([]awstypes.LaunchConfigurationTemplate, error) {
	var output []awstypes.LaunchConfigurationTemplate

	pages := drs.NewDescribeLaunchConfigurationTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.Items...)
	}

	return output, nil
}

type launchConfigurationTemplateResourceModel struct {
	framework.WithRegionModel
	ARN                                 types.String                                                     `tfsdk:"arn"`
	CopyPrivateIP                       types.Bool                                                       `tfsdk:"copy_private_ip"`
	CopyTags                            types.Bool                                                       `tfsdk:"copy_tags"`
	ExportBucketARN                     types.String                                                     `tfsdk:"export_bucket_arn"`
	ID                                  types.String                                                     `tfsdk:"id"`
	LaunchDisposition                   fwtypes.StringEnum[awstypes.LaunchDisposition]                   `tfsdk:"launch_disposition"`
	LaunchIntoSourceInstance            types.Bool                                                       `tfsdk:"launch_into_source_instance"`
	Licensing                           fwtypes.ListNestedObjectValueOf[licensingModel]                  `tfsdk:"licensing"`
	PostLaunchEnabled                   types.Bool                                                       `tfsdk:"post_launch_enabled"`
	RecoveryMode                        fwtypes.StringEnum[awstypes.RecoveryMode]                        `tfsdk:"recovery_mode"`
	Tags                                tftags.Map                                                       `tfsdk:"tags"`
	TagsAll                             tftags.Map                                                       `tfsdk:"tags_all"`
	TargetInstanceTypeRightSizingMethod fwtypes.StringEnum[awstypes.TargetInstanceTypeRightSizingMethod] `tfsdk:"target_instance_type_right_sizing_method"`
}

type licensingModel struct {
	OSByol types.Bool `tfsdk:"os_byol"`
}
