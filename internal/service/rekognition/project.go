// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Project")
func newResourceProject(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProject{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type resourceProject struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

const (
	ResNameProject = "Project"
)

func (r *resourceProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_project"
}

func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"auto_update": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProjectAutoUpdate](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"feature": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CustomizationFeature](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := rekognition.CreateProjectInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.ProjectName = flex.StringFromFramework(ctx, plan.Name)

	if plan.Feature.ValueEnum() == awstypes.CustomizationFeatureCustomLabels {
		in.AutoUpdate = ""
	}

	out, err := conn.CreateProject(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameProject, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.ProjectArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameProject, plan.Name.ValueString(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.ARN = flex.StringToFramework(ctx, out.ProjectArn)
	state.ID = state.Name

	// API  returns empty string so we set a null
	if state.Feature.ValueEnum() == awstypes.CustomizationFeatureCustomLabels {
		state.AutoUpdate = fwtypes.StringEnumNull[awstypes.ProjectAutoUpdate]()
	}

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitProjectCreated(ctx, conn, state.ID.ValueString(), in.Feature, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForCreation, ResNameProject, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceProjectData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProjectByName(ctx, conn, state.ID.ValueString(), awstypes.CustomizationFeature(state.Feature.ValueString()))

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionReading, ResNameProject, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Name = state.ID
	state.ARN = flex.StringToFramework(ctx, out.ProjectArn)

	if state.Feature.ValueString() == "" {
		// API returns empty string for default CUSTOM_LABELS value, so we have to set it forcibly to avoid drift
		state.Feature = fwtypes.StringEnumValue(awstypes.CustomizationFeatureCustomLabels)
	}

	// API returns empty string for default DISABLED value, so we have to set it forcibly to avoid drift
	if state.AutoUpdate.ValueString() == "" {
		if state.Feature.ValueEnum() == awstypes.CustomizationFeatureCustomLabels {
			state.AutoUpdate = fwtypes.StringEnumNull[awstypes.ProjectAutoUpdate]()
		} else {
			state.AutoUpdate = fwtypes.StringEnumValue(awstypes.ProjectAutoUpdateDisabled)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceProjectData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceProjectData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.DeleteProjectInput{
		ProjectArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteProject(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameProject, state.ID.ValueString(), err),
			err.Error(),
		)
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProjectDeleted(ctx, conn, state.ID.ValueString(), state.Feature.ValueEnum(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForDeletion, ResNameProject, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func waitProjectCreated(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature, timeout time.Duration) (*awstypes.ProjectDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProjectStatusCreating),
		Target:                    enum.Slice(awstypes.ProjectStatusCreated),
		Refresh:                   statusProject(ctx, conn, name, feature),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProjectDescription); ok {
		return out, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature, timeout time.Duration) (*awstypes.ProjectDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProjectStatusDeleting),
		Target:                    []string{},
		Refresh:                   statusProject(ctx, conn, name, feature),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ProjectDescription); ok {
		return out, err
	}

	return nil, err
}

func findProjectByName(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature) (*awstypes.ProjectDescription, error) {
	features := []awstypes.CustomizationFeature{}
	if len((string)(feature)) == 0 {
		// we don't know the type on import, so we lookup both
		features = append(features, awstypes.CustomizationFeatureContentModeration, awstypes.CustomizationFeatureCustomLabels)
	} else {
		features = append(features, feature)
	}

	in := &rekognition.DescribeProjectsInput{
		ProjectNames: []string{
			name,
		},
		Features: features,
	}

	out, err := conn.DescribeProjects(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.ProjectDescriptions) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.ProjectDescriptions[0], nil
}

func statusProject(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findProjectByName(ctx, conn, name, feature)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

type resourceProjectData struct {
	ARN        types.String                                      `tfsdk:"arn"`
	AutoUpdate fwtypes.StringEnum[awstypes.ProjectAutoUpdate]    `tfsdk:"auto_update"`
	Feature    fwtypes.StringEnum[awstypes.CustomizationFeature] `tfsdk:"feature"`
	ID         types.String                                      `tfsdk:"id"`
	Name       types.String                                      `tfsdk:"name"`
	Timeouts   timeouts.Value                                    `tfsdk:"timeouts"`
}
