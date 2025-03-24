// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"

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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
func newAnomalyDetectorResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &anomalyDetectorResource{}

	return r, nil
}

type anomalyDetectorResource struct {
	framework.ResourceWithConfigure
}

func (r *anomalyDetectorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"detector_name": schema.StringAttribute{
				Optional: true,
			},
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			"evaluation_frequency": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationFrequency](),
				Optional:   true,
			},
			"filter_pattern": schema.StringAttribute{
				Optional: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"log_group_arn_list": schema.ListAttribute{
				CustomType:  fwtypes.ListOfARNType,
				Required:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *anomalyDetectorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data anomalyDetectorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := &cloudwatchlogs.CreateLogAnomalyDetectorInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	outputCLAD, err := conn.CreateLogAnomalyDetector(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating CloudWatch Logs Anomaly Detector", err.Error())

		return
	}

	// Set values for unknowns.
	data.AnomalyDetectorARN = fwflex.StringToFramework(ctx, outputCLAD.AnomalyDetectorArn)

	outputGLAD, err := findLogAnomalyDetectorByARN(ctx, conn, data.AnomalyDetectorARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Anomaly Detector (%s)", data.AnomalyDetectorARN.ValueString()), err.Error())

		return
	}

	data.AnomalyVisibilityTime = fwflex.Int64ToFramework(ctx, outputGLAD.AnomalyVisibilityTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anomalyDetectorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data anomalyDetectorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findLogAnomalyDetectorByARN(ctx, conn, data.AnomalyDetectorARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Anomaly Detector (%s)", data.AnomalyDetectorARN.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *anomalyDetectorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new anomalyDetectorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := &cloudwatchlogs.UpdateLogAnomalyDetectorInput{}

		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateLogAnomalyDetector(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Anomaly Detector (%s)", new.AnomalyDetectorARN.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *anomalyDetectorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data anomalyDetectorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteLogAnomalyDetector(ctx, &cloudwatchlogs.DeleteLogAnomalyDetectorInput{
		AnomalyDetectorArn: data.AnomalyDetectorARN.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Anomaly Detector (%s)", data.AnomalyDetectorARN.ValueString()), err.Error())

		return
	}
}

func (r *anomalyDetectorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func findLogAnomalyDetectorByARN(ctx context.Context, conn *cloudwatchlogs.Client, arn string) (*cloudwatchlogs.GetLogAnomalyDetectorOutput, error) {
	input := cloudwatchlogs.GetLogAnomalyDetectorInput{
		AnomalyDetectorArn: aws.String(arn),
	}

	return findLogAnomalyDetector(ctx, conn, &input)
}

func findLogAnomalyDetector(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetLogAnomalyDetectorInput) (*cloudwatchlogs.GetLogAnomalyDetectorOutput, error) {
	output, err := conn.GetLogAnomalyDetector(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type anomalyDetectorResourceModel struct {
	AnomalyDetectorARN    types.String                                     `tfsdk:"arn"`
	AnomalyVisibilityTime types.Int64                                      `tfsdk:"anomaly_visibility_time"`
	DetectorName          types.String                                     `tfsdk:"detector_name"`
	Enabled               types.Bool                                       `tfsdk:"enabled"`
	EvaluationFrequency   fwtypes.StringEnum[awstypes.EvaluationFrequency] `tfsdk:"evaluation_frequency"`
	FilterPattern         types.String                                     `tfsdk:"filter_pattern"`
	KMSKeyID              types.String                                     `tfsdk:"kms_key_id"`
	LogGroupARNList       fwtypes.ListOfARN                                `tfsdk:"log_group_arn_list"`
	Tags                  tftags.Map                                       `tfsdk:"tags"`
	TagsAll               tftags.Map                                       `tfsdk:"tags_all"`
}
