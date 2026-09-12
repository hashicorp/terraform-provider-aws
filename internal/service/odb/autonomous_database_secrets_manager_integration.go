// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_autonomous_database_secrets_manager_integration", name="Autonomous Database Secrets Manager Integration")
// @Testing(serialize=true)
func newResourceAutonomousDatabaseSecretsManagerIntegration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAutonomousDatabaseSecretsManagerIntegration{}
	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameAutonomousDatabaseSecretsManagerIntegration = "Autonomous Database Secrets Manager Integration"
	autonomousDatabaseSecretsManagerIntegrationID      = "secrets-manager"
)

type resourceAutonomousDatabaseSecretsManagerIntegration struct {
	framework.ResourceWithModel[autonomousDatabaseSecretsManagerIntegrationResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	statusType := fwtypes.StringEnumType[odbtypes.OciIamRoleStatus]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttribute(),
			names.AttrRoleARN: framework.ARNAttributeComputedOnly(),
			names.AttrStatus: schema.StringAttribute{
				CustomType:  statusType,
				Computed:    true,
				Description: "Current lifecycle status of the Oracle Database@AWS Secrets Manager service role.",
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed:    true,
				Description: "Additional information about the service role lifecycle status.",
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
		Description: "Enables the Oracle Database@AWS Autonomous Database Serverless integration with AWS Secrets Manager.",
	}
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ODBClient(ctx)

	var plan autonomousDatabaseSecretsManagerIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.InitializeServiceInput{
		AutonomousDatabaseOciAwsSecretsManagerIntegration: odbtypes.AccessEnabled,
	}
	_, err := conn.InitializeService(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionCreating, ResNameAutonomousDatabaseSecretsManagerIntegration, autonomousDatabaseSecretsManagerIntegrationID, err),
			err.Error(),
		)
		return
	}

	role, err := waitAutonomousDatabaseSecretsManagerIntegrationCreated(ctx, conn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForCreation, ResNameAutonomousDatabaseSecretsManagerIntegration, autonomousDatabaseSecretsManagerIntegrationID, err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(autonomousDatabaseSecretsManagerIntegrationID)
	flattenAutonomousDatabaseSecretsManagerIntegration(role, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseSecretsManagerIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := findAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, ResNameAutonomousDatabaseSecretsManagerIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(autonomousDatabaseSecretsManagerIntegrationID)
	flattenAutonomousDatabaseSecretsManagerIntegration(role, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseSecretsManagerIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.InitializeServiceInput{
		AutonomousDatabaseOciAwsSecretsManagerIntegration: odbtypes.AccessDisabled,
	}
	_, err := conn.InitializeService(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionDeleting, ResNameAutonomousDatabaseSecretsManagerIntegration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if err := waitAutonomousDatabaseSecretsManagerIntegrationDeleted(ctx, conn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionWaitingForDeletion, ResNameAutonomousDatabaseSecretsManagerIntegration, state.ID.String(), err),
			err.Error(),
		)
	}
}

func findAutonomousDatabaseSecretsManagerIntegration(ctx context.Context, conn *odb.Client) (*odbtypes.OciIamRole, error) {
	input := odb.GetOciOnboardingStatusInput{}
	out, err := conn.GetOciOnboardingStatus(ctx, &input)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, &retry.NotFoundError{LastError: fmt.Errorf("empty GetOciOnboardingStatus result")}
	}

	for _, role := range out.AutonomousDatabaseOciIntegrationIamRoles {
		if role.AwsIntegration == odbtypes.OciAwsIntegrationSecretsManager {
			return &role, nil
		}
	}

	return nil, &retry.NotFoundError{LastError: fmt.Errorf("%s not found", ResNameAutonomousDatabaseSecretsManagerIntegration)}
}

func statusAutonomousDatabaseSecretsManagerIntegration(conn *odb.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		role, err := findAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return role, string(role.Status), nil
	}
}

func waitAutonomousDatabaseSecretsManagerIntegrationCreated(ctx context.Context, conn *odb.Client, timeout time.Duration) (*odbtypes.OciIamRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.OciIamRoleStatusProvisioning),
		Target:  enum.Slice(odbtypes.OciIamRoleStatusAvailable),
		Refresh: statusAutonomousDatabaseSecretsManagerIntegration(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OciIamRole); ok {
		return out, err
	}

	return nil, err
}

func waitAutonomousDatabaseSecretsManagerIntegrationDeleted(ctx context.Context, conn *odb.Client, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.OciIamRoleStatusTerminating),
		Target:  []string{},
		Refresh: statusAutonomousDatabaseSecretsManagerIntegration(conn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func flattenAutonomousDatabaseSecretsManagerIntegration(role *odbtypes.OciIamRole, model *autonomousDatabaseSecretsManagerIntegrationResourceModel) {
	model.RoleARN = types.StringPointerValue(role.IamRoleArn)
	model.Status = fwtypes.StringEnumValue(role.Status)
	model.StatusReason = types.StringPointerValue(role.StatusReason)
}

type autonomousDatabaseSecretsManagerIntegrationResourceModel struct {
	framework.WithRegionModel
	ID           types.String                                  `tfsdk:"id"`
	RoleARN      types.String                                  `tfsdk:"role_arn"`
	Status       fwtypes.StringEnum[odbtypes.OciIamRoleStatus] `tfsdk:"status"`
	StatusReason types.String                                  `tfsdk:"status_reason"`
	Timeouts     timeouts.Value                                `tfsdk:"timeouts"`
}
