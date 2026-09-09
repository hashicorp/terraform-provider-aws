// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_model_invocation_job", name="Model Invocation Job")
// @ArnIdentity("job_arn")
// @Testing(preCheck="testAccPreCheckModelInvocationJob")
// @Testing(importIgnore="status")
// @Testing(plannableImportAction="NoOp")
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
func newModelInvocationJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &modelInvocationJobResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type modelInvocationJobResource struct {
	framework.ResourceWithModel[modelInvocationJobResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *modelInvocationJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"end_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"error_record_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"job_arn": framework.ARNAttributeComputedOnly(),
			"job_expiration_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"model_invocation_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ModelInvocationType](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"processed_record_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ModelInvocationJobStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"submit_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"success_record_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"timeout_duration_in_hours": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
					int32planmodifier.RequiresReplace(),
				},
			},
			"total_record_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"input_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelInvocationJobInputDataConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_input_data_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3InputDataConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_bucket_owner": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"s3_input_format": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.S3InputFormat](),
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											fwvalidators.S3URI(),
										},
									},
								},
							},
						},
					},
				},
			},
			"output_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelInvocationJobOutputDataConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_output_data_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3OutputDataConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_bucket_owner": schema.StringAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"s3_encryption_key_id": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										Computed:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
									},
									"s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											fwvalidators.S3URI(),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
			names.AttrVPCConfig: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[modelInvocationJobVpcConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *modelInvocationJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan modelInvocationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrock.CreateModelInvocationJobInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(create.UniqueId(ctx))

	output, err := conn.CreateModelInvocationJob(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.JobName.ValueString())
		return
	}

	arn := aws.ToString(output.JobArn)

	findOutput, err := tfresource.RetryWhenNewResourceNotFound(ctx, r.CreateTimeout(ctx, plan.Timeouts), func(ctx context.Context) (*bedrock.GetModelInvocationJobOutput, error) {
		return findModelInvocationJobByARN(ctx, conn, arn)
	}, true)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, findOutput, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *modelInvocationJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state modelInvocationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := state.JobARN.ValueString()
	job, err := findModelInvocationJobByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, job, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *modelInvocationJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state modelInvocationJobResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		return
	}

	// StopModelInvocationJob only supports jobs that haven't already reached
	// a terminal state; terminal jobs need no action.
	switch state.Status.ValueEnum() {
	case awstypes.ModelInvocationJobStatusCompleted,
		awstypes.ModelInvocationJobStatusFailed,
		awstypes.ModelInvocationJobStatusStopped,
		awstypes.ModelInvocationJobStatusPartiallyCompleted,
		awstypes.ModelInvocationJobStatusExpired:
		return
	}

	arn := state.JobARN.ValueString()

	conn := r.Meta().BedrockClient(ctx)
	input := bedrock.StopModelInvocationJobInput{
		JobIdentifier: aws.String(arn),
	}

	_, err := conn.StopModelInvocationJob(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitModelInvocationJobStopped(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

// modelIDFromResource returns id from an AWS-returned model identifier.
// GetModelInvocationJob resolves a model_id containing an inference profile
// ID (e.g. "us.amazon.nova-2-lite-v1:0") to its fully-qualified ARN
// (e.g. "arn:aws:bedrock:us-west-2:123456789012:inference-profile/us.amazon.nova-2-lite-v1:0").
// Truncate the value back to just model_id.
func modelIDFromResource(id string) string {
	if !arn.IsARN(id) {
		return id
	}

	parsed, err := arn.Parse(id)
	if err != nil {
		return id
	}

	if _, modelID, ok := strings.Cut(parsed.Resource, "/"); ok {
		return modelID
	}

	return id
}

func (r *modelInvocationJobResource) flatten(ctx context.Context, job *bedrock.GetModelInvocationJobOutput, data *modelInvocationJobResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, job, data)...)
	if diags.HasError() {
		return diags
	}

	data.ModelID = fwflex.StringValueToFramework(ctx, modelIDFromResource(aws.ToString(job.ModelId)))

	return diags
}

func findModelInvocationJobByARN(ctx context.Context, conn *bedrock.Client, arn string) (*bedrock.GetModelInvocationJobOutput, error) {
	input := bedrock.GetModelInvocationJobInput{
		JobIdentifier: aws.String(arn),
	}

	output, err := conn.GetModelInvocationJob(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusModelInvocationJob(conn *bedrock.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findModelInvocationJobByARN(ctx, conn, arn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitModelInvocationJobStopped(ctx context.Context, conn *bedrock.Client, arn string, timeout time.Duration) (*bedrock.GetModelInvocationJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ModelInvocationJobStatusSubmitted,
			awstypes.ModelInvocationJobStatusValidating,
			awstypes.ModelInvocationJobStatusScheduled,
			awstypes.ModelInvocationJobStatusInProgress,
			awstypes.ModelInvocationJobStatusStopping,
		),
		Target:  enum.Slice(awstypes.ModelInvocationJobStatusStopped),
		Refresh: statusModelInvocationJob(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*bedrock.GetModelInvocationJobOutput); ok {
		return output, err
	}

	return nil, err
}

type modelInvocationJobResourceModel struct {
	framework.WithRegionModel
	EndTime                timetypes.RFC3339                                                        `tfsdk:"end_time"`
	ErrorRecordCount       types.Int64                                                              `tfsdk:"error_record_count"`
	InputDataConfig        fwtypes.ListNestedObjectValueOf[modelInvocationJobInputDataConfigModel]  `tfsdk:"input_data_config"`
	JobARN                 types.String                                                             `tfsdk:"job_arn"`
	JobExpirationTime      timetypes.RFC3339                                                        `tfsdk:"job_expiration_time"`
	JobName                types.String                                                             `tfsdk:"job_name"`
	ModelID                types.String                                                             `tfsdk:"model_id"`
	ModelInvocationType    fwtypes.StringEnum[awstypes.ModelInvocationType]                         `tfsdk:"model_invocation_type"`
	OutputDataConfig       fwtypes.ListNestedObjectValueOf[modelInvocationJobOutputDataConfigModel] `tfsdk:"output_data_config"`
	ProcessedRecordCount   types.Int64                                                              `tfsdk:"processed_record_count"`
	RoleARN                fwtypes.ARN                                                              `tfsdk:"role_arn"`
	SkipDestroy            types.Bool                                                               `tfsdk:"skip_destroy"`
	Status                 fwtypes.StringEnum[awstypes.ModelInvocationJobStatus]                    `tfsdk:"status"`
	SubmitTime             timetypes.RFC3339                                                        `tfsdk:"submit_time"`
	SuccessRecordCount     types.Int64                                                              `tfsdk:"success_record_count"`
	TimeoutDurationInHours types.Int32                                                              `tfsdk:"timeout_duration_in_hours"`
	Timeouts               timeouts.Value                                                           `tfsdk:"timeouts"`
	TotalRecordCount       types.Int64                                                              `tfsdk:"total_record_count"`
	VpcConfig              fwtypes.ListNestedObjectValueOf[modelInvocationJobVpcConfigModel]        `tfsdk:"vpc_config"`
}

// modelInvocationJobInputDataConfigModel maps to the
// awstypes.ModelInvocationJobInputDataConfig union (S3InputDataConfig).
type modelInvocationJobInputDataConfigModel struct {
	S3InputDataConfig fwtypes.ListNestedObjectValueOf[s3InputDataConfigModel] `tfsdk:"s3_input_data_config"`
}

var (
	_ fwflex.Expander  = modelInvocationJobInputDataConfigModel{}
	_ fwflex.Flattener = &modelInvocationJobInputDataConfigModel{}
)

func (m modelInvocationJobInputDataConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.S3InputDataConfig.IsNull():
		data, d := m.S3InputDataConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ModelInvocationJobInputDataConfigMemberS3InputDataConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *modelInvocationJobInputDataConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ModelInvocationJobInputDataConfigMemberS3InputDataConfig:
		var data s3InputDataConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.S3InputDataConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("input data config flatten: %T", v))
	}

	return diags
}

type s3InputDataConfigModel struct {
	S3URI         types.String `tfsdk:"s3_uri"`
	S3BucketOwner types.String `tfsdk:"s3_bucket_owner"`
	S3InputFormat types.String `tfsdk:"s3_input_format"`
}

// modelInvocationJobOutputDataConfigModel maps to the
// awstypes.ModelInvocationJobOutputDataConfig union (S3OutputDataConfig).
type modelInvocationJobOutputDataConfigModel struct {
	S3OutputDataConfig fwtypes.ListNestedObjectValueOf[s3OutputDataConfigModel] `tfsdk:"s3_output_data_config"`
}

var (
	_ fwflex.Expander  = modelInvocationJobOutputDataConfigModel{}
	_ fwflex.Flattener = &modelInvocationJobOutputDataConfigModel{}
)

func (m modelInvocationJobOutputDataConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.S3OutputDataConfig.IsNull():
		data, d := m.S3OutputDataConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.ModelInvocationJobOutputDataConfigMemberS3OutputDataConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}

	return nil, diags
}

func (m *modelInvocationJobOutputDataConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ModelInvocationJobOutputDataConfigMemberS3OutputDataConfig:
		var data s3OutputDataConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.S3OutputDataConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("output data config flatten: %T", v))
	}

	return diags
}

type s3OutputDataConfigModel struct {
	S3URI             types.String `tfsdk:"s3_uri"`
	S3BucketOwner     types.String `tfsdk:"s3_bucket_owner"`
	S3EncryptionKeyID fwtypes.ARN  `tfsdk:"s3_encryption_key_id"`
}

type modelInvocationJobVpcConfigModel struct {
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetOfString `tfsdk:"subnet_ids"`
}
