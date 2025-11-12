// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @Action(aws_ssoadmin_rename_application, name="Rename Application")
func newRenameApplicationAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &renameApplicationAction{}, nil
}

var (
	_ action.Action = (*renameApplicationAction)(nil)
)

type renameApplicationAction struct {
	framework.ActionWithModel[renameApplicationActionModel]
}

type renameApplicationActionModel struct {
	framework.WithRegionModel
	ApplicationArn types.String `tfsdk:"application_arn"`
	NewName        types.String `tfsdk:"new_name"`
}

func (a *renameApplicationAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Renames an AWS SSO Admin application. This action allows for imperative renaming of SSO applications.",
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Description: "The ARN of the SSO application to rename.",
				Required:    true,
			},
			"new_name": schema.StringAttribute{
				Description: "The new name for the SSO application.",
				Required:    true,
			},
		},
	}
}

func (a *renameApplicationAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config renameApplicationActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().SSOAdminClient(ctx)

	applicationArn := config.ApplicationArn.ValueString()
	newName := config.NewName.ValueString()

	tflog.Info(ctx, "Starting SSO Admin rename application action", map[string]any{
		"application_arn": applicationArn,
		"new_name":        newName,
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Renaming SSO application %s to %s...", applicationArn, newName),
	})

	// Get current application details
	describeInput := &ssoadmin.DescribeApplicationInput{
		ApplicationArn: aws.String(applicationArn),
	}

	describeOutput, err := conn.DescribeApplication(ctx, describeInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Describe SSO Application",
			fmt.Sprintf("Could not describe SSO application %s: %s", applicationArn, err),
		)
		return
	}

	// Update application with new name
	updateInput := &ssoadmin.UpdateApplicationInput{
		ApplicationArn: aws.String(applicationArn),
		Name:           aws.String(newName),
		Description:    describeOutput.Description,
		Status:         describeOutput.Status,
	}

	_, err = conn.UpdateApplication(ctx, updateInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Rename SSO Application",
			fmt.Sprintf("Could not rename SSO application %s: %s", applicationArn, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("SSO application %s renamed to %s successfully", applicationArn, newName),
	})

	tflog.Info(ctx, "SSO Admin rename application action completed successfully", map[string]any{
		"application_arn": applicationArn,
		"new_name":        newName,
	})
}
