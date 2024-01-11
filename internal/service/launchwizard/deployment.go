// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package launchwizard

import (
	"context"
	"errors"
	"regexp"
	"time"

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
						regexp.MustCompile(`^[A-Za-z0-9_\s\.-]+$`),
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
					// mapplanmodifier.RequiresReplace(),
					mapplanmodifier.RequiresReplaceIf(specificationRequiresReplaceIf, "Specifications", "Specifications"),
				},
				Validators:  []validator.Map{},
				Description: "Specifications",
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

// func specificationRequiresReplaceIf(ctx context.Context, req planmodifier.MapRequest, resp *mapplanmodifier.RequiresReplaceIfFuncResponse) {
// 	resp.RequiresReplace = true

// 	//get current state
// 	var state resourceDeploymentData
// 	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
// 	if resp.Diagnostics.HasError() {
// 		return
// 	}

// 	spec_state := flex.ExpandFrameworkStringMap(ctx, state.Specifications)
// 	*spec_state["SaveDeploymentArtifacts"] = "No"

// 	// logging.Log(ctx, logging.Debug, "spec_state: %v", spec_state)
// 	// logging.Log(ctx, logging.Debug, "req.PlanValue: %v", req.PlanValue)

// 	println("spec_state: ", spec_state)
// 	// println("req.PlanValue: ", req.PlanValue)

// 	//compare state with config
// 	// var config resourceDeploymentData
// 	//spec_config := flex.ExpandFrameworkStringMap(ctx, req.ConfigValue.Elements()["specifications"].Elements())
	
// 	// resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)


// 	//req.ConfigValue.Get(ctx, "SaveDeploymentArtifacts", func(v types.Value) {
	
// }

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

	out, err := conn.CreateDeployment(ctx, dry_run_in)
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

	out, err = conn.CreateDeployment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionCreating, ResNameDeployment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DeploymentId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionCreating, ResNameDeployment, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.DeploymentId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	wait_out, err := waitDeploymentCreated(ctx, conn, plan.ID.ValueString(), createTimeout)

	plan.ResourceGroup = flex.StringToFramework(ctx, wait_out.ResourceGroup)
	plan.Status = flex.StringToFramework(ctx, (*string)(&wait_out.Status))

	// spec_temp := flex.ExpandFrameworkStringMap(ctx, flex.FlattenFrameworkStringValueMap(ctx, wait_out.Specifications))

	// //workaround as specification is not returned properly by Get API. TODO: Ask AWS to fix the API
	// if *spec_temp["SaveDeploymentArtifacts"] == "true" {
	// 	*spec_temp["SaveDeploymentArtifacts"] = "Yes"
	// } else {
	// 	*spec_temp["SaveDeploymentArtifacts"] = "No"
	// }

	// //bug in API; remove "deploymentScenario" from specifications
	// delete(spec_temp, "deploymentScenario")


	// plan.Specifications = flex.FlattenFrameworkStringMap(ctx, spec_temp)	

	// //the password attribute is not returned by the API when conducting a read operation;
	// if current_value, ok := spec_temp["DatabasePassword"]; ok {
	// 	if *current_value != "" {
	// 		*spec_temp["DatabasePassword"] = db_password
	// 	}
	// }

	// sap_password := flex.ExpandFrameworkStringValueMap(ctx, state.Specifications)["SapPassword"]

	// if current_value, ok := spec_temp["SapPassword"]; ok {
	// 	if *current_value != "" {
	// 		*spec_temp["SapPassword"] = sap_password
	// 	}
	// }

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForCreation, ResNameDeployment, plan.Name.String(), err),
			err.Error(),
		)

		//Delete Deployment if creation failed
		in := &launchwizard.DeleteDeploymentInput{
			DeploymentId: aws.String(plan.ID.ValueString()),
		}

		_, err := conn.DeleteDeployment(ctx, in)

		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return
			}
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionDeleting, ResNameDeployment, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		deleteTimeout := r.DeleteTimeout(ctx, plan.Timeouts)
		_, err = waitDeploymentDeleted(ctx, conn, plan.ID.ValueString(), deleteTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForDeletion, ResNameDeployment, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// func (r *resourceDeployment) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
// 	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
// 		var old, new resourceDeploymentData

// 		response.Diagnostics.Append(request.State.Get(ctx, &old)...)

// 		if response.Diagnostics.HasError() {
// 			return
// 		}

// 		spec_temp := flex.ExpandFrameworkStringMap(ctx, old.Specifications)

// 		//the password attribute is returned as "******" by the API when conducting a read operation;
// 		// ensure those values do not cause a replacement

// 		if current_value, ok := spec_temp["DatabasePassword"]; ok {
// 			if *current_value != "" {
// 				*spec_temp["DatabasePassword"] = "******"
// 			}
// 		}

// 		if current_value, ok := spec_temp["SapPassword"]; ok {
// 			if *current_value != "" {
// 				*spec_temp["SapPassword"] = "******"
// 			}
// 		}

// 		new.Specifications = flex.FlattenFrameworkStringMap(ctx, spec_temp)

// 		response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

// 		if response.Diagnostics.HasError() {
// 			return
// 		}

// 		// // When you modify a rule, you cannot change the rule's source type.
// 		// if new, old := new.sourceAttributeName(), old.sourceAttributeName(); new != old {
// 		// 	response.RequiresReplace = []path.Path{path.Root(old), path.Root(new)}
// 		// }
// 	}
// }

func (r *resourceDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LaunchWizardClient(ctx)

	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDeploymentByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionSetting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.DeploymentPatternName = flex.StringToFramework(ctx, out.PatternName)
	state.WorkloadName = flex.StringToFramework(ctx, out.WorkloadName)

	spec_temp := flex.ExpandFrameworkStringMap(ctx, flex.FlattenFrameworkStringValueMap(ctx, out.Specifications))

	//workaround as specification is not returned properly by Get API. TODO: Ask AWS to fix the API
	if *spec_temp["SaveDeploymentArtifacts"] == "true" {
		*spec_temp["SaveDeploymentArtifacts"] = "Yes"
	} else {
		*spec_temp["SaveDeploymentArtifacts"] = "No"
	}

	//bug in API; remove "deploymentScenario" from specifications
	delete(spec_temp, "deploymentScenario")

	//the password attribute is not returned by the API when conducting a read operation;
	db_password := flex.ExpandFrameworkStringValueMap(ctx, state.Specifications)["DatabasePassword"]

	if current_value, ok := spec_temp["DatabasePassword"]; ok {
		if *current_value != "" {
			*spec_temp["DatabasePassword"] = db_password
		}
	}

	sap_password := flex.ExpandFrameworkStringValueMap(ctx, state.Specifications)["SapPassword"]

	if current_value, ok := spec_temp["SapPassword"]; ok {
		if *current_value != "" {
			*spec_temp["SapPassword"] = sap_password
		}
	}

	state.Specifications = flex.FlattenFrameworkStringMap(ctx, spec_temp)

	state.ResourceGroup = flex.StringToFramework(ctx, out.ResourceGroup)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDeployment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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
	if tfresource.NotFound(err) || deployment.Status == awstypes.DeploymentStatusDeleted {
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionDeleting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDeploymentDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.LaunchWizard, create.ErrActionWaitingForDeletion, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitDeploymentCreated(ctx context.Context, conn *launchwizard.Client, id string, timeout time.Duration) (*awstypes.DeploymentData, error) {
	// stateConf := &retry.StateChangeConf{
	// 	Pending:                   []string{string(awstypes.DeploymentStatusCreating), string(awstypes.DeploymentStatusInProgress), string(awstypes.DeploymentStatusValidating)},
	// 	Target:                    []string{string(awstypes.DeploymentStatusCompleted)},
	// 	Refresh:                   statusDeployment(ctx, conn, id),
	// 	Timeout:                   timeout,
	// 	NotFoundChecks:            20,
	// 	ContinuousTargetOccurence: 2,
	// }
	stateConf := &retry.StateChangeConf{ //Used during testing. TODO: Remove
		Pending:                   []string{string(awstypes.DeploymentStatusCreating), string(awstypes.DeploymentStatusValidating)},
		Target:                    []string{string(awstypes.DeploymentStatusCompleted), string(awstypes.DeploymentStatusInProgress)},
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
		Pending: []string{string(awstypes.DeploymentStatusDeleteInProgress), string(awstypes.DeploymentStatusDeleteInitiating), string(awstypes.DeploymentStatusInProgress)},
		Target:  []string{string(awstypes.DeploymentStatusDeleted), string(awstypes.DeploymentStatusCompleted)},
		Refresh: statusDeployment(ctx, conn, id),
		Timeout: timeout,
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

type resourceDeploymentData struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	DeploymentPatternName types.String `tfsdk:"deployment_pattern"`
	WorkloadName          types.String `tfsdk:"workload_name"`
	Specifications        types.Map    `tfsdk:"specifications"`
	ResourceGroup         types.String `tfsdk:"resource_group"`
	Status                types.String `tfsdk:"status"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
