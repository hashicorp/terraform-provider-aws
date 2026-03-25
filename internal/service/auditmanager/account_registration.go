// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_account_registration", name="Account Registration")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(generator=false)
// @Testing(hasExistsFunction=false, checkDestroyNoop=true)
// @Testing(preIdentityVersion="v5.100.0")
func newAccountRegistrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountRegistrationResource{}, nil
}

type accountRegistrationResource struct {
	framework.ResourceWithModel[accountRegistrationResourceModel]
	framework.WithImportByIdentity
}

func (r *accountRegistrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *accountRegistrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, diags := r.registerAccount(ctx, data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx)) // Registration is applied per region, so use this as the ID.
	data.Status = fwflex.StringValueToFramework(ctx, output.Status)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *accountRegistrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	// There is no API to get account registration attributes like delegated admin account
	// and KMS key. Read will instead call the GetAccountStatus API to confirm an active
	// account status.
	output, err := findAccountRegistration(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Account Registration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.Status = fwflex.StringValueToFramework(ctx, output.Status)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountRegistrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old accountRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !new.DelegatedAdminAccount.Equal(old.DelegatedAdminAccount) ||
		!new.KMSKey.Equal(old.KMSKey) {
		output, diags := r.registerAccount(ctx, new)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// Set values for unknowns.
		new.Status = fwflex.StringValueToFramework(ctx, output.Status)
	} else {
		new.Status = old.Status
	}

	new.DeregisterOnDestroy = old.DeregisterOnDestroy

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accountRegistrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accountRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	if data.DeregisterOnDestroy.ValueBool() {
		input := auditmanager.DeregisterAccountInput{}
		_, err := conn.DeregisterAccount(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("deregistering Audit Manager Account (%s)", data.ID.ValueString()), err.Error())

			return
		}
	}
}

func (r *accountRegistrationResource) registerAccount(ctx context.Context, data accountRegistrationResourceModel) (*auditmanager.GetAccountStatusOutput, diag.Diagnostics) {
	var diags diag.Diagnostics

	var input auditmanager.RegisterAccountInput
	diags.Append(fwflex.Expand(ctx, data, &input)...)
	if diags.HasError() {
		return nil, diags
	}

	conn := r.Meta().AuditManagerClient(ctx)
	id := r.Meta().Region(ctx)

	_, err := conn.RegisterAccount(ctx, &input)

	if err != nil {
		diags.AddError(fmt.Sprintf("registering Audit Manager Account (%s)", id), err.Error())

		return nil, diags
	}

	output, err := waitAccountRegistered(ctx, conn)

	if err != nil {
		diags.AddError(fmt.Sprintf("waiting for Audit Manager Account (%s) registered", id), err.Error())

		return nil, diags
	}

	return output, diags
}

func findAccountRegistration(ctx context.Context, conn *auditmanager.Client) (*auditmanager.GetAccountStatusOutput, error) {
	input := auditmanager.GetAccountStatusInput{}
	output, err := conn.GetAccountStatus(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if status := output.Status; status == awstypes.AccountStatusInactive {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func statusAccountRegistration(conn *auditmanager.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAccountRegistration(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAccountRegistered(ctx context.Context, conn *auditmanager.Client) (*auditmanager.GetAccountStatusOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AccountStatusPendingActivation),
		Target:  enum.Slice(awstypes.AccountStatusActive),
		Refresh: statusAccountRegistration(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*auditmanager.GetAccountStatusOutput); ok {
		return output, err
	}

	return nil, err
}

type accountRegistrationResourceModel struct {
	framework.WithRegionModel
	DelegatedAdminAccount types.String `tfsdk:"delegated_admin_account"`
	DeregisterOnDestroy   types.Bool   `tfsdk:"deregister_on_destroy"`
	KMSKey                types.String `tfsdk:"kms_key"`
	ID                    types.String `tfsdk:"id"`
	Status                types.String `tfsdk:"status"`
}
