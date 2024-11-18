// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_instance_state", name="RDS Instance State")
func newResourceRDSInstanceState(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceRDSInstanceState{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRDSInstanceState = "RDS Instance State"
)

type resourceRDSInstanceState struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByID
}

func (r *resourceRDSInstanceState) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rds_instance_state"
}

func (r *resourceRDSInstanceState) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Required: true,
				// Validators: []validator.String{
				// 	enum.FrameworkValidate[awstypes.DBInstanceStatusInfo](),
				// },
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceRDSInstanceState) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RDSClient(ctx)

	var plan resourceRDSInstanceStateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := plan.Identifier.ValueString()

	instance, err := waitDBInstanceAvailable(ctx, conn, instanceID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Instance (%s)", instanceID), err.Error())

		return
	}

	if err := updateRDSInstanceState(ctx, conn, instanceID, aws.ToString(instance.DBInstanceStatus), plan.State.ValueString(), r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Instance (%s)", instanceID), err.Error())
	}

	plan.State = flex.StringToFramework(ctx, instance.DBInstanceStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceRDSInstanceState) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RDSClient(ctx)

	var state resourceRDSInstanceStateData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDBInstanceByID(ctx, conn, state.Identifier.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionSetting, ResNameRDSInstanceState, state.Identifier.String(), err),
			err.Error(),
		)
		return
	}

	state.State = flex.StringToFramework(ctx, out.DBInstanceStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRDSInstanceState) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().RDSClient(ctx)

	var plan, state resourceRDSInstanceStateData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := waitDBInstanceAvailable(ctx, conn, state.Identifier.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Instance (%s)", state.Identifier.ValueString()), err.Error())

		return
	}

	if !plan.State.Equal(state.State) {
		if err := updateRDSInstanceState(ctx, conn, state.Identifier.ValueString(), state.State.ValueString(), plan.State.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("waiting for RDS Instance (%s)", state.Identifier.ValueString()), err.Error())
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func updateRDSInstanceState(ctx context.Context, conn *rds.Client, id string, currentState string, configuredState string, timeout time.Duration) error {
	if currentState == configuredState {
		return nil
	}

	if configuredState == "stopped" {
		if err := stopInstance(ctx, conn, id, timeout); err != nil {
			return err
		}
	}

	if configuredState == "available" {
		if err := startInstance(ctx, conn, id, timeout); err != nil {
			return err
		}
	}

	return nil
}

type resourceRDSInstanceStateData struct {
	Identifier types.String   `tfsdk:"identifier"`
	State      types.String   `tfsdk:"state"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
