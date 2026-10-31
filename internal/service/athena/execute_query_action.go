// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/athena"
	awstypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_athena_execute_query, name="Execute query")
func newStartQueryExecutionAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &startQueryExecutionAction{}, nil
}

type startQueryExecutionAction struct {
	framework.ActionWithModel[startQueryExecutionActionModel]
}

type startQueryExecutionActionModel struct {
	framework.WithRegionModel
	ExecutionParameters   fwtypes.ListValueOf[types.String]                           `tfsdk:"execution_parameters"`
	QueryExecutionContext fwtypes.ListNestedObjectValueOf[QueryExecutionContextModel] `tfsdk:"query_execution_context"`
	QueryString           types.String                                                `tfsdk:"query_string"`
	ResultConfiguration   fwtypes.ListNestedObjectValueOf[ResultConfigurationModel]   `tfsdk:"result_configuration"`
	Workgroup             types.String                                                `tfsdk:"workgroup"`
	Timeout               types.Int64                                                 `tfsdk:"timeout"`
}

type QueryExecutionContextModel struct {
	Catalog  types.String `tfsdk:"catalog"`
	Database types.String `tfsdk:"database"`
}

type ResultConfigurationModel struct {
	AclConfiguration        fwtypes.ListNestedObjectValueOf[AclConfigurationModel]        `tfsdk:"acl_configuration"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[EncryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	ExpectedBucketOwner     types.String                                                  `tfsdk:"expected_bucket_owner"`
	OutputLocation          types.String                                                  `tfsdk:"output_location"`
}

type EncryptionConfigurationModel struct {
	EncryptionOption types.String `tfsdk:"encryption_option"`
	KmsKey           types.String `tfsdk:"kms_key"`
}

type AclConfigurationModel struct {
	S3AclOption types.String `tfsdk:"s3_acl_option"`
}

func (a *startQueryExecutionAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This action runs the SQL query statements contained in the query_string.",
		Attributes: map[string]schema.Attribute{
			"execution_parameters": schema.ListAttribute{
				Description: "A list of values for the parameters in a query. The values are applied sequentially to the parameters in the query in the order in which the parameters occur.",
				Optional:    true,
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1024),
				},
			},
			"query_string": schema.StringAttribute{
				Description: "The SQL query statements to be executed.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 262144),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds for query execution. Defaults to 300 seconds",
				Optional:    true,
			},
			"workgroup": schema.StringAttribute{
				Description: "The name of the workgroup in which the query is being started.",
				Optional:    true,
			},
		},

		Blocks: map[string]schema.Block{
			"query_execution_context": schema.ListNestedBlock{
				Description: "The database and data catalog context in which the query execution occurs.",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[QueryExecutionContextModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"catalog": schema.StringAttribute{
							Description: "The name of the data catalog used in the query execution.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						names.AttrDatabase: schema.StringAttribute{
							Description: "The name of the database used in the query execution. The database must exist in the catalog.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
						},
					},
				},
			},
			"result_configuration": schema.ListNestedBlock{
				Description: "Specifies information about where and how to save the results of the query execution",
				CustomType:  fwtypes.NewListNestedObjectTypeOf[ResultConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrExpectedBucketOwner: schema.StringAttribute{
							Description: "The AWS account ID that you expect to be the owner of the Amazon S3 bucket",
							Optional:    true,
							Validators: []validator.String{
								fwvalidators.AWSAccountID(),
							},
						},
						"output_location": schema.StringAttribute{
							Description: "The location in Amazon S3 where your query and calculation results are stored, such as s3://path/to/query/bucket/",
							Optional:    true,
							Validators: []validator.String{
								fwvalidators.S3URI(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"acl_configuration": schema.ListNestedBlock{
							Description: "Indicates that an Amazon S3 canned ACL should be set to control ownership of stored query results.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_acl_option": schema.StringAttribute{
										Description: "The Amazon S3 canned ACL that Athena should specify when storing query results",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("BUCKET_OWNER_FULL_CONTROL"),
										},
									},
								},
							},
						},
						names.AttrEncryptionConfiguration: schema.ListNestedBlock{
							Description: "If query and calculation results are encrypted in Amazon S3, indicates the encryption option used.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"encryption_option": schema.StringAttribute{
										Description: "Indicates whether Amazon S3 server-side encryption with Amazon S3-managed keys (SSE_S3), server-side encryption with KMS-managed keys (SSE_KMS), or client-side encryption with KMS-managed keys (CSE_KMS) is used.",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf("SSE_S3", "SSE_KMS", "CSE_KMS"),
										},
									},
									names.AttrKMSKey: schema.StringAttribute{
										Description: "For SSE_KMS and CSE_KMS, this is the KMS key ARN or ID.",
										Optional:    true,
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

func (a *startQueryExecutionAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config startQueryExecutionActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().AthenaClient(ctx)

	// Set default timeout if undefined
	timeout := fwactions.TimeoutOr(config.Timeout, 10*time.Minute)

	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Executing Athena query...")

	input := &athena.StartQueryExecutionInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, config, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Executing Athena query", map[string]any{
		"input": *input,
	})

	output, err := conn.StartQueryExecution(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to start Athena Query execution",
			fmt.Sprintf("Could not execute query: %s", err),
		)
		return
	}

	cb(ctx, "Query execution started, waiting for completion...")
	execId := output.QueryExecutionId

	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.QueryExecution], error) {
		input := &athena.GetQueryExecutionInput{QueryExecutionId: execId}
		output, err := conn.GetQueryExecution(ctx, input)
		if err != nil {
			return actionwait.FetchResult[*awstypes.QueryExecution]{}, err
		}

		res := output.QueryExecution
		return actionwait.FetchResult[*awstypes.QueryExecution]{Status: actionwait.Status(res.Status.State), Value: res}, nil
	}, actionwait.Options[*awstypes.QueryExecution]{
		Timeout:       timeout,
		Interval:      actionwait.WithBackoffDelay(backoff.DefaultSDKv2HelperRetryCompatibleDelay()),
		SuccessStates: []actionwait.Status{actionwait.Status(awstypes.QueryExecutionStateSucceeded)},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(awstypes.QueryExecutionStateRunning),
			actionwait.Status(awstypes.QueryExecutionStateQueued),
		},
		FailureStates: []actionwait.Status{
			actionwait.Status(awstypes.QueryExecutionStateFailed),
			actionwait.Status(awstypes.QueryExecutionStateCancelled),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			cb(ctx, "Athena query currently in state: %s", fr.Status)
		},
	})

	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError("Query execution timed out", "Query did not complete within the specified timeout")
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError("Query execution failed", "Query execution failed with status: "+failureErr.Error())
		} else {
			resp.Diagnostics.AddError("Error while waiting to query to complete", err.Error())
		}
		return
	}

	cb(ctx, "Athena query completed successfully. Execution ID: %s", *execId)
	tflog.Info(ctx, "Athena query completed successfully", map[string]any{
		"athena_execution_id": *execId,
	})
}
