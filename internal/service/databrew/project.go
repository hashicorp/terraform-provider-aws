// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databrew"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databrew/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_databrew_project", name="Project")
func newResourceProject(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProject{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameProject = "Project"
)

type resourceProject struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_databrew_project"
}

func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"name": schema.StringAttribute{
				Required: true,
				// PlanModifiers: []planmodifier.String{
				// 	stringplanmodifier.RequiresReplace(),
				// },
			},
			"dataset_name": schema.StringAttribute{
				Required: true,
			},
			"recipe_name": schema.StringAttribute{
				Required: true,
			},
			"sample": schema.Int64Attribute{
				Optional: true,
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "Role ARN for the project to use in Glue DataBrew",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var plan resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &databrew.CreateProjectInput{
		Name:        aws.String(plan.Name.ValueString()),
		DatasetName: aws.String(plan.DatasetName.ValueString()),
		RecipeName:  aws.String(plan.RecipeName.ValueString()),
	}

	// if !plan.Sample.IsNull() {
	// 	in.Sample = aws.Int64(plan.Sample.ValueInt64())
	// }

	out, err := conn.CreateProject(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionCreating, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Name == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionCreating, ResNameProject, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.Name = flex.StringToFramework(ctx, out.Name)
	plan.DatasetName = flex.StringToFramework(ctx, in.DatasetName)
	plan.RecipeName = flex.StringToFramework(ctx, in.RecipeName)
	// plan.Sample = flex.Int64ToFramework(ctx, out.Project.Sample)
	plan.RoleArn = flex.StringToFramework(ctx, in.RoleArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitProjectCreated(ctx, conn, plan.Name.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionWaitingForCreation, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 5. Set the arguments and attributes
	// 6. Set the state
	conn := r.Meta().DataBrewClient(ctx)

	var state resourceProjectData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProjectByName(ctx, conn, state.Name.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionSetting, ResNameProject, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	state.DatasetName = flex.StringToFramework(ctx, out.DatasetName)
	state.RecipeName = flex.StringToFramework(ctx, out.RecipeName)
	// state.Sample = flex.Int64ToFramework(ctx, out.Sample)
	state.RoleArn = flex.StringToFramework(ctx, out.RoleArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var plan, state resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RoleArn.Equal(state.RoleArn) ||
		!plan.RecipeName.Equal(state.RecipeName) ||
		!plan.DatasetName.Equal(state.DatasetName) ||
		!plan.Sample.Equal(state.Sample) ||
		!!plan.Name.Equal(state.Name) {

		in := &databrew.UpdateProjectInput{
			Name: aws.String(plan.Name.ValueString()),
			// RecipeName:  aws.String(plan.RecipeName.ValueString()),
			// DatasetName: aws.String(plan.DatasetName.ValueString()),
			// Sample:      aws.Int64(plan.Sample.ValueInt64()),
			RoleArn: aws.String(plan.RoleArn.ValueString()),
		}

		// if !plan.Sample.IsNull() {
		// 	in.Sample = aws.Int64(plan.Sample.ValueInt64())
		// }

		out, err := conn.UpdateProject(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameProject, plan.Name.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Name == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataBrew, create.ErrActionUpdating, ResNameProject, plan.Name.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.Name = flex.StringToFramework(ctx, out.Name)
		// TODO: Implement Sample subtype
		// plan.Sample
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitProjectUpdated(ctx, conn, plan.Name.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionWaitingForUpdate, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataBrewClient(ctx)

	var state resourceProjectData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &databrew.DeleteProjectInput{
		Name: aws.String(state.Name.ValueString()),
	}

	_, err := conn.DeleteProject(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionDeleting, ResNameProject, state.Name.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProjectDeleted(ctx, conn, state.Name.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataBrew, create.ErrActionWaitingForDeletion, ResNameProject, state.Name.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitProjectCreated(ctx context.Context, conn *databrew.Client, id string, timeout time.Duration) (*awstypes.Project, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusProject(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Project); ok {
		return out, err
	}

	return nil, err
}

func waitProjectUpdated(ctx context.Context, conn *databrew.Client, name string, timeout time.Duration) (*awstypes.Project, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusProject(ctx, conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Project); ok {
		return out, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *databrew.Client, id string, timeout time.Duration) (*awstypes.Project, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusProject(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Project); ok {
		return out, err
	}

	return nil, err
}

func statusProject(ctx context.Context, conn *databrew.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findProjectByName(ctx, conn, name)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.SessionStatus), nil
	}
}

func findProjectByName(ctx context.Context, conn *databrew.Client, name string) (*databrew.DescribeProjectOutput, error) {
	in := &databrew.DescribeProjectInput{
		Name: aws.String(name),
	}

	out, err := conn.DescribeProject(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Name == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceProjectData struct {
	Name        types.String   `tfsdk:"name"`
	DatasetName types.String   `tfsdk:"dataset_name"`
	RecipeName  types.String   `tfsdk:"recipe_name"`
	Sample      types.Int64    `tfsdk:"sample"`
	RoleArn     types.String   `tfsdk:"role_arn"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
