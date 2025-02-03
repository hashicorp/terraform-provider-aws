// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_costoptimizationhub_enrollment_status", name="Enrollment Status")
func newResourceEnrollmentStatus(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEnrollmentStatus{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameEnrollmentStatus = "Enrollment Status"
)

type resourceEnrollmentStatus struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceEnrollmentStatus) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.EnrollmentStatus](),
				},
			},
			"include_member_accounts": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func (r *resourceEnrollmentStatus) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceEnrollmentStatusData
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	//Input for UpdateEnrollmentStatus
	input := &costoptimizationhub.UpdateEnrollmentStatusInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Status = awstypes.EnrollmentStatus("Active") // Computed

	conn := r.Meta().CostOptimizationHubClient(ctx)

	out, err := conn.UpdateEnrollmentStatus(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", err),
			err.Error(),
		)
		return
	}

	if out == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", nil),
			errors.New("empty out").Error(),
		)
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	data.Status = fwflex.StringValueToFramework(ctx, aws.ToString(out.Status))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceEnrollmentStatus) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceEnrollmentStatusData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CostOptimizationHubClient(ctx)

	out, err := findEnrollmentStatus(ctx, conn)
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionSetting, ResNameEnrollmentStatus, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	//For this Enrollment resource, The non-existence of this resource will mean status will be "Inactive"
	//So if that is the case, remove the resource from data
	if out != nil && len(out.Items) > 0 && out.Items[0].Status == "Inactive" {
		response.State.RemoveResource(ctx)
		return
	}

	// Set attributes for import.
	// A gratuitous call to Autoflex since status is in out.Items[0].Status.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.Status = fwflex.StringValueToFramework(ctx, out.Items[0].Status)

	// out includes the IncludeMemberAccounts field ATM but it is always nil. Thus, we cannot update state
	// and drift detection is not possible. (However, we can still update if the configuration changes.)
	// If drift detection becomes possible, we can uncomment the following code:

	// data.IncludeMemberAccounts = types.BoolValue(false)
	// if out.IncludeMemberAccounts != nil {
	// 	data.IncludeMemberAccounts = fwflex.BoolToFramework(ctx, out.IncludeMemberAccounts)
	// }

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceEnrollmentStatus) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceEnrollmentStatusData
	response.Diagnostics.Append(request.Plan.Get(ctx, &old)...)
	response.Diagnostics.Append(request.State.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// out includes the IncludeMemberAccounts field ATM but it is always nil. Thus, we cannot update state
	// and drift detection is not possible. However, we can still update if the configuration changes.
	if !old.IncludeMemberAccounts.Equal(new.IncludeMemberAccounts) {
		input := &costoptimizationhub.UpdateEnrollmentStatusInput{
			Status:                awstypes.EnrollmentStatus("Active"),
			IncludeMemberAccounts: old.IncludeMemberAccounts.ValueBoolPointer(),
		}

		conn := r.Meta().CostOptimizationHubClient(ctx)

		out, err := conn.UpdateEnrollmentStatus(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, old.ID.String(), err),
				err.Error(),
			)
			return
		}

		if out == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, old.ID.String(), nil),
				errors.New("empty out").Error(),
			)
			return
		}

		old.ID = new.ID
		old.Status = fwflex.StringValueToFramework(ctx, *out.Status)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &old)...)
}

func (r *resourceEnrollmentStatus) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceEnrollmentStatusData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &costoptimizationhub.UpdateEnrollmentStatusInput{
		Status: awstypes.EnrollmentStatus("Inactive"),
	}

	conn := r.Meta().CostOptimizationHubClient(ctx)

	out, err := conn.UpdateEnrollmentStatus(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", err),
			err.Error(),
		)
		return
	}

	if out == nil || out.Status == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CostOptimizationHub, create.ErrActionCreating, ResNameEnrollmentStatus, "UpdateEnrollmentStatus", nil),
			errors.New("empty out").Error(),
		)
		return
	}
}

func findEnrollmentStatus(ctx context.Context, conn *costoptimizationhub.Client) (*costoptimizationhub.ListEnrollmentStatusesOutput, error) {
	input := &costoptimizationhub.ListEnrollmentStatusesInput{
		IncludeOrganizationInfo: false, //Pass input false to get only this account's status (and not its member accounts)
	}

	out, err := conn.ListEnrollmentStatuses(ctx, input)
	if err != nil {
		return nil, err
	}

	// out includes the IncludeMemberAccounts field ATM but it is always nil

	return out, nil
}

type resourceEnrollmentStatusData struct {
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
	IncludeMemberAccounts types.Bool   `tfsdk:"include_member_accounts"`
}
