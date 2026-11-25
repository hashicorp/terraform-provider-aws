// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcore/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_bedrockagentcore_start_batch_evaluation, name="Start Batch Evaluation")
func newStartBatchEvaluationAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &startBatchEvaluationAction{}, nil
}

var (
	_ action.Action = (*startBatchEvaluationAction)(nil)
)

type startBatchEvaluationAction struct {
	framework.ActionWithModel[startBatchEvaluationModel]
}

type startBatchEvaluationModel struct {
	framework.WithRegionModel
	BatchEvaluationName types.String                                                `tfsdk:"batch_evaluation_name"`
	ClientToken         types.String                                                `tfsdk:"client_token"`
	DataSourceConfig    fwtypes.ListNestedObjectValueOf[batchDataSourceConfigModel] `tfsdk:"data_source_config"`
	Description         types.String                                                `tfsdk:"description"`
	Evaluators          fwtypes.ListNestedObjectValueOf[batchEvaluatorModel]        `tfsdk:"evaluators"`
	KMSKeyARN           fwtypes.ARN                                                 `tfsdk:"kms_key_arn"`
	OutputConfig        fwtypes.ListNestedObjectValueOf[batchOutputConfigModel]     `tfsdk:"output_config"`
	Tags                tftags.Map                                                  `tfsdk:"tags" autoflex:"-"`
	WaitForCompletion   types.Bool                                                  `tfsdk:"wait_for_completion" autoflex:"-"`
}

type batchDataSourceConfigModel struct {
	CloudWatchLogs               fwtypes.ListNestedObjectValueOf[batchCloudWatchLogsModel]               `tfsdk:"cloudwatch_logs"`
	OnlineEvaluationConfigSource fwtypes.ListNestedObjectValueOf[batchOnlineEvaluationConfigSourceModel] `tfsdk:"online_evaluation_config_source"`
}

type batchCloudWatchLogsModel struct {
	LogGroupNames fwtypes.ListOfString `tfsdk:"log_group_names"`
	ServiceNames  fwtypes.ListOfString `tfsdk:"service_names"`
}

type batchOnlineEvaluationConfigSourceModel struct {
	OnlineEvaluationConfigArn fwtypes.ARN `tfsdk:"online_evaluation_config_arn"`
}

type batchEvaluatorModel struct {
	EvaluatorID types.String `tfsdk:"evaluator_id"`
}

type batchOutputConfigModel struct {
	CloudWatchConfig fwtypes.ListNestedObjectValueOf[batchCloudWatchOutputModel] `tfsdk:"cloudwatch_config"`
}

type batchCloudWatchOutputModel struct {
	LogGroupName types.String `tfsdk:"log_group_name"`
}

func (a *startBatchEvaluationAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts a batch evaluation job that evaluates agent performance across multiple sessions. Batch evaluations pull agent traces from CloudWatch Logs or an existing online evaluation configuration and run specified evaluators against them.",
		Attributes: map[string]schema.Attribute{
			"batch_evaluation_name": schema.StringAttribute{
				Description: "Unique name for the batch evaluation",
				Required:    true,
				Validators: []validator.String{
					lazystringvalidator.RegexMatches(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`, "must start with a letter and contain only letters, numbers, or underscores"),
				},
			},
			"client_token": schema.StringAttribute{
				Description: "Unique, case-sensitive identifier to ensure that the API request completes no more than one time. If this token matches a previous request, the service ignores the request, but does not return an error.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(33, 256),
					lazystringvalidator.RegexMatches(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,256}$`, "must match the AWS client token format"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the batch evaluation.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(200),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrTags: tftags.TagsAttribute(),
			"wait_for_completion": schema.BoolAttribute{
				Description: "If true, the action will wait for the batch evaluation to complete before returning. If false, the action returns immediately after starting the job. Default is false.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"data_source_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[batchDataSourceConfigModel](ctx),
				Description: "Data source configuration that specifies where to pull agent session traces from for evaluation.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloudwatch_logs": schema.ListNestedBlock{
							Description: "Configuration for pulling agent session traces from CloudWatch Logs.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[batchCloudWatchLogsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_names": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Description: "List of CloudWatch log group names to read agent traces from. Maximum of 10 log groups.",
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 10),
											listvalidator.ValueStringsAre(lazystringvalidator.RegexMatches(`^[.\-_/#A-Za-z0-9]+$`, "must contain only valid CloudWatch log group name characters.")),
										},
									},
									"service_names": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Description: "List of agent service names to filter traces within the specified log groups.",
										Required:    true,
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 1),
										},
									},
								},
							},
						},
						"online_evaluation_config_source": schema.ListNestedBlock{
							Description: "Reference to an existing online evaluation configuration to use as the data source for batch evaluation.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[batchOnlineEvaluationConfigSourceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"online_evaluation_config_arn": schema.StringAttribute{
										CustomType:  fwtypes.ARNType,
										Description: "ARN of the online evaluation configuration to use as the session source.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
			"evaluators": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[batchEvaluatorModel](ctx),
				Description: "List of evaluators to apply during the batch evaluation. Can include both built-in evaluators and custom evaluators. Maximum of 10 evaluators.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"evaluator_id": schema.StringAttribute{
							Description: "Unique ID of the evaluator to apply.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 111),
								lazystringvalidator.RegexMatches(`^(Builtin\.[a-zA-Z0-9._-]+|ThirdParty\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+|[a-zA-Z][a-zA-Z0-9-_]{0,99}-[a-zA-Z0-9]{10})$`, "Invalid evaluator ID format"),
							},
						},
					},
				},
			},
			"output_config": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[batchOutputConfigModel](ctx),
				Description: "Output destination configuration.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cloudwatch_config": schema.ListNestedBlock{
							Description: "CloudWatch Logs configuration for writing evaluation results.",
							CustomType:  fwtypes.NewListNestedObjectTypeOf[batchCloudWatchOutputModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_name": schema.StringAttribute{
										Description: "The CloudWatch Logs group name where evaluation results will be published.",
										Required:    true,
										Validators: []validator.String{
											lazystringvalidator.RegexMatches(`^[.\-_/#A-Za-z0-9]+$`, "must contain only valid CloudWatch log group name characters"),
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

func (a *startBatchEvaluationAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	
}