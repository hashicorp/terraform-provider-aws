// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/computeoptimizer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_computeoptimizer_enrollment_status", name="Enrollment Status")
func newEnrollmentStatusResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &enrollmentStatusResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type enrollmentStatusResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByID
}

func (r *enrollmentStatusResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"number_of_member_accounts_opted_in": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("Active", "Inactive"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *enrollmentStatusResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data enrollmentStatusResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	input := &computeoptimizer.UpdateEnrollmentStatusInput{
		IncludeMemberAccounts: fwflex.BoolValueFromFramework(ctx, data.MemberAccountsEnrolled),
		Status:                awstypes.Status(fwflex.StringValueFromFramework(ctx, data.Status)),
	}

	_, err := conn.UpdateEnrollmentStatus(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Compute Optimizer Enrollment Status", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))

	output, err := waitEnrollmentStatusUpdated(ctx, conn, string(input.Status), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Compute Optimizer Enrollment Status (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	data.NumberOfMemberAccountsOptedIn = fwflex.Int32ToFrameworkInt64(ctx, output.NumberOfMemberAccountsOptedIn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *enrollmentStatusResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data enrollmentStatusResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	output, err := findEnrollmentStatus(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Compute Optimizer Enrollment Status (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *enrollmentStatusResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new enrollmentStatusResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ComputeOptimizerClient(ctx)

	input := &computeoptimizer.UpdateEnrollmentStatusInput{
		Status: awstypes.Status(fwflex.StringValueFromFramework(ctx, new.Status)),
	}

	_, err := conn.UpdateEnrollmentStatus(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("updating Compute Optimizer Enrollment Status", err.Error())

		return
	}

	output, err := waitEnrollmentStatusUpdated(ctx, conn, string(input.Status), r.CreateTimeout(ctx, new.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Compute Optimizer Enrollment Status (%s) update", new.ID.ValueString()), err.Error())

		return
	}

	new.NumberOfMemberAccountsOptedIn = fwflex.Int32ToFrameworkInt64(ctx, output.NumberOfMemberAccountsOptedIn)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func findEnrollmentStatus(ctx context.Context, conn *computeoptimizer.Client) (*computeoptimizer.GetEnrollmentStatusOutput, error) {
	input := &computeoptimizer.GetEnrollmentStatusInput{}

	output, err := conn.GetEnrollmentStatus(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusEnrollmentStatus(ctx context.Context, conn *computeoptimizer.Client) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEnrollmentStatus(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEnrollmentStatusUpdated(ctx context.Context, conn *computeoptimizer.Client, targetStatus string, timeout time.Duration) (*computeoptimizer.GetEnrollmentStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusPending),
		Target:  []string{targetStatus},
		Refresh: statusEnrollmentStatus(ctx, conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*computeoptimizer.GetEnrollmentStatusOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type enrollmentStatusResourceModel struct {
	ID                            types.String   `tfsdk:"id"`
	MemberAccountsEnrolled        types.Bool     `tfsdk:"include_member_accounts"`
	NumberOfMemberAccountsOptedIn types.Int64    `tfsdk:"number_of_member_accounts_opted_in"`
	Status                        types.String   `tfsdk:"status"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
}
