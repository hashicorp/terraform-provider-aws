// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceAccountRegistration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAccountRegistration{}, nil
}

const (
	ResNameAccountRegistration = "AccountRegistration"
)

type resourceAccountRegistration struct {
	framework.ResourceWithConfigure
}

func (r *resourceAccountRegistration) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_account_registration"
}

func (r *resourceAccountRegistration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"delegated_admin_account": schema.StringAttribute{
				Optional: true,
			},
			"deregister_on_destroy": schema.BoolAttribute{
				Optional: true,
			},
			names.AttrKMSKey: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceAccountRegistration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)
	// Registration is applied per region, so use this as the ID
	id := r.Meta().Region

	var plan resourceAccountRegistrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.RegisterAccountInput{}
	if !plan.DelegatedAdminAccount.IsNull() {
		in.DelegatedAdminAccount = aws.String(plan.DelegatedAdminAccount.ValueString())
	}
	if !plan.KmsKey.IsNull() {
		in.KmsKey = aws.String(plan.KmsKey.ValueString())
	}
	out, err := conn.RegisterAccount(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAccountRegistration, id, nil),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = types.StringValue(id)
	state.Status = flex.StringValueToFramework(ctx, out.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAccountRegistration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAccountRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// There is no API to get account registration attributes like delegated admin account
	// and KMS key. Read will instead call the GetAccountStatus API to confirm an active
	// account status.
	out, err := conn.GetAccountStatus(ctx, &auditmanager.GetAccountStatusInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameAccountRegistration, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	if out.Status == awstypes.AccountStatusInactive {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = flex.StringValueToFramework(ctx, out.Status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccountRegistration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan, state resourceAccountRegistrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.DelegatedAdminAccount.Equal(state.DelegatedAdminAccount) ||
		!plan.KmsKey.Equal(state.KmsKey) {
		in := auditmanager.RegisterAccountInput{}
		if !plan.DelegatedAdminAccount.IsNull() {
			in.DelegatedAdminAccount = aws.String(plan.DelegatedAdminAccount.ValueString())
		}
		if !plan.KmsKey.IsNull() {
			in.KmsKey = aws.String(plan.KmsKey.ValueString())
		}
		out, err := conn.RegisterAccount(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionUpdating, ResNameAccountRegistration, state.ID.String(), nil),
				err.Error(),
			)
			return
		}

		state.DelegatedAdminAccount = plan.DelegatedAdminAccount
		state.KmsKey = plan.KmsKey
		state.Status = flex.StringValueToFramework(ctx, out.Status)
	}

	if !plan.DeregisterOnDestroy.Equal(state.DeregisterOnDestroy) {
		state.DeregisterOnDestroy = plan.DeregisterOnDestroy
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAccountRegistration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAccountRegistrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeregisterOnDestroy.ValueBool() {
		_, err := conn.DeregisterAccount(ctx, &auditmanager.DeregisterAccountInput{})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameAccountRegistration, state.ID.String(), nil),
				err.Error(),
			)
		}
	}
}

func (r *resourceAccountRegistration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

type resourceAccountRegistrationData struct {
	DelegatedAdminAccount types.String `tfsdk:"delegated_admin_account"`
	DeregisterOnDestroy   types.Bool   `tfsdk:"deregister_on_destroy"`
	KmsKey                types.String `tfsdk:"kms_key"`
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
}
