// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package rekognition

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rekognition_project", name="Project")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("name", identityDuplicateAttributes="id")
// @Testing(preIdentityVersion="v6.56.0")
func newProjectResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &projectResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type projectResource struct {
	framework.ResourceWithModel[projectResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

const (
	ResNameProject = "Project"
)

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	in := rekognition.CreateProjectInput{
		Tags: getTagsIn(ctx),
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &in, flex.WithFieldNamePrefix("Project")))
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Feature.IsNull() || plan.Feature.ValueEnum() == awstypes.CustomizationFeatureCustomLabels {
		in.AutoUpdate = ""
	}

	out, err := conn.CreateProject(ctx, &in)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	if out == nil || out.ProjectArn == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.ValueString())
		return
	}

	plan.setID()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitProjectCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrName), plan.Name.ValueString())
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &plan, flex.WithFieldNamePrefix("Project")))
	if resp.Diagnostics.HasError() {
		return
	}

	// set default since API does not return it
	if plan.Feature.IsNull() {
		plan.Feature = fwtypes.StringEnumValue(awstypes.CustomizationFeatureCustomLabels)
	}

	// API  returns empty string so we set a null
	if plan.Feature.ValueEnum() == awstypes.CustomizationFeatureCustomLabels {
		plan.AutoUpdate = fwtypes.StringEnumNull[awstypes.ProjectAutoUpdate]()
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state projectResourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.setID()

	out, err := findProjectByName(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Project")))
	if resp.Diagnostics.HasError() {
		return
	}

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

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var state projectResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rekognition.DeleteProjectInput{
		ProjectArn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteProject(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProjectDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

func (r *projectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var feature fwtypes.StringEnum[awstypes.CustomizationFeature]
	var autoUpdate fwtypes.StringEnum[awstypes.ProjectAutoUpdate]

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.GetAttribute(ctx, path.Root("feature"), &feature))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.GetAttribute(ctx, path.Root("auto_update"), &autoUpdate))
	if resp.Diagnostics.HasError() {
		return
	}

	if feature.ValueEnum() == awstypes.CustomizationFeatureContentModeration && (autoUpdate.IsNull() || autoUpdate.IsUnknown()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("auto_update"),
			"Invalid Auto Update value",
			fmt.Sprintf("`auto_update` must be set when `feature` is %q.", feature),
		)

		return
	}
}

func waitProjectCreated(ctx context.Context, conn *rekognition.Client, name string, timeout time.Duration) (*awstypes.ProjectDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProjectStatusCreating),
		Target:                    enum.Slice(awstypes.ProjectStatusCreated),
		Refresh:                   statusProject(conn, name),
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

func waitProjectDeleted(ctx context.Context, conn *rekognition.Client, name string, timeout time.Duration) (*awstypes.ProjectDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ProjectStatusDeleting),
		Target:                    []string{},
		Refresh:                   statusProject(conn, name),
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

func findProjectByName(ctx context.Context, conn *rekognition.Client, name string) (*awstypes.ProjectDescription, error) {
	input := rekognition.DescribeProjectsInput{
		ProjectNames: []string{
			name,
		},
		Features: awstypes.CustomizationFeature("").Values(), // pass all possible to values to filter since it defaults to CUSTOM_LABELS
	}

	out, err := conn.DescribeProjects(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(out.ProjectDescriptions)
}

func statusProject(conn *rekognition.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findProjectByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func (p *projectResourceModel) setID() {
	p.ID = p.Name
}

type projectResourceModel struct {
	framework.WithRegionModel
	ARN        types.String                                      `tfsdk:"arn"`
	AutoUpdate fwtypes.StringEnum[awstypes.ProjectAutoUpdate]    `tfsdk:"auto_update"`
	Feature    fwtypes.StringEnum[awstypes.CustomizationFeature] `tfsdk:"feature" autoflex:"noflatten"`
	ID         types.String                                      `tfsdk:"id"`
	Name       types.String                                      `tfsdk:"name"`
	Tags       tftags.Map                                        `tfsdk:"tags"`
	TagsAll    tftags.Map                                        `tfsdk:"tags_all"`
	Timeouts   timeouts.Value                                    `tfsdk:"timeouts"`
}
