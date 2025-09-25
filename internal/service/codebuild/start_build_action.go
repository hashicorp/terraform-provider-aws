// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_codebuild_start_build, name="CodeBuild Start Build")
func newStartBuildAction(context.Context) (action.ActionWithConfigure, error) {
	return &startBuildAction{}, nil
}

type startBuildAction struct {
	framework.ActionWithModel[startBuildActionModel]
}

type startBuildActionModel struct {
	ProjectName                    types.String                                                `tfsdk:"project_name"`
	SourceVersion                  types.String                                                `tfsdk:"source_version"`
	Timeout                        types.Int64                                                 `tfsdk:"timeout"`
	EnvironmentVariablesOverride   fwtypes.ListNestedObjectValueOf[environmentVariableModel]   `tfsdk:"environment_variables_override"`
	BuildID                        types.String                                                `tfsdk:"build_id"`
}

type environmentVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Type  types.String `tfsdk:"type"`
}

func (a *startBuildAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts a CodeBuild project build",
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Description: "Name of the CodeBuild project",
				Required:    true,
			},
			"source_version": schema.StringAttribute{
				Description: "Version of the build input to be built",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds for the build operation",
				Optional:    true,
			},
			"build_id": schema.StringAttribute{
				Description: "ID of the started build",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"environment_variables_override": schema.ListNestedBlock{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[environmentVariableModel](ctx),
				Description: "Environment variables to override for this build",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Description: "Environment variable name",
							Required:    true,
						},
						names.AttrValue: schema.StringAttribute{
							Description: "Environment variable value",
							Required:    true,
						},
						names.AttrType: schema.StringAttribute{
							Description: "Environment variable type",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (a *startBuildAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var model startBuildActionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().CodeBuildClient(ctx)

	timeout := 30 * time.Minute
	if !model.Timeout.IsNull() {
		timeout = time.Duration(model.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting CodeBuild project build", map[string]interface{}{
		"project_name": model.ProjectName.ValueString(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Starting CodeBuild project build...",
	})

	var input codebuild.StartBuildInput
	resp.Diagnostics.Append(fwflex.Expand(ctx, model, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.StartBuild(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("Starting CodeBuild project build", err.Error())
		return
	}

	buildID := *output.Build.Id
	model.BuildID = types.StringValue(buildID)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Build started, waiting for completion...",
	})

	// Poll for build completion
	deadline := time.Now().Add(timeout)
	pollInterval := 30 * time.Second
	progressInterval := 2 * time.Minute
	lastProgressUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Build monitoring cancelled", "Context was cancelled")
			return
		default:
		}

		if time.Now().After(deadline) {
			resp.Diagnostics.AddError("Build timeout", "Build did not complete within the specified timeout")
			return
		}

		batchGetBuildsOutput, err := conn.BatchGetBuilds(ctx, &codebuild.BatchGetBuildsInput{
			Ids: []string{buildID},
		})
		if err != nil {
			resp.Diagnostics.AddError("Getting build status", err.Error())
			return
		}

		if len(batchGetBuildsOutput.Builds) == 0 {
			resp.Diagnostics.AddError("Build not found", "Build was not found in BatchGetBuilds response")
			return
		}

		build := batchGetBuildsOutput.Builds[0]
		status := build.BuildStatus

		if time.Since(lastProgressUpdate) >= progressInterval {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: "Build currently in state: " + string(status),
			})
			lastProgressUpdate = time.Now()
		}

		switch status {
		case awstypes.StatusTypeSucceeded:
			resp.SendProgress(action.InvokeProgressEvent{
				Message: "Build completed successfully",
			})
			return
		case awstypes.StatusTypeFailed, awstypes.StatusTypeFault, awstypes.StatusTypeStopped, awstypes.StatusTypeTimedOut:
			resp.Diagnostics.AddError("Build failed", "Build completed with status: "+string(status))
			return
		case awstypes.StatusTypeInProgress:
			// Continue polling
		default:
			resp.Diagnostics.AddError("Unexpected build status", "Received unexpected build status: "+string(status))
			return
		}

		time.Sleep(pollInterval)
	}
}
