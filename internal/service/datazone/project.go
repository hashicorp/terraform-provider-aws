// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package datazone

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_project", name="Project")
func newProjectResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &projectResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameProject = "Project"
)

type projectResource struct {
	framework.ResourceWithModel[projectResourceModel]
	framework.WithTimeouts
}

func (r *projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(2048),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^dzd[-_][a-zA-Z0-9_-]{1,36}$`), "must conform to: ^dzd[-_][a-zA-Z0-9_-]{1,36}$ "),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"glossary_terms": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 20),
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_-]{1,36}$`), "must conform to: ^[a-zA-Z0-9_-]{1,36}$ ")),
				},
				Optional: true,
			},

			names.AttrName: schema.StringAttribute{
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w -]+$`), "must conform to: ^[\\w -]+$ "),
					stringvalidator.LengthBetween(1, 64),
				},
				Required: true,
			},
			"created_by": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),

			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"failure_reasons": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsProjectDeletionError](ctx),
				Computed:   true,
			},

			"last_updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"project_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProjectStatus](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_deletion_check": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().DataZoneClient(ctx)

	in := &datazone.CreateProjectInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateProject(ctx, in)
	if resp.Diagnostics.HasError() {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameProject, plan.Name.String(), nil),
			errors.New("failure when creating").Error(),
		)
		return
	}
	if !(out.FailureReasons == nil) && len(out.FailureReasons) > 0 {
		for _, x := range out.FailureReasons {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameProject, plan.Name.String(), nil),
				errors.New("error message: "+*x.Message+" error code: "+*x.Code).Error(),
			)
		}
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitProjectCreated(ctx, conn, plan.DomainIdentifier.ValueString(), plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForCreation, ResNameProject, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DataZoneClient(ctx)
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProjectByID(ctx, conn, state.DomainIdentifier.ValueString(), state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionSetting, ResNameProject, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var plan, state projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Description.Equal(state.Description) || !plan.GlossaryTerms.Equal(state.GlossaryTerms) || !plan.Name.Equal(state.Name) {
		in := &datazone.UpdateProjectInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

		if resp.Diagnostics.HasError() {
			return
		}
		in.Identifier = plan.ID.ValueStringPointer()
		out, err := conn.UpdateProject(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DataZone, create.ErrActionUpdating, ResNameProject, plan.ID.String(), nil),
				errors.New("empty output from project update").Error(),
			)
			return
		}
		out.ProjectStatus = "ACTIVE"
		resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DataZoneClient(ctx)

	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &datazone.DeleteProjectInput{
		DomainIdentifier: state.DomainIdentifier.ValueStringPointer(),
		Identifier:       state.ID.ValueStringPointer(),
	}
	if !state.SkipDeletionCheck.IsNull() {
		in.SkipDeletionCheck = state.SkipDeletionCheck.ValueBoolPointer()
	}

	_, err := conn.DeleteProject(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil && !errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "is already DELETING") {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionDeleting, ResNameProject, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitProjectDeleted(ctx, conn, state.DomainIdentifier.ValueString(), state.ID.ValueString(), deleteTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForDeletion, ResNameProject, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")

	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "DomainIdentifier:Id"`, req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
}

func waitProjectCreated(ctx context.Context, conn *datazone.Client, domain string, identifier string, timeout time.Duration) (*datazone.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ProjectStatusActive),
		Refresh:                   statusProject(conn, domain, identifier),
		Timeout:                   timeout,
		NotFoundChecks:            40,
		ContinuousTargetOccurence: 10,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetProjectOutput); ok {
		return out, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *datazone.Client, domain string, identifier string, timeout time.Duration) (*datazone.GetProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.ProjectStatusDeleting, awstypes.ProjectStatusActive),
		Target:       []string{},
		Refresh:      statusProject(conn, domain, identifier),
		Delay:        5 * time.Second,
		PollInterval: 10 * time.Second,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*datazone.GetProjectOutput); ok {
		return out, err
	}

	return nil, err
}

func statusProject(conn *datazone.Client, domain string, identifier string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findProjectByID(ctx, conn, domain, identifier)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if len(out.FailureReasons) > 0 {
			if err := errors.Join(tfslices.ApplyToAll(out.FailureReasons, func(e awstypes.ProjectDeletionError) error {
				return errors.New(aws.ToString(e.Message))
			})...); err != nil {
				return nil, "", err
			}
		}

		return out, string(out.ProjectStatus), nil
	}
}

func findProjectByID(ctx context.Context, conn *datazone.Client, domain string, identifier string) (*datazone.GetProjectOutput, error) {
	in := &datazone.GetProjectInput{
		DomainIdentifier: aws.String(domain),
		Identifier:       aws.String(identifier),
	}
	out, err := conn.GetProject(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type projectResourceModel struct {
	framework.WithRegionModel
	Description       types.String                                            `tfsdk:"description"`
	DomainIdentifier  types.String                                            `tfsdk:"domain_identifier"`
	Name              types.String                                            `tfsdk:"name"`
	CreatedBy         types.String                                            `tfsdk:"created_by"`
	ID                types.String                                            `tfsdk:"id"`
	CreatedAt         timetypes.RFC3339                                       `tfsdk:"created_at"`
	FailureReasons    fwtypes.ListNestedObjectValueOf[dsProjectDeletionError] `tfsdk:"failure_reasons"`
	LastUpdatedAt     timetypes.RFC3339                                       `tfsdk:"last_updated_at"`
	ProjectStatus     fwtypes.StringEnum[awstypes.ProjectStatus]              `tfsdk:"project_status"`
	Timeouts          timeouts.Value                                          `tfsdk:"timeouts"`
	SkipDeletionCheck types.Bool                                              `tfsdk:"skip_deletion_check"`
	GlossaryTerms     fwtypes.ListValueOf[types.String]                       `tfsdk:"glossary_terms"`
}

type dsProjectDeletionError struct {
	Code    types.String `tfsdk:"code"`
	Message types.String `tfsdk:"message"`
}
