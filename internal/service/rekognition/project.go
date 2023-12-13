package rekognition

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Project")
func newResourceProject(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProject{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceProject struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

const (
	ResNameProject = "Project"
)

func (r *resourceProject) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_project"
}

func (r *resourceProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_update": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.ProjectAutoUpdate](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"feature": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.CustomizationFeature](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
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

	in := rekognition.CreateProjectInput{
		ProjectName: aws.String(plan.Name.ValueString()),
	}
	if !plan.AutoUpdate.IsNull() {
		in.AutoUpdate = awstypes.ProjectAutoUpdate(plan.AutoUpdate.ValueString())
	}
	if !plan.Feature.IsNull() {
		in.Feature = awstypes.CustomizationFeature(plan.Feature.ValueString())
	}

	out, err := conn.CreateProject(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ProjectArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionCreating, ResNameProject, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.ARN = flex.StringValueToFramework(ctx, *out.ProjectArn)
	state.Id = state.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitProjectCreated(ctx, conn, state.Id.ValueString(), in.Feature, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionWaitingForDeletion, ResNameProject, state.Id.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state resourceProjectData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindProjectByName(ctx, conn, state.Id.ValueString(), awstypes.CustomizationFeature(state.Feature.ValueString()))
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionReading, ResNameProject, state.Id.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.ProjectArn)
	state.AutoUpdate = flex.StringToFramework(ctx, (*string)(&out.AutoUpdate))
	state.Feature = flex.StringToFramework(ctx, (*string)(&out.Feature))

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
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Rekognition, create.ErrActionDeleting, ResNameProject, state.Id.ValueString(), err),
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

func FindProjectByName(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature) (*awstypes.ProjectDescription, error) {
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
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || len(out.ProjectDescriptions) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.ProjectDescriptions[0], nil
}

func statusProject(ctx context.Context, conn *rekognition.Client, name string, feature awstypes.CustomizationFeature) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindProjectByName(ctx, conn, name, feature)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

type resourceProjectData struct {
	ARN        types.String   `tfsdk:"arn"`
	AutoUpdate types.String   `tfsdk:"auto_update"`
	Feature    types.String   `tfsdk:"feature"`
	Id         types.String   `tfsdk:"id"`
	Name       types.String   `tfsdk:"name"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
