// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newResourceTenantDatabase(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTenantDatabase{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTenantDatabase = "Tenant Database"
)

type resourceTenantDatabase struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts

	// Use a Mutex to ensure resources are created / updated / deleted sequentially.
	// Parallel operations would fail.
	lock sync.Mutex
}

func (r *resourceTenantDatabase) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_db_tenant_database"
}

func (r *resourceTenantDatabase) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"character_set_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"db_instance_identifier": schema.StringAttribute{
				Required: true,
			},
			"dbi_resource_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deletion_protection": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"final_db_snapshot_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z]`), "must begin with alphabetic character"),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must only contain alphanumeric characters and hyphens"),
					stringvalidator.RegexMatches(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					stringvalidator.RegexMatches(regexache.MustCompile(`-$`), "cannot end in a hyphen"),
				},
			},
			"master_username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"master_user_password": schema.StringAttribute{
				Optional: true,
			},
			"nchar_character_set_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_final_snapshot": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenant_database_create_time": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_database_resource_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_db_name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceTenantDatabase) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := r.Meta().RDSConn(ctx)

	var plan resourceTenantDatabaseData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.MasterUserPassword.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("master_user_password"),
			"Required field is not set",
			"A value must be provided for this attribute when creating a Tenant Database",
		)
		return
	}

	if !plan.SkipFinalSnapshot.ValueBool() && plan.FinalDBSnapshotIdentifier.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("final_snapshot_identifier"),
			"Missing Attribute Configuration",
			fmt.Sprintf("Tenant DB \"%s\": final_snapshot_identifier is required when skip_final_snapshot is false", plan.TenantDBName.ValueString()),
		)
		return
	}

	in := &rds.CreateTenantDatabaseInput{
		DBInstanceIdentifier: aws.String(plan.DBInstanceIdentifier.ValueString()),
		MasterUsername:       aws.String(plan.MasterUsername.ValueString()),
		MasterUserPassword:   aws.String(plan.MasterUserPassword.ValueString()),
		TenantDBName:         aws.String(plan.TenantDBName.ValueString()),
		Tags:                 getTagsIn(ctx),
	}

	if !plan.CharacterSetName.IsNull() && !plan.CharacterSetName.IsUnknown() {
		in.CharacterSetName = aws.String(plan.CharacterSetName.ValueString())
	}
	if !plan.NcharCharacterSetName.IsNull() && !plan.NcharCharacterSetName.IsUnknown() {
		in.NcharCharacterSetName = aws.String(plan.NcharCharacterSetName.ValueString())
	}

	_, err := waitDBInstanceAvailableSDKv1(ctx, conn, plan.DBInstanceIdentifier.ValueString(), time.Duration(10*float64(time.Minute)))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, "DB Instance", plan.DBInstanceIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	out, err := conn.CreateTenantDatabase(in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameTenantDatabase, plan.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.TenantDatabase == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameTenantDatabase, plan.TenantDBName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.TenantDatabase.TenantDatabaseARN)
	plan.TenantDatabaseResourceId = flex.StringToFramework(ctx, out.TenantDatabase.TenantDatabaseResourceId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitTenantDatabaseCreated(ctx, conn, plan.TenantDatabaseResourceId.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, ResNameTenantDatabase, plan.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}

	plan.CharacterSetName = flex.StringToFramework(ctx, out.TenantDatabase.CharacterSetName)
	plan.DbiResourceId = flex.StringToFramework(ctx, out.TenantDatabase.DbiResourceId)
	plan.DeletionProtection = flex.BoolToFramework(ctx, out.TenantDatabase.DeletionProtection)
	plan.NcharCharacterSetName = flex.StringToFramework(ctx, out.TenantDatabase.NcharCharacterSetName)
	plan.TenantDatabaseCreateTime, _ = flex.TimeToFramework(ctx, out.TenantDatabase.TenantDatabaseCreateTime).ToStringValue(ctx)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTenantDatabase) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RDSConn(ctx)

	var state resourceTenantDatabaseData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	db, err := findTenantDatabaseByName(ctx, conn, state.DBInstanceIdentifier.ValueString(), state.TenantDBName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionSetting, ResNameTenantDatabase, state.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, db.TenantDatabaseARN)
	state.CharacterSetName = flex.StringToFramework(ctx, db.CharacterSetName)
	state.DBInstanceIdentifier = flex.StringToFramework(ctx, db.DBInstanceIdentifier)
	state.DbiResourceId = flex.StringToFramework(ctx, db.DbiResourceId)
	state.DeletionProtection = flex.BoolToFramework(ctx, db.DeletionProtection)
	state.MasterUsername = flex.StringToFramework(ctx, db.MasterUsername)
	state.NcharCharacterSetName = flex.StringToFramework(ctx, db.NcharCharacterSetName)
	state.TenantDatabaseCreateTime, _ = flex.TimeToFramework(ctx, db.TenantDatabaseCreateTime).ToStringValue(ctx)
	state.TenantDatabaseResourceId = flex.StringToFramework(ctx, db.TenantDatabaseResourceId)
	state.TenantDBName = flex.StringToFramework(ctx, db.TenantDBName)

	setTagsOut(ctx, db.TagList)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTenantDatabase) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := r.Meta().RDSConn(ctx)

	var plan, state resourceTenantDatabaseData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TenantDBName.Equal(state.TenantDBName) ||
		!plan.MasterUserPassword.Equal(state.MasterUserPassword) {

		in := &rds.ModifyTenantDatabaseInput{
			DBInstanceIdentifier: aws.String(plan.DBInstanceIdentifier.ValueString()),
			TenantDBName:         aws.String(state.TenantDBName.ValueString()),
		}

		if !plan.TenantDBName.Equal(state.TenantDBName) {
			in.NewTenantDBName = aws.String(plan.TenantDBName.ValueString())
		}

		if !plan.MasterUserPassword.IsNull() {
			in.MasterUserPassword = aws.String(plan.MasterUserPassword.ValueString())
		}

		out, err := conn.ModifyTenantDatabaseWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RDS, create.ErrActionUpdating, ResNameTenantDatabase, plan.TenantDBName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.TenantDatabase == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.RDS, create.ErrActionUpdating, ResNameTenantDatabase, plan.TenantDBName.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.TenantDatabase.TenantDatabaseARN)
		plan.TenantDBName = flex.StringToFramework(ctx, out.TenantDatabase.TenantDBName)
	}

	_, err := waitDBInstanceAvailableSDKv1(ctx, conn, plan.DBInstanceIdentifier.ValueString(), time.Duration(10*float64(time.Minute)))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, "DB Instance", plan.DBInstanceIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err = waitTenantDatabaseUpdated(ctx, conn, plan.TenantDatabaseResourceId.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForUpdate, ResNameTenantDatabase, plan.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTenantDatabase) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.lock.Lock()
	defer r.lock.Unlock()

	conn := r.Meta().RDSConn(ctx)

	var state resourceTenantDatabaseData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rds.DeleteTenantDatabaseInput{
		DBInstanceIdentifier:      aws.String(state.DBInstanceIdentifier.ValueString()),
		TenantDBName:              aws.String(state.TenantDBName.ValueString()),
		FinalDBSnapshotIdentifier: aws.String(state.FinalDBSnapshotIdentifier.ValueString()),
		SkipFinalSnapshot:         aws.Bool(state.SkipFinalSnapshot.ValueBool()),
	}

	_, err := waitDBInstanceAvailableSDKv1(ctx, conn, state.DBInstanceIdentifier.ValueString(), time.Duration(10*float64(time.Minute)))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, "DB Instance", state.DBInstanceIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteTenantDatabaseWithContext(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionDeleting, ResNameTenantDatabase, state.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTenantDatabaseDeleted(ctx, conn, state.TenantDatabaseResourceId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForDeletion, ResNameTenantDatabase, state.TenantDBName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTenantDatabase) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceTenantDatabase) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: db_instance_identifier/tenant_db_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("db_instance_identifier"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_db_name"), idParts[1])...)
}

const (
	TenantDBStatusAvailable        = "available"
	TenantDBStatusRenaming         = "renaming"
	TenantDBStatusCredentialsReset = "resetting-master-credentials"
	TenantDBStatusDeleting         = "deleting"
)

func waitTenantDatabaseCreated(ctx context.Context, conn *rds.RDS, tenantDBResourceId string, timeout time.Duration) (*rds.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{TenantDBStatusAvailable},
		Refresh:                   statusTenantDatabase(ctx, conn, tenantDBResourceId),
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.TenantDatabase); ok {
		return out, err
	}

	return nil, err
}

func waitTenantDatabaseUpdated(ctx context.Context, conn *rds.RDS, tenantDBResourceId string, timeout time.Duration) (*rds.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{TenantDBStatusRenaming, TenantDBStatusCredentialsReset},
		Target:                    []string{TenantDBStatusAvailable},
		Refresh:                   statusTenantDatabase(ctx, conn, tenantDBResourceId),
		PollInterval:              10 * time.Second,
		Delay:                     1 * time.Minute,
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.TenantDatabase); ok {
		return out, err
	}

	return nil, err
}

func waitTenantDatabaseDeleted(ctx context.Context, conn *rds.RDS, tenantDBResourceId string, timeout time.Duration) (*rds.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TenantDBStatusDeleting},
		Target:  []string{},
		Refresh: statusTenantDatabase(ctx, conn, tenantDBResourceId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*rds.TenantDatabase); ok {
		return out, err
	}

	return nil, err
}

func statusTenantDatabase(ctx context.Context, conn *rds.RDS, tenantDBResourceId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findTenantDatabaseById(ctx, conn, tenantDBResourceId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.Status), nil
	}
}

func findTenantDatabaseById(ctx context.Context, conn *rds.RDS, tenantDBResourceId string) (*rds.TenantDatabase, error) {
	input := &rds.DescribeTenantDatabasesInput{
		Filters: []*rds.Filter{
			{
				Name:   aws.String("tenant-database-resource-id"),
				Values: []*string{&tenantDBResourceId},
			},
		},
	}

	output, err := findTenantDatabases(ctx, conn, input, tfslices.PredicateTrue[*rds.TenantDatabase]())
	if err != nil {
		return nil, err
	}

	db, err := tfresource.AssertSinglePtrResult(output)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func findTenantDatabaseByName(ctx context.Context, conn *rds.RDS, dBInstanceIdentifier string, tenantDBName string) (*rds.TenantDatabase, error) {
	input := &rds.DescribeTenantDatabasesInput{
		DBInstanceIdentifier: aws.String(dBInstanceIdentifier),
		TenantDBName:         aws.String(tenantDBName),
	}

	output, err := findTenantDatabases(ctx, conn, input, tfslices.PredicateTrue[*rds.TenantDatabase]())
	if err != nil {
		return nil, err
	}

	db, err := tfresource.AssertSinglePtrResult(output)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func findTenantDatabases(ctx context.Context, conn *rds.RDS, input *rds.DescribeTenantDatabasesInput, filter tfslices.Predicate[*rds.TenantDatabase]) ([]*rds.TenantDatabase, error) {
	var output []*rds.TenantDatabase

	err := conn.DescribeTenantDatabasesPagesWithContext(ctx, input, func(page *rds.DescribeTenantDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TenantDatabases {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type resourceTenantDatabaseData struct {
	ARN                       types.String   `tfsdk:"arn"`
	CharacterSetName          types.String   `tfsdk:"character_set_name"`
	DBInstanceIdentifier      types.String   `tfsdk:"db_instance_identifier"`
	DbiResourceId             types.String   `tfsdk:"dbi_resource_id"`
	DeletionProtection        types.Bool     `tfsdk:"deletion_protection"`
	FinalDBSnapshotIdentifier types.String   `tfsdk:"final_db_snapshot_identifier"`
	MasterUsername            types.String   `tfsdk:"master_username"`
	MasterUserPassword        types.String   `tfsdk:"master_user_password"`
	NcharCharacterSetName     types.String   `tfsdk:"nchar_character_set_name"`
	SkipFinalSnapshot         types.Bool     `tfsdk:"skip_final_snapshot"`
	Tags                      types.Map      `tfsdk:"tags"`
	TagsAll                   types.Map      `tfsdk:"tags_all"`
	TenantDatabaseCreateTime  types.String   `tfsdk:"tenant_database_create_time"`
	TenantDatabaseResourceId  types.String   `tfsdk:"tenant_database_resource_id"`
	TenantDBName              types.String   `tfsdk:"tenant_db_name"`
	Timeouts                  timeouts.Value `tfsdk:"timeouts"`
}
