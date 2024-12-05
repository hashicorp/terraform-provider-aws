// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_anomaly_detector", name="Anomaly Detector")
// @Tags(identifierAttribute="arn")
// @Testing(importStateIdFunc="testAccAnomalyDetectorImportStateIDFunc")
// @Testing(importStateIdAttribute="arn")
// @Testing(importIgnore="enabled")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/logs;cloudwatchlogs.GetLogAnomalyDetectorOutput")
func newResourceAnomalyDetector(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAnomalyDetector{}

	return r, nil
}

const (
	ResNameAnomalyDetector = "Anomaly Detector"
)

type resourceAnomalyDetector struct {
	framework.ResourceWithConfigure
}

func (r *resourceAnomalyDetector) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudwatch_log_anomaly_detector"
}

func (r *resourceAnomalyDetector) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"log_group_arn_list": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			"anomaly_visibility_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(7, 90),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"detector_name": schema.StringAttribute{
				Optional: true,
			},
			"evaluation_frequency": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationFrequency](),
				Optional:   true,
			},
			"filter_pattern": schema.StringAttribute{
				Optional: true,
			},
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *resourceAnomalyDetector) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan resourceAnomalyDetectorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudwatchlogs.CreateLogAnomalyDetectorInput{
		Tags: getTagsIn(ctx),
	}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateLogAnomalyDetector(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameAnomalyDetector, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameAnomalyDetector, plan.ARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.AnomalyDetectorArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAnomalyDetector) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceAnomalyDetectorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findLogAnomalyDetectorByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionSetting, ResNameAnomalyDetector, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.AnomalyVisibilityTime = flex.Int64ToFramework(ctx, out.AnomalyVisibilityTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAnomalyDetector) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan, state resourceAnomalyDetectorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Calculate(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := &cloudwatchlogs.UpdateLogAnomalyDetectorInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.AnomalyDetectorArn = plan.ARN.ValueStringPointer()
		in.AnomalyVisibilityTime = plan.AnomalyVisibilityTime.ValueInt64Pointer()

		out, err := conn.UpdateLogAnomalyDetector(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameAnomalyDetector, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameAnomalyDetector, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAnomalyDetector) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceAnomalyDetectorData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &cloudwatchlogs.DeleteLogAnomalyDetectorInput{
		AnomalyDetectorArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteLogAnomalyDetector(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionDeleting, ResNameAnomalyDetector, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAnomalyDetector) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func (r *resourceAnomalyDetector) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findLogAnomalyDetectorByARN(ctx context.Context, conn *cloudwatchlogs.Client, arn string) (*cloudwatchlogs.GetLogAnomalyDetectorOutput, error) {
	in := &cloudwatchlogs.GetLogAnomalyDetectorInput{
		AnomalyDetectorArn: aws.String(arn),
	}

	out, err := conn.GetLogAnomalyDetector(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceAnomalyDetectorData struct {
	ARN                   types.String                                     `tfsdk:"arn"`
	LogGroupARNList       fwtypes.ListValueOf[types.String]                `tfsdk:"log_group_arn_list"`
	AnomalyVisibilityTime types.Int64                                      `tfsdk:"anomaly_visibility_time"`
	DetectorName          types.String                                     `tfsdk:"detector_name"`
	Enabled               types.Bool                                       `tfsdk:"enabled"`
	EvaluationFrequency   fwtypes.StringEnum[awstypes.EvaluationFrequency] `tfsdk:"evaluation_frequency"`
	FilterPattern         types.String                                     `tfsdk:"filter_pattern"`
	KMSKeyID              types.String                                     `tfsdk:"kms_key_id"`
	Tags                  tftags.Map                                       `tfsdk:"tags"`
	TagsAll               tftags.Map                                       `tfsdk:"tags_all"`
}
