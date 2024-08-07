// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_m2_deployment", name="Deployment")
func newDeploymentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deploymentResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

type deploymentResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*deploymentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_m2_deployment"
}

func (r *deploymentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_version": schema.Int64Attribute{
				Required: true,
			},
			"deployment_id": schema.StringAttribute{
				Computed: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_stop": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"start": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *deploymentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data deploymentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	input := &m2.CreateDeploymentInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	output, err := conn.CreateDeployment(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Mainframe Modernization Deployment", err.Error())

		return
	}

	// Set values for unknowns.
	data.DeploymentID = fwflex.StringToFramework(ctx, output.DeploymentId)
	data.setID()

	timeout := r.CreateTimeout(ctx, data.Timeouts)
	if _, err := waitDeploymentCreated(ctx, conn, data.ApplicationID.ValueString(), data.DeploymentID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Deployment (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	if data.Start.ValueBool() {
		applicationID := data.ApplicationID.ValueString()
		if _, err := startApplication(ctx, conn, applicationID, timeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("starting Mainframe Modernization Application (%s)", applicationID), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *deploymentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data deploymentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().M2Client(ctx)

	outputGD, err := findDeploymentByTwoPartKey(ctx, conn, data.ApplicationID.ValueString(), data.DeploymentID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Mainframe Modernization Deployment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGD, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	outputGA, err := findApplicationByID(ctx, conn, data.ApplicationID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Mainframe Modernization Application (%s)", data.ApplicationID.ValueString()), err.Error())

		return
	}

	data.Start = types.BoolValue(outputGA.Status == awstypes.ApplicationLifecycleRunning)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *deploymentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new deploymentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	timeout := r.UpdateTimeout(ctx, new.Timeouts)
	if !new.ApplicationVersion.Equal(old.ApplicationVersion) {
		applicationID := new.ApplicationID.ValueString()

		// Stop the application if it was running.
		if old.Start.ValueBool() {
			if _, err := stopApplicationIfRunning(ctx, conn, applicationID, new.ForceStop.ValueBool(), timeout); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("stopping Mainframe Modernization Application (%s)", applicationID), err.Error())

				return
			}
		}

		input := &m2.CreateDeploymentInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		output, err := conn.CreateDeployment(ctx, input)

		if err != nil {
			response.Diagnostics.AddError("creating Mainframe Modernization Deployment", err.Error())

			return
		}

		// Set values for unknowns.
		new.DeploymentID = fwflex.StringToFramework(ctx, output.DeploymentId)
		new.setID()

		if _, err := waitDeploymentUpdated(ctx, conn, new.ApplicationID.ValueString(), new.DeploymentID.ValueString(), timeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Deployment (%s) update", new.ID.ValueString()), err.Error())

			return
		}

		// Start the application if plan says to.
		if new.Start.ValueBool() {
			applicationID := new.ApplicationID.ValueString()
			if _, err := startApplication(ctx, conn, applicationID, timeout); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("starting Mainframe Modernization Application (%s)", applicationID), err.Error())
				return
			}
		}

		response.Diagnostics.Append(response.State.Set(ctx, new)...)
		return
	}

	// Start/stop deployment if no other update is needed
	if !old.Start.Equal(new.Start) {
		applicationID := new.ApplicationID.ValueString()
		new.DeploymentID = old.DeploymentID
		if new.Start.ValueBool() {
			if _, err := startApplication(ctx, conn, applicationID, timeout); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("starting Mainframe Modernization Application (%s)", applicationID), err.Error())
			}
		} else {
			if _, err := stopApplicationIfRunning(ctx, conn, applicationID, new.ForceStop.ValueBool(), timeout); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("stopping Mainframe Modernization Application (%s)", applicationID), err.Error())
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, new)...)
}

func (r *deploymentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data deploymentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().M2Client(ctx)

	timeout := r.DeleteTimeout(ctx, data.Timeouts)
	if data.Start.ValueBool() {
		applicationID := data.ApplicationID.ValueString()
		if _, err := stopApplicationIfRunning(ctx, conn, applicationID, data.ForceStop.ValueBool(), timeout); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("stopping Mainframe Modernization Application (%s)", applicationID), err.Error())

			return
		}
	}

	_, err := conn.DeleteApplicationFromEnvironment(ctx, &m2.DeleteApplicationFromEnvironmentInput{
		ApplicationId: fwflex.StringFromFramework(ctx, data.ApplicationID),
		EnvironmentId: fwflex.StringFromFramework(ctx, data.EnvironmentID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Mainframe Modernization Deployment (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitApplicationDeletedFromEnvironment(ctx, conn, data.ApplicationID.ValueString(), timeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Mainframe Modernization Deployment (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *deploymentResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.State.Raw.IsNull() && !request.Plan.Raw.IsNull() {
		var plan, state deploymentResourceModel
		response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(request.State.Get(ctx, &state)...)
		if response.Diagnostics.HasError() {
			return
		}

		if !plan.ApplicationVersion.Equal(state.ApplicationVersion) {
			// If the ApplicationVersion changes, ID becomes unknown.
			plan.ID = types.StringUnknown()
		}

		response.Diagnostics.Append(response.Plan.Set(ctx, &plan)...)
	}
}

func findDeploymentByTwoPartKey(ctx context.Context, conn *m2.Client, applicationID, deploymentID string) (*m2.GetDeploymentOutput, error) {
	input := &m2.GetDeploymentInput{
		ApplicationId: aws.String(applicationID),
		DeploymentId:  aws.String(deploymentID),
	}

	output, err := conn.GetDeployment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusDeployment(ctx context.Context, conn *m2.Client, applicationID, deploymentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDeploymentByTwoPartKey(ctx, conn, applicationID, deploymentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDeploymentCreated(ctx context.Context, conn *m2.Client, applicationID, deploymentID string, timeout time.Duration) (*m2.GetDeploymentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DeploymentLifecycleDeploying),
		Target:  enum.Slice(awstypes.DeploymentLifecycleSucceeded),
		Refresh: statusDeployment(ctx, conn, applicationID, deploymentID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetDeploymentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitDeploymentUpdated(ctx context.Context, conn *m2.Client, applicationID, deploymentID string, timeout time.Duration) (*m2.GetDeploymentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DeploymentLifecycleDeployUpdate),
		Target:  enum.Slice(awstypes.DeploymentLifecycleSucceeded),
		Refresh: statusDeployment(ctx, conn, applicationID, deploymentID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*m2.GetDeploymentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type deploymentResourceModel struct {
	ApplicationID      types.String   `tfsdk:"application_id"`
	ApplicationVersion types.Int64    `tfsdk:"application_version"`
	DeploymentID       types.String   `tfsdk:"deployment_id"`
	EnvironmentID      types.String   `tfsdk:"environment_id"`
	ForceStop          types.Bool     `tfsdk:"force_stop"`
	ID                 types.String   `tfsdk:"id"`
	Start              types.Bool     `tfsdk:"start"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

const (
	deploymentResourceIDPartCount = 2
)

func (data *deploymentResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, deploymentResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.ApplicationID = types.StringValue(parts[0])
	data.DeploymentID = types.StringValue(parts[1])

	return nil
}

func (data *deploymentResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationID.ValueString(), data.DeploymentID.ValueString()}, deploymentResourceIDPartCount, false)))
}
