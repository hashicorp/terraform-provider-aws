package rekognition

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Project")
// @Tags(identifierAttribute="arn")
func NewResourceProject(_ context.Context) (resource.ResourceWithConfigure, error) {
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

func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"name": schema.StringAttribute{
				Required: true,
			},
			"auto_update": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.ProjectAutoUpdate](),
				},
			},
			"feature": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.CustomizationFeature](),
				},
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
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	_, err = waitBotCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LexV2Models, create.ErrActionWaitingForDeletion, ResNameBot, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

type resourceProjectData struct {
	ARN        types.String   `tfsdk:"arn"`
	Name       types.String   `tfsdk:"name"`
	AutoUpdate types.String   `tfsdk:"auto_update"`
	Feature    types.String   `tfsdk:"feature"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
