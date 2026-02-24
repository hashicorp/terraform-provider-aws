// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
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
	framework.WithRegionModel
	ProjectName                  types.String                                              `tfsdk:"project_name"`
	SourceVersion                types.String                                              `tfsdk:"source_version"`
	Timeout                      types.Int64                                               `tfsdk:"timeout"`
	EnvironmentVariablesOverride fwtypes.ListNestedObjectValueOf[environmentVariableModel] `tfsdk:"environment_variables_override"`
	BuildID                      types.String                                              `tfsdk:"build_id"`
}

type environmentVariableModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Type  types.String `tfsdk:"type"`
}

func (a *startBuildAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts a CodeBuild project build. This action is synchronous and waits for the build to complete. When using with action_trigger lifecycle events, use before_create to ensure dependent resources wait for build artifacts.",
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

	tflog.Info(ctx, "Starting CodeBuild project build", map[string]any{
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

	buildID := aws.ToString(output.Build.Id)
	model.BuildID = types.StringValue(buildID)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Build started, waiting for completion...",
	})

	// Poll for build completion using actionwait with backoff strategy
	// Use backoff since builds can take a long time and status changes less frequently
	// as the build progresses - start with frequent polling then back off
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.Build], error) {
		input := codebuild.BatchGetBuildsInput{Ids: []string{buildID}}
		batch, berr := conn.BatchGetBuilds(ctx, &input)
		if berr != nil {
			return actionwait.FetchResult[*awstypes.Build]{}, berr
		}
		if len(batch.Builds) == 0 {
			return actionwait.FetchResult[*awstypes.Build]{}, fmt.Errorf("build not found in BatchGetBuilds response")
		}
		b := batch.Builds[0]
		return actionwait.FetchResult[*awstypes.Build]{Status: actionwait.Status(b.BuildStatus), Value: &b}, nil
	}, actionwait.Options[*awstypes.Build]{
		Timeout:          timeout,
		Interval:         actionwait.WithBackoffDelay(backoff.DefaultSDKv2HelperRetryCompatibleDelay()),
		ProgressInterval: 2 * time.Minute,
		SuccessStates:    []actionwait.Status{actionwait.Status(awstypes.StatusTypeSucceeded)},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(awstypes.StatusTypeInProgress),
		},
		FailureStates: []actionwait.Status{
			actionwait.Status(awstypes.StatusTypeFailed),
			actionwait.Status(awstypes.StatusTypeFault),
			actionwait.Status(awstypes.StatusTypeStopped),
			actionwait.Status(awstypes.StatusTypeTimedOut),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{Message: "Build currently in state: " + string(fr.Status)})
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError("Build timeout", "Build did not complete within the specified timeout")
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError("Build failed", "Build completed with status: "+err.Error())
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError("Unexpected build status", err.Error())
		} else {
			resp.Diagnostics.AddError("Error waiting for build", err.Error())
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{Message: "Build completed successfully"})
}
