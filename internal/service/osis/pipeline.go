// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Pipeline")
// @Tags(identifierAttribute="arn")
func newResourcePipeline(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePipeline{}

	r.SetDefaultCreateTimeout(45 * time.Minute)
	r.SetDefaultUpdateTimeout(45 * time.Minute)
	r.SetDefaultDeleteTimeout(45 * time.Minute)

	return r, nil
}

const (
	ResNamePipeline       = "Pipeline"
	iamPropagationTimeout = time.Minute * 1
)

type resourcePipeline struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourcePipeline) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_osis_pipeline"
}

func (r *resourcePipeline) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"ingest_endpoint_urls": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"max_units": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"min_units": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"pipeline_configuration_body": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 24000),
				},
			},
			"pipeline_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 28),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"buffer_options": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"persistent_buffer_enabled": schema.BoolAttribute{
							Required: true,
						},
					},
				},
			},
			"encryption_at_rest_options": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_key_arn": schema.StringAttribute{
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							CustomType: fwtypes.ARNType,
							Required:   true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(7, 2048),
							},
						},
					},
				},
			},
			"log_publishing_options": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"is_logging_enabled": schema.BoolAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"cloudwatch_log_destination": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 512),
										},
									},
								},
							},
						},
					},
				},
			},
			"vpc_options": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"security_group_ids": schema.SetAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
								setvalidator.ValueStringsAre(
									stringvalidator.All(
										stringvalidator.LengthAtLeast(11),
										stringvalidator.LengthAtMost(20),
									),
								),
							},
						},
						"subnet_ids": schema.SetAttribute{
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 12),
								setvalidator.ValueStringsAre(
									stringvalidator.All(
										stringvalidator.LengthAtLeast(15),
										stringvalidator.LengthAtMost(24),
									),
								),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourcePipeline) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchIngestionClient(ctx)

	var plan resourcePipelineData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &osis.CreatePipelineInput{
		MaxUnits:                  aws.Int32(int32(plan.MaxUnits.ValueInt64())),
		MinUnits:                  aws.Int32(int32(plan.MinUnits.ValueInt64())),
		PipelineConfigurationBody: aws.String(plan.PipelineConfigurationBody.ValueString()),
		PipelineName:              aws.String(plan.PipelineName.ValueString()),
		Tags:                      getTagsIn(ctx),
	}

	if !plan.BufferOptions.IsNull() {
		var bufferOptions []bufferOptionsData
		resp.Diagnostics.Append(plan.BufferOptions.ElementsAs(ctx, &bufferOptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.BufferOptions = expandBufferOptions(bufferOptions)
	}

	if !plan.EncryptionAtRestOptions.IsNull() {
		var encryptionAtRestOptions []encryptionAtRestOptionsData
		resp.Diagnostics.Append(plan.EncryptionAtRestOptions.ElementsAs(ctx, &encryptionAtRestOptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.EncryptionAtRestOptions = expandEncryptionAtRestOptions(encryptionAtRestOptions)
	}

	if !plan.LogPublishingOptions.IsNull() {
		var logPublishingOptions []logPublishingOptionsData
		resp.Diagnostics.Append(plan.LogPublishingOptions.ElementsAs(ctx, &logPublishingOptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		logPublishingOptionsInput, d := expandLogPublishingOptions(ctx, logPublishingOptions)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.LogPublishingOptions = logPublishingOptionsInput
	}

	if !plan.VpcOptions.IsNull() {
		var vpcOptions []vpcOptionsData
		resp.Diagnostics.Append(plan.VpcOptions.ElementsAs(ctx, &vpcOptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.VpcOptions = expandVPCOptions(ctx, vpcOptions)
	}

	// Retry for IAM eventual consistency
	var out *osis.CreatePipelineOutput
	err := tfresource.Retry(ctx, iamPropagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.CreatePipeline(ctx, in)
		if err != nil {
			var ve *awstypes.ValidationException
			if errors.As(err, &ve) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionCreating, ResNamePipeline, plan.PipelineName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Pipeline == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionCreating, ResNamePipeline, plan.PipelineName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, out.Pipeline.PipelineName)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitPipelineCreated(ctx, conn, aws.ToString(out.Pipeline.PipelineName), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionWaitingForCreation, ResNamePipeline, plan.PipelineName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, waitOut)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePipeline) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchIngestionClient(ctx)

	var state resourcePipelineData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPipelineByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionSetting, ResNamePipeline, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(state.refreshFromOutput(ctx, out)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePipeline) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchIngestionClient(ctx)

	var plan, state resourcePipelineData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.BufferOptions.Equal(state.BufferOptions) ||
		!plan.EncryptionAtRestOptions.Equal(state.EncryptionAtRestOptions) ||
		!plan.LogPublishingOptions.Equal(state.LogPublishingOptions) ||
		!plan.MaxUnits.Equal(state.MaxUnits) ||
		!plan.MinUnits.Equal(state.MinUnits) ||
		!plan.PipelineConfigurationBody.Equal(state.PipelineConfigurationBody) {
		in := &osis.UpdatePipelineInput{
			PipelineName: aws.String(plan.PipelineName.ValueString()),
		}

		if !plan.MaxUnits.IsNull() {
			in.MaxUnits = aws.Int32(int32(plan.MaxUnits.ValueInt64()))
		}

		if !plan.MinUnits.IsNull() {
			in.MinUnits = aws.Int32(int32(plan.MinUnits.ValueInt64()))
		}

		if !plan.PipelineConfigurationBody.IsNull() {
			in.PipelineConfigurationBody = aws.String(plan.PipelineConfigurationBody.ValueString())
		}

		if !plan.BufferOptions.IsNull() {
			var bufferOptions []bufferOptionsData
			resp.Diagnostics.Append(plan.BufferOptions.ElementsAs(ctx, &bufferOptions, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.BufferOptions = expandBufferOptions(bufferOptions)
		}

		if !plan.EncryptionAtRestOptions.IsNull() {
			var encryptionAtRestOptions []encryptionAtRestOptionsData
			resp.Diagnostics.Append(plan.EncryptionAtRestOptions.ElementsAs(ctx, &encryptionAtRestOptions, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.EncryptionAtRestOptions = expandEncryptionAtRestOptions(encryptionAtRestOptions)
		}
		if !plan.LogPublishingOptions.IsNull() {
			var logPublishingOptions []logPublishingOptionsData
			resp.Diagnostics.Append(plan.LogPublishingOptions.ElementsAs(ctx, &logPublishingOptions, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			logPublishingOptionsInput, d := expandLogPublishingOptions(ctx, logPublishingOptions)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.LogPublishingOptions = logPublishingOptionsInput
		}

		out, err := conn.UpdatePipeline(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionUpdating, ResNamePipeline, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Pipeline == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionUpdating, ResNamePipeline, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitPipelineUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionWaitingForUpdate, ResNamePipeline, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(plan.refreshFromOutput(ctx, waitOut)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourcePipeline) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchIngestionClient(ctx)

	var state resourcePipelineData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &osis.DeletePipelineInput{
		PipelineName: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeletePipeline(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionDeleting, ResNamePipeline, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitPipelineDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchIngestion, create.ErrActionWaitingForDeletion, ResNamePipeline, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// refreshFromOutput writes state data from an AWS response object
func (pd *resourcePipelineData) refreshFromOutput(ctx context.Context, out *awstypes.Pipeline) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	pd.ARN = flex.StringToFramework(ctx, out.PipelineArn)
	pd.ID = flex.StringToFramework(ctx, out.PipelineName)
	pd.PipelineName = flex.StringToFramework(ctx, out.PipelineName)
	pd.PipelineConfigurationBody = flex.StringToFramework(ctx, out.PipelineConfigurationBody)
	minUnits := int64(out.MinUnits)
	pd.MinUnits = flex.Int64ToFramework(ctx, &minUnits)
	maxUnits := int64(out.MaxUnits)
	pd.MaxUnits = flex.Int64ToFramework(ctx, &maxUnits)
	pd.IngestEndpointUrls = flex.FlattenFrameworkStringValueSet(ctx, out.IngestEndpointUrls)

	bufferOptions, d := flattenBufferOptions(ctx, out.BufferOptions)
	diags.Append(d...)
	pd.BufferOptions = bufferOptions

	encryptionAtRestOptions, d := flattenEncryptionAtRestOptions(ctx, out.EncryptionAtRestOptions)
	diags.Append(d...)
	pd.EncryptionAtRestOptions = encryptionAtRestOptions

	logPublishingOptions, d := flattenLogPublishingOptions(ctx, out.LogPublishingOptions)
	diags.Append(d...)
	pd.LogPublishingOptions = logPublishingOptions

	setTagsOut(ctx, out.Tags)
	return diags
}

func (r *resourcePipeline) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourcePipeline) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitPipelineCreated(ctx context.Context, conn *osis.Client, id string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusCreating, awstypes.PipelineStatusStarting),
		Target:     enum.Slice(awstypes.PipelineStatusActive),
		Refresh:    statusPipeline(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Pipeline); ok {
		return out, err
	}

	return nil, err
}

func waitPipelineUpdated(ctx context.Context, conn *osis.Client, id string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusUpdating),
		Target:     enum.Slice(awstypes.PipelineStatusActive),
		Refresh:    statusPipeline(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Pipeline); ok {
		return out, err
	}

	return nil, err
}

func waitPipelineDeleted(ctx context.Context, conn *osis.Client, id string, timeout time.Duration) (*awstypes.Pipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.PipelineStatusDeleting),
		Target:     []string{},
		Refresh:    statusPipeline(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Pipeline); ok {
		return out, err
	}

	return nil, err
}

func statusPipeline(ctx context.Context, conn *osis.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findPipelineByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findPipelineByID(ctx context.Context, conn *osis.Client, id string) (*awstypes.Pipeline, error) {
	in := &osis.GetPipelineInput{
		PipelineName: aws.String(id),
	}

	out, err := conn.GetPipeline(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Pipeline == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Pipeline, nil
}

func flattenBufferOptions(ctx context.Context, apiObject *awstypes.BufferOptions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: bufferOptionsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"persistent_buffer_enabled": flex.BoolToFramework(ctx, apiObject.PersistentBufferEnabled),
	}
	objVal, d := types.ObjectValue(bufferOptionsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

//func flattenEncryptionAtRestOptions(ctx context.Context, apiObject *awstypes.EncryptionAtRestOptions) (types.List, diag.Diagnostics) {
//	var diags diag.Diagnostics
//	elemType := fwtypes.NewObjectTypeOf[encryptionAtRestOptionsData](ctx).ObjectType
//
//	if apiObject == nil {
//		return types.ListValueMust(elemType, []attr.Value{}), diags
//	}
//
//	values := make([]attr.Value, len(apiObjects))
//	for i, o := range apiObjects {
//		values[i] = flattenMonitorData(ctx, o).value(ctx)
//	}
//
//	objVal := &encryptionAtRestOptionsData{
//		KmsKeyArn: flex.StringToFrameworkARN(ctx, apiObject.KmsKeyArn),
//	}
//
//	obj := map[string]attr.Value{
//		"kms_key_arn": flex.StringToFrameworkARN(ctx, apiObject.KmsKeyArn),
//	}
//	//objVal, d := types.ObjectValue(encryptionAtRestOptionsAttrTypes, obj)
//	//diags.Append(d...)
//
//	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
//	diags.Append(d...)
//
//	return listVal, diags
//}

func flattenEncryptionAtRestOptions(ctx context.Context, apiObject *awstypes.EncryptionAtRestOptions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: encryptionAtRestOptionsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		//"kms_key_arn": flex.StringToFrameworkARN(ctx, apiObject.KmsKeyArn),
		"kms_key_arn": flex.StringToFramework(ctx, apiObject.KmsKeyArn),
	}
	objVal, d := types.ObjectValue(encryptionAtRestOptionsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenLogPublishingOptions(ctx context.Context, apiObject *awstypes.LogPublishingOptions) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: logPublishingOptionsAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	cloudWatchLogDestination, d := flattenCloudWatchLogDestination(ctx, apiObject.CloudWatchLogDestination)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"is_logging_enabled":         flex.BoolToFramework(ctx, apiObject.IsLoggingEnabled),
		"cloudwatch_log_destination": cloudWatchLogDestination,
	}
	objVal, d := types.ObjectValue(logPublishingOptionsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenCloudWatchLogDestination(ctx context.Context, apiObject *awstypes.CloudWatchLogDestination) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: cloudWatchLogDestinationAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"log_group": flex.StringToFramework(ctx, apiObject.LogGroup),
	}
	objVal, d := types.ObjectValue(cloudWatchLogDestinationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandBufferOptions(tfList []bufferOptionsData) *awstypes.BufferOptions {
	if len(tfList) == 0 {
		return nil
	}
	bo := tfList[0]
	return &awstypes.BufferOptions{
		PersistentBufferEnabled: aws.Bool(bo.PersistentBufferEnabled.ValueBool()),
	}
}

func expandEncryptionAtRestOptions(tfList []encryptionAtRestOptionsData) *awstypes.EncryptionAtRestOptions {
	if len(tfList) == 0 {
		return nil
	}
	earo := tfList[0]
	return &awstypes.EncryptionAtRestOptions{
		KmsKeyArn: aws.String(earo.KmsKeyArn.ValueString()),
	}
}

func expandLogPublishingOptions(ctx context.Context, tfList []logPublishingOptionsData) (*awstypes.LogPublishingOptions, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}

	lpo := tfList[0]
	apiObject := &awstypes.LogPublishingOptions{}
	if !lpo.IsLoggingEnabled.IsNull() {
		apiObject.IsLoggingEnabled = aws.Bool(lpo.IsLoggingEnabled.ValueBool())
	}

	if !lpo.CloudWatchLogDestination.IsNull() {
		var cloudWatchLogDestination []cloudWatchLogDestinationData
		diags.Append(lpo.CloudWatchLogDestination.ElementsAs(ctx, &cloudWatchLogDestination, false)...)
		apiObject.CloudWatchLogDestination = expandCloudWatchLogDestination(cloudWatchLogDestination)
	}

	return apiObject, diags
}

func expandCloudWatchLogDestination(tfList []cloudWatchLogDestinationData) *awstypes.CloudWatchLogDestination {
	if len(tfList) == 0 {
		return nil
	}
	cwld := tfList[0]
	return &awstypes.CloudWatchLogDestination{
		LogGroup: aws.String(cwld.LogGroup.ValueString()),
	}
}

func expandVPCOptions(ctx context.Context, tfList []vpcOptionsData) *awstypes.VpcOptions {
	if len(tfList) == 0 {
		return nil
	}
	vo := tfList[0]
	apiObject := &awstypes.VpcOptions{
		SubnetIds: flex.ExpandFrameworkStringValueSet(ctx, vo.SubnetIds),
	}

	if !vo.SecurityGroupIds.IsNull() {
		apiObject.SecurityGroupIds = flex.ExpandFrameworkStringValueSet(ctx, vo.SecurityGroupIds)
	}

	return apiObject
}

type resourcePipelineData struct {
	ARN                       types.String   `tfsdk:"arn"`
	BufferOptions             types.List     `tfsdk:"buffer_options"`
	EncryptionAtRestOptions   types.List     `tfsdk:"encryption_at_rest_options"`
	ID                        types.String   `tfsdk:"id"`
	IngestEndpointUrls        types.Set      `tfsdk:"ingest_endpoint_urls"`
	LogPublishingOptions      types.List     `tfsdk:"log_publishing_options"`
	MaxUnits                  types.Int64    `tfsdk:"max_units"`
	MinUnits                  types.Int64    `tfsdk:"min_units"`
	PipelineConfigurationBody types.String   `tfsdk:"pipeline_configuration_body"`
	PipelineName              types.String   `tfsdk:"pipeline_name"`
	Tags                      types.Map      `tfsdk:"tags"`
	TagsAll                   types.Map      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
	VpcOptions                types.List     `tfsdk:"vpc_options"`
}

type bufferOptionsData struct {
	PersistentBufferEnabled types.Bool `tfsdk:"persistent_buffer_enabled"`
}

type encryptionAtRestOptionsData struct {
	KmsKeyArn fwtypes.ARN `tfsdk:"kms_key_arn"`
}

type logPublishingOptionsData struct {
	CloudWatchLogDestination types.List `tfsdk:"cloudwatch_log_destination"`
	IsLoggingEnabled         types.Bool `tfsdk:"is_logging_enabled"`
}

type cloudWatchLogDestinationData struct {
	LogGroup types.String `tfsdk:"log_group"`
}

type vpcOptionsData struct {
	SecurityGroupIds types.Set `tfsdk:"security_group_ids"`
	SubnetIds        types.Set `tfsdk:"subnet_ids"`
}

var (
	bufferOptionsAttrTypes = map[string]attr.Type{
		"persistent_buffer_enabled": types.BoolType,
	}

	encryptionAtRestOptionsAttrTypes = map[string]attr.Type{
		"kms_key_arn": types.StringType,
	}

	logPublishingOptionsAttrTypes = map[string]attr.Type{
		"cloudwatch_log_destination": types.ListType{ElemType: types.ObjectType{AttrTypes: cloudWatchLogDestinationAttrTypes}},
		"is_logging_enabled":         types.BoolType,
	}

	cloudWatchLogDestinationAttrTypes = map[string]attr.Type{
		"log_group": types.StringType,
	}
)
