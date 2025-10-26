// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// sendCommandPollInterval defines polling cadence for send command action.
const sendCommandPollInterval = 5 * time.Second

// @Action(aws_ssm_send_command, name="Send Command")
func newSendCommandAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &sendCommandAction{}, nil
}

var (
	_ action.Action = (*sendCommandAction)(nil)
)

type sendCommandAction struct {
	framework.ActionWithModel[sendCommandModel]
}

type sendCommandModel struct {
	framework.WithRegionModel
	InstanceIds    fwtypes.List   `tfsdk:"instance_ids"`
	DocumentName   fwtypes.String `tfsdk:"document_name"`
	Parameters     fwtypes.Map    `tfsdk:"parameters"`
	OutputS3Bucket fwtypes.String `tfsdk:"output_s3_bucket"`
	Timeout        fwtypes.Int64  `tfsdk:"timeout"`
}

func (a *sendCommandAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Executes commands on EC2 instances using AWS Systems Manager Run Command.",
		Attributes: map[string]schema.Attribute{
			"instance_ids": schema.ListAttribute{
				Description: "List of EC2 instance IDs to execute the command on",
				Required:    true,
				ElementType: fwtypes.StringType,
			},
			"document_name": schema.StringAttribute{
				Description: "Name of the SSM document to execute",
				Required:    true,
			},
			"parameters": schema.MapAttribute{
				Description: "Parameters to pass to the command document",
				Optional:    true,
				ElementType: fwtypes.ListType{ElemType: fwtypes.StringType},
			},
			"output_s3_bucket": schema.StringAttribute{
				Description: "S3 bucket name to store command output",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for command execution (default: 1800)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 7200),
				},
			},
		},
	}
}

func (a *sendCommandAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config sendCommandModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().SSMClient(ctx)

	var instanceIds []string
	resp.Diagnostics.Append(config.InstanceIds.ElementsAs(ctx, &instanceIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	documentName := config.DocumentName.ValueString()

	// Set default timeout if not provided
	timeout := 1800 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting SSM send command action", map[string]any{
		"instance_ids":    instanceIds,
		"document_name":   documentName,
		names.AttrTimeout: timeout.String(),
	})

	// Validate instances exist
	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Validating instances...",
	})

	err := a.validateInstances(ctx, conn, instanceIds)
	if err != nil {
		resp.Diagnostics.AddError(
			"Instance Validation Failed",
			fmt.Sprintf("Could not validate instances: %s", err),
		)
		return
	}

	// Validate document exists
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Validating document '%s'...", documentName),
	})

	err = a.validateDocument(ctx, conn, documentName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Document Validation Failed",
			fmt.Sprintf("Could not validate document: %s", err),
		)
		return
	}

	// Send initial progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Sending command '%s' to %d instance(s)...", documentName, len(instanceIds)),
	})

	// Prepare SendCommand input
	input := &ssm.SendCommandInput{
		InstanceIds:  instanceIds,
		DocumentName: aws.String(documentName),
	}

	// Add parameters if provided
	if !config.Parameters.IsNull() {
		parameters := make(map[string][]string)
		diags := config.Parameters.ElementsAs(ctx, &parameters, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Parameters = parameters
	}

	// Add S3 bucket if provided
	if !config.OutputS3Bucket.IsNull() {
		input.OutputS3BucketName = aws.String(config.OutputS3Bucket.ValueString())
	}

	// Send command
	output, err := conn.SendCommand(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Send Command",
			fmt.Sprintf("Could not send SSM command: %s", err),
		)
		return
	}

	commandId := aws.ToString(output.Command.CommandId)
	tflog.Debug(ctx, "Command sent", map[string]any{
		"command_id": commandId,
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Command sent (ID: %s), waiting for execution to complete...", commandId),
	})

	// Wait for command to complete on all instances
	err = a.waitForCommandCompletion(ctx, conn, commandId, instanceIds, timeout, resp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Command Execution Failed",
			fmt.Sprintf("Error during command execution: %s", err),
		)
		return
	}

	// Final success message
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Command '%s' completed successfully on all instances", documentName),
	})

	tflog.Info(ctx, "SSM send command action completed successfully", map[string]any{
		"command_id": commandId,
	})
}

// waitForCommandCompletion waits for the command to complete on all instances
func (a *sendCommandAction) waitForCommandCompletion(ctx context.Context, conn *ssm.Client, commandId string, instanceIds []string, timeout time.Duration, resp *action.InvokeResponse) error {
	startTime := time.Now()

	for {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout waiting for command to complete after %s", timeout)
		}

		allComplete, err := a.areAllInvocationsComplete(ctx, conn, commandId, instanceIds)
		if err != nil {
			return err
		}

		if allComplete {
			// Check for failures
			for _, instanceId := range instanceIds {
				invocation, err := a.getCommandInvocationStatus(ctx, conn, commandId, instanceId)
				if err != nil {
					return err
				}

				if invocation.Status != types.CommandInvocationStatusSuccess {
					return fmt.Errorf("command failed on instance %s with status '%s' (exit code: %d): %s",
						instanceId,
						invocation.Status,
						invocation.ResponseCode,
						aws.ToString(invocation.StandardErrorContent))
				}
			}
			return nil
		}

		// Send progress update
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Command is still executing on instances, elapsed time: %s...", time.Since(startTime).Round(time.Second)),
		})

		time.Sleep(sendCommandPollInterval)
	}
}

// getCommandInvocationStatus retrieves the status of a command invocation for a specific instance
func (a *sendCommandAction) getCommandInvocationStatus(ctx context.Context, conn *ssm.Client, commandId, instanceId string) (*ssm.GetCommandInvocationOutput, error) {
	input := &ssm.GetCommandInvocationInput{
		CommandId:  aws.String(commandId),
		InstanceId: aws.String(instanceId),
	}

	output, err := conn.GetCommandInvocation(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("getting command invocation for instance %s: %w", instanceId, err)
	}

	return output, nil
}

// areAllInvocationsComplete checks if all command invocations have completed
func (a *sendCommandAction) areAllInvocationsComplete(ctx context.Context, conn *ssm.Client, commandId string, instanceIds []string) (bool, error) {
	for _, instanceId := range instanceIds {
		invocation, err := a.getCommandInvocationStatus(ctx, conn, commandId, instanceId)
		if err != nil {
			return false, err
		}

		// Check if still in progress
		if invocation.Status == types.CommandInvocationStatusInProgress ||
			invocation.Status == types.CommandInvocationStatusPending {
			return false, nil
		}
	}

	return true, nil
}

// validateInstances checks if all specified instances exist and are managed by SSM
func (a *sendCommandAction) validateInstances(ctx context.Context, conn *ssm.Client, instanceIds []string) error {
	input := &ssm.DescribeInstanceInformationInput{
		Filters: []types.InstanceInformationStringFilter{
			{
				Key:    aws.String("InstanceIds"),
				Values: instanceIds,
			},
		},
	}

	output, err := conn.DescribeInstanceInformation(ctx, input)
	if err != nil {
		return fmt.Errorf("describing instances: %w", err)
	}

	if len(output.InstanceInformationList) != len(instanceIds) {
		foundIds := make(map[string]bool)
		for _, info := range output.InstanceInformationList {
			foundIds[aws.ToString(info.InstanceId)] = true
		}

		var missingIds []string
		for _, id := range instanceIds {
			if !foundIds[id] {
				missingIds = append(missingIds, id)
			}
		}

		return fmt.Errorf("instances not found or not managed by SSM: %v", missingIds)
	}

	return nil
}

// validateDocument checks if the specified document exists
func (a *sendCommandAction) validateDocument(ctx context.Context, conn *ssm.Client, documentName string) error {
	input := &ssm.DescribeDocumentInput{
		Name: aws.String(documentName),
	}

	_, err := conn.DescribeDocument(ctx, input)
	if err != nil {
		return fmt.Errorf("document '%s' does not exist: %w", documentName, err)
	}

	return nil
}
