// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_odb_autonomous_database_secrets_manager_integration", name="Autonomous Database Secrets Manager Integration")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(serialize=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/odb/types;odbtypes;odbtypes.OciIamRole")
// @Testing(preCheckWithRegion="testAccAutonomousDatabaseSecretsManagerIntegrationPreCheckRegion")
func newResourceAutonomousDatabaseSecretsManagerIntegration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAutonomousDatabaseSecretsManagerIntegration{}
	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const ResNameAutonomousDatabaseSecretsManagerIntegration = "Autonomous Database Secrets Manager Integration"

type resourceAutonomousDatabaseSecretsManagerIntegration struct {
	framework.ResourceWithModel[autonomousDatabaseSecretsManagerIntegrationResourceModel]
	framework.WithImportByIdentity
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
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.InitializeServiceInput{
		AutonomousDatabaseOciAwsSecretsManagerIntegration: odbtypes.AccessEnabled,
	}
	region := r.Meta().Region(ctx)
	tflog.Debug(ctx, "Enabling ODB Autonomous Database Secrets Manager integration", map[string]any{names.AttrRegion: region})
	_, err := conn.InitializeService(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, region)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), region))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrRegion), region))
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := waitAutonomousDatabaseSecretsManagerIntegrationCreated(ctx, conn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, region)
		return
	}

	plan.ID = types.StringValue(region)
	flattenAutonomousDatabaseSecretsManagerIntegration(role, &plan)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseSecretsManagerIntegrationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := findAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
	if retry.NotFound(err) {
		tflog.Debug(ctx, "ODB Autonomous Database Secrets Manager integration no longer exists; removing from state")
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	// Read treats an integration being disabled outside Terraform as drift. The
	// waiter still needs the raw role to wait for termination to finish.
	switch role.Status {
	case odbtypes.OciIamRoleStatusTerminating:
		tflog.Debug(ctx, "ODB Autonomous Database Secrets Manager integration is terminating; removing from state", map[string]any{names.AttrStatus: role.Status})
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(smarterr.Errorf("integration is terminating (%s)", role.Status)))
		resp.State.RemoveResource(ctx)
		return
	case odbtypes.OciIamRoleStatusTerminateFailed:
		tflog.Warn(ctx, "ODB Autonomous Database Secrets Manager integration termination failed; retaining state", map[string]any{names.AttrStatus: role.Status})
		reason := aws.ToString(role.StatusReason)
		if reason == "" {
			reason = "no status reason was returned"
		}
		smerr.AddError(ctx, &resp.Diagnostics, smarterr.Errorf("integration termination failed (%s): %s; resolve the service role failure before retrying", role.Status, reason), smerr.ID, state.ID.ValueString())
		return
	}

	state.ID = types.StringValue(r.Meta().Region(ctx))
	flattenAutonomousDatabaseSecretsManagerIntegration(role, &state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceAutonomousDatabaseSecretsManagerIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ODBClient(ctx)

	var state autonomousDatabaseSecretsManagerIntegrationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := odb.InitializeServiceInput{
		AutonomousDatabaseOciAwsSecretsManagerIntegration: odbtypes.AccessDisabled,
	}
	tflog.Debug(ctx, "Disabling ODB Autonomous Database Secrets Manager integration", map[string]any{names.AttrRegion: r.Meta().Region(ctx)})
	_, err := conn.InitializeService(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	if err := waitAutonomousDatabaseSecretsManagerIntegrationDeleted(ctx, conn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

func findAutonomousDatabaseSecretsManagerIntegration(ctx context.Context, conn *odb.Client) (*odbtypes.OciIamRole, error) {
	input := odb.GetOciOnboardingStatusInput{}
	out, err := conn.GetOciOnboardingStatus(ctx, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: fmt.Errorf("empty GetOciOnboardingStatus result")})
	}

	for _, role := range out.AutonomousDatabaseOciIntegrationIamRoles {
		if role.AwsIntegration == odbtypes.OciAwsIntegrationSecretsManager {
			return &role, nil
		}
	}

	return nil, smarterr.NewError(&retry.NotFoundError{LastError: fmt.Errorf("%s not found", ResNameAutonomousDatabaseSecretsManagerIntegration)})
}

func statusAutonomousDatabaseSecretsManagerIntegration(conn *odb.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		role, err := findAutonomousDatabaseSecretsManagerIntegration(ctx, conn)
		if retry.NotFound(err) {
			tflog.Trace(ctx, "ODB Autonomous Database Secrets Manager integration is not yet visible or has been removed")
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		tflog.Trace(ctx, "Read ODB Autonomous Database Secrets Manager integration lifecycle state", map[string]any{names.AttrStatus: role.Status})
		return role, string(role.Status), nil
	}
}

func waitAutonomousDatabaseSecretsManagerIntegrationCreated(ctx context.Context, conn *odb.Client, timeout time.Duration) (*odbtypes.OciIamRole, error) {
	tflog.Debug(ctx, "Waiting for ODB Autonomous Database Secrets Manager integration to become available")
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.OciIamRoleStatusProvisioning),
		Target:  enum.Slice(odbtypes.OciIamRoleStatusAvailable),
		Refresh: statusAutonomousDatabaseSecretsManagerIntegration(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OciIamRole); ok {
		return out, autonomousDatabaseSecretsManagerIntegrationWaitError(ctx, out, err)
	}

	return nil, smarterr.NewError(err)
}

func waitAutonomousDatabaseSecretsManagerIntegrationDeleted(ctx context.Context, conn *odb.Client, timeout time.Duration) error {
	tflog.Debug(ctx, "Waiting for ODB Autonomous Database Secrets Manager integration to be removed")
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(odbtypes.OciIamRoleStatusAvailable, odbtypes.OciIamRoleStatusProvisioning, odbtypes.OciIamRoleStatusTerminating),
		Target:  []string{},
		Refresh: statusAutonomousDatabaseSecretsManagerIntegration(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*odbtypes.OciIamRole); ok {
		return autonomousDatabaseSecretsManagerIntegrationWaitError(ctx, out, err)
	}
	return smarterr.NewError(err)
}

func autonomousDatabaseSecretsManagerIntegrationWaitError(ctx context.Context, role *odbtypes.OciIamRole, err error) error {
	if err == nil {
		return nil
	}
	if role.StatusReason != nil {
		retry.SetLastError(err, errors.New(aws.ToString(role.StatusReason)))
	}
	switch role.Status {
	case odbtypes.OciIamRoleStatusProvisionFailed:
		tflog.Warn(ctx, "ODB Autonomous Database Secrets Manager integration provisioning failed", map[string]any{names.AttrStatus: role.Status})
		err = fmt.Errorf("integration provisioning failed; resolve the service role failure before retrying: %w", err)
	case odbtypes.OciIamRoleStatusTerminateFailed:
		tflog.Warn(ctx, "ODB Autonomous Database Secrets Manager integration termination failed", map[string]any{names.AttrStatus: role.Status})
		err = fmt.Errorf("integration termination failed; resolve the service role failure before retrying: %w", err)
	}
	return smarterr.NewError(err)
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
