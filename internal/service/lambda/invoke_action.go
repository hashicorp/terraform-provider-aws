// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_lambda_invoke, name="Invoke")
func newInvokeAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &invokeAction{}, nil
}

var (
	_ action.Action = (*invokeAction)(nil)
)

type invokeAction struct {
	framework.ActionWithModel[invokeActionModel]
}

type invokeActionModel struct {
	framework.WithRegionModel
	FunctionName   types.String                                `tfsdk:"function_name"`
	Payload        types.String                                `tfsdk:"payload"`
	Qualifier      types.String                                `tfsdk:"qualifier"`
	InvocationType fwtypes.StringEnum[awstypes.InvocationType] `tfsdk:"invocation_type"`
	LogType        fwtypes.StringEnum[awstypes.LogType]        `tfsdk:"log_type"`
	ClientContext  types.String                                `tfsdk:"client_context"`
	TenantId       types.String                                `tfsdk:"tenant_id"`
}

func (a *invokeAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Invokes an AWS Lambda function with the specified payload. This action allows for imperative invocation of Lambda functions with full control over invocation parameters.",
		Attributes: map[string]schema.Attribute{
			"function_name": schema.StringAttribute{
				Description: "The name, ARN, or partial ARN of the Lambda function to invoke. You can specify a function name (e.g., my-function), a qualified function name (e.g., my-function:PROD), or a partial ARN (e.g., 123456789012:function:my-function).",
				Required:    true,
			},
			"payload": schema.StringAttribute{
				Description: "The JSON payload to send to the Lambda function. This should be a valid JSON string that represents the event data for your function.",
				Required:    true,
				Validators: []validator.String{
					validators.JSON(),
				},
			},
			"qualifier": schema.StringAttribute{
				Description: "The version or alias of the Lambda function to invoke. If not specified, the $LATEST version will be invoked.",
				Optional:    true,
			},
			"invocation_type": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.InvocationType](),
				Description: "The invocation type. Valid values are 'RequestResponse' (synchronous), 'Event' (asynchronous), and 'DryRun' (validate parameters without invoking). Defaults to 'RequestResponse'.",
				Optional:    true,
			},
			"log_type": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.LogType](),
				Description: "Set to 'Tail' to include the execution log in the response. Only applies to synchronous invocations ('RequestResponse' invocation type). Defaults to 'None'.",
				Optional:    true,
			},
			"client_context": schema.StringAttribute{
				Description: "Up to 3,583 bytes of base64-encoded data about the invoking client to pass to the function in the context object. This is only used for mobile applications.",
				Optional:    true,
			},
			"tenant_id": schema.StringAttribute{
				Description: "The Tenant Id for lambda function invocation. This is mandatory, if tenancy_config is enabled in lambda function",
				Optional:    true,
			},
		},
	}
}

func (a *invokeAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config invokeActionModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().LambdaClient(ctx)

	functionName := config.FunctionName.ValueString()
	payload := config.Payload.ValueString()

	// Set default values for optional parameters
	invocationType := awstypes.InvocationTypeRequestResponse
	if !config.InvocationType.IsNull() && !config.InvocationType.IsUnknown() {
		invocationType = config.InvocationType.ValueEnum()
	}

	logType := awstypes.LogTypeNone
	if !config.LogType.IsNull() && !config.LogType.IsUnknown() {
		logType = config.LogType.ValueEnum()
	}

	tflog.Info(ctx, "Starting Lambda function invocation action", map[string]any{
		"function_name":      functionName,
		"invocation_type":    string(invocationType),
		"log_type":           string(logType),
		"payload_length":     len(payload),
		"has_qualifier":      !config.Qualifier.IsNull(),
		"has_client_context": !config.ClientContext.IsNull(),
	})

	// Send initial progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Invoking Lambda function %s...", functionName),
	})

	// Build the invoke input
	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		Payload:        []byte(payload),
		InvocationType: invocationType,
		LogType:        logType,
	}

	// Set optional parameters
	if !config.Qualifier.IsNull() {
		input.Qualifier = config.Qualifier.ValueStringPointer()
	}
	// Set optional parameters
	if !config.TenantId.IsNull() {
		input.TenantId = config.TenantId.ValueStringPointer()
	}

	if !config.ClientContext.IsNull() {
		clientContext := config.ClientContext.ValueString()
		// Validate that client context is base64 encoded
		if _, err := base64.StdEncoding.DecodeString(clientContext); err != nil {
			resp.Diagnostics.AddError(
				"Invalid Client Context",
				fmt.Sprintf("Client context must be base64 encoded: %s", err),
			)
			return
		}
		input.ClientContext = aws.String(clientContext)
	}

	// Perform the invocation
	output, err := conn.Invoke(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Invoke Lambda Function",
			fmt.Sprintf("Could not invoke Lambda function %s: %s", functionName, err),
		)
		return
	}

	// Handle function errors
	if output.FunctionError != nil {
		functionError := aws.ToString(output.FunctionError)
		payloadStr := string(output.Payload)

		resp.Diagnostics.AddError(
			"Lambda Function Execution Error",
			fmt.Sprintf("Lambda function %s returned an error (%s): %s", functionName, functionError, payloadStr),
		)
		return
	}

	// Handle different invocation types
	switch invocationType {
	case awstypes.InvocationTypeRequestResponse:
		a.handleSyncInvocation(resp, functionName, output, logType)

	case awstypes.InvocationTypeEvent:
		// For asynchronous invocations, we only get confirmation that the request was accepted
		statusCode := output.StatusCode
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Lambda function %s invoked asynchronously (status: %d)", functionName, statusCode),
		})

	case awstypes.InvocationTypeDryRun:
		// For dry run, we validate parameters without actually invoking
		statusCode := output.StatusCode
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Lambda function %s dry run completed successfully (status: %d)", functionName, statusCode),
		})
	}

	tflog.Info(ctx, "Lambda function invocation action completed successfully", map[string]any{
		"function_name":      functionName,
		"invocation_type":    string(invocationType),
		names.AttrStatusCode: output.StatusCode,
		"executed_version":   aws.ToString(output.ExecutedVersion),
		"has_logs":           output.LogResult != nil,
		"payload_length":     len(output.Payload),
	})
}

func (a *invokeAction) handleSyncInvocation(resp *action.InvokeResponse, functionName string, output *lambda.InvokeOutput, logType awstypes.LogType) {
	statusCode := output.StatusCode
	payloadLength := len(output.Payload)

	// Send success message
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Lambda function %s invoked successfully (status: %d, payload: %d bytes)",
			functionName, statusCode, payloadLength),
	})

	// Output logs if available
	if logType != awstypes.LogTypeTail || output.LogResult == nil {
		return
	}

	logData, err := base64.StdEncoding.DecodeString(aws.ToString(output.LogResult))
	if err != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Failed to decode Lambda logs: %s", err),
		})
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Lambda function logs:\n%s", string(logData)),
	})
}
