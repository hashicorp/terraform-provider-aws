// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package launchwizard

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/launchwizard"
	awstypes "github.com/aws/aws-sdk-go-v2/service/launchwizard/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Deployment")
func newResourceDeployment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDeployment{}

	r.SetDefaultCreateTimeout(180 * time.Minute)
	r.SetDefaultUpdateTimeout(180 * time.Minute)
	r.SetDefaultDeleteTimeout(180 * time.Minute)

	return r, nil
}

const (
	ResNameDeployment = "Deployment"
)

type resourceDeployment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_launchwizard_deployment"
}

func (r *resourceDeployment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(50),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[A-Za-z0-9_\s\.-]+$`),
						"Name must be alphanumeric, underscores, spaces, dots, and dashes only. Must be between 1 and 50 characters long.",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the deployment",
			},
			"deployment_pattern": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(256),
				},
				Description: "Deployment Pattern",
			},
			"workload_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(256),
				},
				Description: "Type of Workload",
			},
			"specifications": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplaceIf(
						requiresReplaceUnlessPasswordIsEmpty(),
						"Replace",
						"Replace",
					),
				},
				Validators:  []validator.Map{},
				Description: "Specifications",
			},
			"skip_destroy_after_failure": schema.BoolAttribute{
				Optional:    true,
				Description: "Ensures the deployment doesn't get deleted immediately after a failure. Taints resource.",
			},
			"resource_group": schema.StringAttribute{
				Computed:    true,
				Description: "Resource Group",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status",
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

func (r *resourceDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LaunchWizardClient(ctx)

	var plan resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//dry run for validation
	dry_run_in := &launchwizard.CreateDeploymentInput{
		Name:                  aws.String(plan.Name.ValueString()),
		DeploymentPatternName: aws.String(plan.DeploymentPatternName.ValueString()),
		WorkloadName:          aws.String(plan.WorkloadName.ValueString()),
		Specifications:        flex.ExpandFrameworkStringValueMap(ctx, plan.Specifications),
		DryRun:                true,
	}

	_, err := conn.CreateDeployment(ctx, dry_run_in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionCreating, ResNameDeployment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	//actual deployment
	in := &launchwizard.CreateDeploymentInput{
		Name:                  aws.String(plan.Name.ValueString()),
		DeploymentPatternName: aws.String(plan.DeploymentPatternName.ValueString()),
		WorkloadName:          aws.String(plan.WorkloadName.ValueString()),
		Specifications:        flex.ExpandFrameworkStringValueMap(ctx, plan.Specifications),
	}

	create_out, err := conn.CreateDeployment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionCreating, ResNameDeployment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if create_out == nil || create_out.DeploymentId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionCreating, ResNameDeployment, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, create_out.DeploymentId)

	check_out, err := findDeploymentByID(ctx, conn, plan.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.Name.String(), err),
			err.Error(),
		)
	}

	plan.ResourceGroup = flex.StringToFramework(ctx, check_out.ResourceGroup)
	plan.Status = flex.StringToFramework(ctx, (*string)(&check_out.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	wait_out, err := waitDeploymentCreated(ctx, conn, plan.ID.ValueString(), createTimeout)

	if err != nil {
		if plan.SkipDestroyAfterFailure.ValueBool() {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.Name.String(), err)+"Deployment will be replaced on next apply to allow troubleshooting.",
				err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.Name.String(), err)+"Deployment will be deleted.",
				err.Error(),
			)
		}

		if !plan.SkipDestroyAfterFailure.ValueBool() {
			//Delete Deployment if creation failed
			in := &launchwizard.DeleteDeploymentInput{
				DeploymentId: aws.String(plan.ID.ValueString()),
			}

			_, err := conn.DeleteDeployment(ctx, in)

			if err != nil {
				// if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.ID.String(), err), err.Error())
				// }
			}

			deleteTimeout := r.DeleteTimeout(ctx, plan.Timeouts)
			_, err = waitDeploymentDeleted(ctx, conn, plan.ID.ValueString(), deleteTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.ID.String(), err),
					err.Error(),
				)
			}
		}
	}

	if wait_out != nil {
		plan.Status = flex.StringToFramework(ctx, (*string)(&wait_out.Status))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LaunchWizardClient(ctx)
	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDeploymentByID(ctx, conn, state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionSetting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if tfresource.NotFound(err) || checkDeleted(out.Status) {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.DeploymentPatternName = flex.StringToFramework(ctx, out.PatternName)
	state.WorkloadName = flex.StringToFramework(ctx, out.WorkloadName)
	state.ResourceGroup = flex.StringToFramework(ctx, out.ResourceGroup)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))

	spec_temp := flex.ExpandFrameworkStringMap(ctx, flex.FlattenFrameworkStringValueMap(ctx, out.Specifications))

	//workaround as specification is not returned properly by Get API. TODO: Ask AWS to fix the API
	if spec_temp["SaveDeploymentArtifacts"] == aws.String("true") {
		*spec_temp["SaveDeploymentArtifacts"] = "Yes"
	} else {
		*spec_temp["SaveDeploymentArtifacts"] = "No"
	}

	//bug in API; remove "deploymentScenario" from specifications
	delete(spec_temp, "deploymentScenario")

	//the password attribute is not returned by the API when conducting a read operation;
	db_password := flex.ExpandFrameworkStringValueMap(ctx, state.Specifications)["DatabasePassword"]

	if _, ok := spec_temp["DatabasePassword"]; ok {
		*spec_temp["DatabasePassword"] = db_password
	}

	sap_password := flex.ExpandFrameworkStringValueMap(ctx, state.Specifications)["SapPassword"]

	if _, ok := spec_temp["SapPassword"]; ok {
		*spec_temp["SapPassword"] = sap_password
	}

	state.Specifications = flex.FlattenFrameworkStringMap(ctx, spec_temp)

	//check status as it might be "in progress" (e.g. in case of session timeout)
	switch out.Status {
	case
		awstypes.DeploymentStatusInProgress,
		awstypes.DeploymentStatusCreating,
		awstypes.DeploymentStatusValidating:
		resp.Diagnostics.AddWarning("Deployment still in progress", "Deployment still in progress. Possibly because of previous session timeout.")
	case
		awstypes.DeploymentStatusFailed,
		awstypes.DeploymentStatusDeleteFailed:
		resp.State.RemoveResource(ctx)
		// resp.Diagnostics.AddError("Deployment Failed", "Deployment needs to be replaced.")
		resp.Diagnostics.AddWarning("Deployment Failed", "Deployment needs to be replaced.")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDeployment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	//there is no update API.
	//possible updates:
	//1. password in the state after an import, because the password is not returned by the API when conducting a read operation
	//2. virtual argument "skip_destroy_after_failure" to allow troubleshooting of failed deployments

	//1. show info if user tries to update the password
	if !reflect.DeepEqual(plan.Specifications, state.Specifications) {
		resp.Diagnostics.AddWarning("Password Update", "The API does not support updating the stack, incl. the password. The Update operation only writes the specified password to the state.")
	}

	//2. update virtual argument "skip_destroy_after_failure"
	if plan.SkipDestroyAfterFailure.ValueBool() != state.SkipDestroyAfterFailure.ValueBool() {
		state.SkipDestroyAfterFailure = plan.SkipDestroyAfterFailure
	}

	//pass through resource_group and status
	plan.ResourceGroup = state.ResourceGroup
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDeployment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LaunchWizardClient(ctx)

	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//check if already deleted
	deployment, err := findDeploymentByID(ctx, conn, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionDeleting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if tfresource.NotFound(err) || checkDeleted(deployment.Status) {
		resp.State.RemoveResource(ctx)
		return
	}

	in := &launchwizard.DeleteDeploymentInput{
		DeploymentId: aws.String(state.ID.ValueString()),
	}

	_, err = conn.DeleteDeployment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionDeleting, ResNameDeployment, state.ID.String(), err), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	wait_out, err := waitDeploymentDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForDeletion, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if wait_out.Status != awstypes.DeploymentStatusDeleted {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForDeletion, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	//deployment name is used as the import identifier instead of the deployment id. Deployment Id is not visible from the console
	conn := r.Meta().LaunchWizardClient(ctx)
	in := &launchwizard.ListDeploymentsInput{}
	pages := launchwizard.NewListDeploymentsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionImporting, ResNameDeployment, "", err),
				err.Error(),
			)
			return
		}

		for _, deployment := range page.Deployments {
			if deployment.Name != nil && req.ID == aws.ToString(deployment.Name) {
				//set the import identifier to the deployment id
				req.ID = aws.ToString(deployment.Id)
				resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
				return
			}
		}
	}

	resp.Diagnostics.AddError(create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionImporting, ResNameDeployment, "", nil), errors.New("deployment not found").Error())
}

func waitDeploymentCreated(ctx context.Context, conn *launchwizard.Client, id string, timeout time.Duration) (*awstypes.DeploymentData, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DeploymentStatusCreating, awstypes.DeploymentStatusInProgress, awstypes.DeploymentStatusValidating),
		Target:                    enum.Slice(awstypes.DeploymentStatusCompleted),
		Refresh:                   statusDeployment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DeploymentData); ok {
		return out, err
	}

	return nil, err
}

func waitDeploymentDeleted(ctx context.Context, conn *launchwizard.Client, id string, timeout time.Duration) (*awstypes.DeploymentData, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DeploymentStatusDeleteInProgress, awstypes.DeploymentStatusDeleteInitiating, awstypes.DeploymentStatusInProgress),
		Target:  enum.Slice(awstypes.DeploymentStatusDeleted, awstypes.DeploymentStatusCompleted),
		Refresh: statusDeployment(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DeploymentData); ok {
		return out, err
	}

	return nil, err
}

func statusDeployment(ctx context.Context, conn *launchwizard.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDeploymentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findDeploymentByID(ctx context.Context, conn *launchwizard.Client, id string) (*awstypes.DeploymentData, error) {
	in := &launchwizard.GetDeploymentInput{
		DeploymentId: aws.String(id),
	}

	out, err := conn.GetDeployment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil || out.Deployment == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Deployment, nil
}

func requiresReplaceUnlessPasswordIsEmpty() mapplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
		//drop passwords from both state and config
		//those are not returned by the API when conducting a read operation, so those shouldn't be used to indicate a replacement (as they will always be different when importing a resource)
		var state resourceDeploymentData
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

		if resp.Diagnostics.HasError() {
			return
		}

		spec_state := flex.ExpandFrameworkStringMap(ctx, state.Specifications)
		delete(spec_state, "DatabasePassword")
		delete(spec_state, "SapPassword")

		spec_config := flex.ExpandFrameworkStringMap(ctx, req.ConfigValue)
		delete(spec_config, "DatabasePassword")
		delete(spec_config, "SapPassword")

		//compare state with config
		if reflect.DeepEqual(spec_state, spec_config) {
			resp.RequiresReplace = false
		} else {
			resp.RequiresReplace = true
		}
	}
}

func checkDeleted(status awstypes.DeploymentStatus) bool {
	switch status {
	case awstypes.DeploymentStatusDeleted,
		awstypes.DeploymentStatusDeleteFailed,
		awstypes.DeploymentStatusDeleteInProgress,
		awstypes.DeploymentStatusDeleteInitiating:
		return true
	default:
		return false
	}
}

type resourceDeploymentData struct {
	ID                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	DeploymentPatternName   types.String   `tfsdk:"deployment_pattern"`
	WorkloadName            types.String   `tfsdk:"workload_name"`
	Specifications          types.Map      `tfsdk:"specifications"`
	SkipDestroyAfterFailure types.Bool     `tfsdk:"skip_destroy_after_failure"`
	ResourceGroup           types.String   `tfsdk:"resource_group"`
	Status                  types.String   `tfsdk:"status"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}
