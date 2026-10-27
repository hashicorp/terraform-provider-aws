// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_tenant_database", name="Tenant Database")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("tenant_database_resource_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/rds/types;types.TenantDatabase")
// @Testing(importIgnore="master_user_password_wo_version;skip_final_snapshot")
// @Testing(importStateIdAttribute="tenant_database_resource_id")
// Each test requires an RDS for Oracle CDB instance, so generated tags tests are too slow and costly.
// @Testing(tagsTest=false)
func newTenantDatabaseResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tenantDatabaseResource{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

const (
	tenantDatabaseStatusAvailable                  = "available"
	tenantDatabaseStatusCreating                   = "creating"
	tenantDatabaseStatusDeleting                   = "deleting"
	tenantDatabaseStatusModifying                  = "modifying"
	tenantDatabaseStatusRenaming                   = "renaming"
	tenantDatabaseStatusResettingMasterCredentials = "resetting-master-credentials"
	// Not an RDS status. Reported while the tenant database is available but has pending modifications.
	tenantDatabaseStatusPendingModifications = "tf-pending-modifications"
)

type tenantDatabaseResource struct {
	framework.ResourceWithModel[tenantDatabaseResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *tenantDatabaseResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrFinalSnapshotIdentifier: schema.StringAttribute{
				Optional: true,
			},
			"manage_master_user_password": schema.BoolAttribute{
				Optional: true,
			},
			"master_user_password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("master_user_password_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("master_user_password_wo")),
				},
			},
			"master_user_password_wo": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				WriteOnly: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("master_user_password_wo_version")),
				},
			},
			"master_user_password_wo_version": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("master_user_password_wo")),
				},
			},
			"master_user_secret": framework.ResourceComputedListOfObjectsAttribute[masterUserSecretModel](ctx),
			"master_user_secret_kms_key_id": schema.StringAttribute{
				Optional: true,
			},
			"master_username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *tenantDatabaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan, config tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.TenantDBName)
	var input rds.CreateTenantDatabaseInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Prefer write-only value. It's only in Config, not Plan.
	if !config.MasterUserPasswordWO.IsNull() {
		input.MasterUserPassword = fwflex.StringFromFramework(ctx, config.MasterUserPasswordWO)
	}

	input.Tags = getTagsIn(ctx)

	timeout := r.CreateTimeout(ctx, plan.Timeouts)
	// Only one tenant database operation can run on a DB instance at a time.
	out, err := tfresource.RetryWhenIsA[*rds.CreateTenantDatabaseOutput, *awstypes.InvalidDBInstanceStateFault](ctx, timeout, func(ctx context.Context) (*rds.CreateTenantDatabaseOutput, error) {
		return conn.CreateTenantDatabase(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
	if out == nil || out.TenantDatabase == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, name)
		return
	}

	id := aws.ToString(out.TenantDatabase.TenantDatabaseResourceId)
	// If a wait fails, save enough state to taint the tenant database so it's destroyed rather than orphaned.
	taint := func() {
		for name, v := range map[string]attr.Value{
			names.AttrRegion:                  plan.Region,
			"db_instance_identifier":          plan.DBInstanceIdentifier,
			names.AttrFinalSnapshotIdentifier: plan.FinalSnapshotIdentifier,
			"skip_final_snapshot":             plan.SkipFinalSnapshot,
			"tenant_database_resource_id":     types.StringValue(id),
			"tenant_db_name":                  plan.TenantDBName,
		} {
			smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(name), v))
		}
	}

	tenantDatabase, err := waitTenantDatabaseAvailable(ctx, conn, id, timeout)
	if err != nil {
		taint()
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if err := waitTenantDatabaseInstanceAvailable(ctx, conn, tenantDatabase, timeout); err != nil {
		taint()
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, tenantDatabase, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *tenantDatabaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TenantDatabaseResourceID)
	out, err := findTenantDatabaseByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.TagList)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tenantDatabaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state, config tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.TenantDatabaseResourceID)
	input := rds.ModifyTenantDatabaseInput{
		DBInstanceIdentifier: fwflex.StringFromFramework(ctx, state.DBInstanceIdentifier),
		TenantDBName:         fwflex.StringFromFramework(ctx, state.TenantDBName),
	}
	var modify bool

	if !plan.TenantDBName.Equal(state.TenantDBName) {
		input.NewTenantDBName = fwflex.StringFromFramework(ctx, plan.TenantDBName)
		modify = true
	}

	if !plan.MasterUserPassword.Equal(state.MasterUserPassword) && !plan.MasterUserPassword.IsNull() {
		input.MasterUserPassword = fwflex.StringFromFramework(ctx, plan.MasterUserPassword)
		modify = true
	}

	if !plan.MasterUserPasswordWOVersion.Equal(state.MasterUserPasswordWOVersion) && !config.MasterUserPasswordWO.IsNull() {
		input.MasterUserPassword = fwflex.StringFromFramework(ctx, config.MasterUserPasswordWO)
		modify = true
	}

	if !plan.ManageMasterUserPassword.Equal(state.ManageMasterUserPassword) {
		// Null means not managed, so send false rather than omitting the value.
		input.ManageMasterUserPassword = aws.Bool(fwflex.BoolValueFromFramework(ctx, plan.ManageMasterUserPassword))
		// The KMS key can only be set when turning on management (see ModifyPlan).
		if plan.ManageMasterUserPassword.ValueBool() {
			input.MasterUserSecretKmsKeyId = fwflex.StringFromFramework(ctx, plan.MasterUserSecretKMSKeyID)
		}
		modify = true
	}

	if modify {
		timeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err := tfresource.RetryWhenIsA[*rds.ModifyTenantDatabaseOutput, *awstypes.InvalidDBInstanceStateFault](ctx, timeout, func(ctx context.Context) (*rds.ModifyTenantDatabaseOutput, error) {
			return conn.ModifyTenantDatabase(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		tenantDatabase, err := waitTenantDatabaseAvailable(ctx, conn, id, timeout)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		if err := waitTenantDatabaseInstanceAvailable(ctx, conn, tenantDatabase, timeout); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, tenantDatabase, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	} else {
		// No API change, so computed values are unchanged.
		plan.MasterUserSecret = state.MasterUserSecret
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *tenantDatabaseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RDSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.TenantDatabaseResourceID)
	// Look up the current names; the DB instance may have been renamed.
	tenantDatabase, err := findTenantDatabaseByID(ctx, conn, id)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	input := rds.DeleteTenantDatabaseInput{
		DBInstanceIdentifier: tenantDatabase.DBInstanceIdentifier,
		SkipFinalSnapshot:    fwflex.BoolFromFramework(ctx, data.SkipFinalSnapshot),
		TenantDBName:         tenantDatabase.TenantDBName,
	}
	if !data.SkipFinalSnapshot.ValueBool() {
		input.FinalDBSnapshotIdentifier = fwflex.StringFromFramework(ctx, data.FinalSnapshotIdentifier)
	}

	timeout := r.DeleteTimeout(ctx, data.Timeouts)
	_, err = tfresource.RetryWhenIsA[*rds.DeleteTenantDatabaseOutput, *awstypes.InvalidDBInstanceStateFault](ctx, timeout, func(ctx context.Context) (*rds.DeleteTenantDatabaseOutput, error) {
		return conn.DeleteTenantDatabase(ctx, &input)
	})
	if errs.IsA[*awstypes.DBInstanceNotFoundFault](err) || errs.IsA[*awstypes.TenantDatabaseNotFoundFault](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitTenantDatabaseDeleted(ctx, conn, id, timeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if err := waitTenantDatabaseInstanceAvailable(ctx, conn, tenantDatabase, timeout); err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *tenantDatabaseResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() || request.Plan.Raw.IsNull() {
		return
	}

	var plan, state tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	// RDS can't change the KMS key of an existing managed secret.
	// An unknown key is checked when Terraform plans again during apply with the resolved value.
	if state.ManageMasterUserPassword.ValueBool() && plan.ManageMasterUserPassword.ValueBool() && !plan.MasterUserSecretKMSKeyID.IsUnknown() && !plan.MasterUserSecretKMSKeyID.Equal(state.MasterUserSecretKMSKeyID) {
		response.Diagnostics.AddAttributeError(
			path.Root("master_user_secret_kms_key_id"),
			"Unsupported change",
			"master_user_secret_kms_key_id can't be changed while manage_master_user_password is true. "+
				"Turn off manage_master_user_password, apply, then turn it on with the new key.",
		)
	}
}

func (r *tenantDatabaseResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data tenantDatabaseResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Unlike ConflictsWith, allow a password when management is explicitly false; turning management off requires one.
	if data.ManageMasterUserPassword.ValueBool() {
		for name, v := range map[string]types.String{
			"master_user_password":    data.MasterUserPassword,
			"master_user_password_wo": data.MasterUserPasswordWO,
		} {
			if !v.IsNull() {
				response.Diagnostics.AddAttributeError(
					path.Root(name),
					"Invalid Attribute Combination",
					fmt.Sprintf(`"%s" can't be set when "manage_master_user_password" is true`, name),
				)
			}
		}
	}

	if !data.SkipFinalSnapshot.IsUnknown() && !data.FinalSnapshotIdentifier.IsUnknown() && !data.SkipFinalSnapshot.ValueBool() && data.FinalSnapshotIdentifier.IsNull() {
		response.Diagnostics.AddAttributeError(
			path.Root(names.AttrFinalSnapshotIdentifier),
			"Missing Attribute Configuration",
			`"final_snapshot_identifier" is required unless "skip_final_snapshot" is true`,
		)
	}
}

func findTenantDatabaseByID(ctx context.Context, conn *rds.Client, id string) (*awstypes.TenantDatabase, error) {
	input := rds.DescribeTenantDatabasesInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("tenant-database-resource-id"),
				Values: []string{id},
			},
		},
	}

	return findTenantDatabase(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.TenantDatabase]())
}

func findTenantDatabase(ctx context.Context, conn *rds.Client, input *rds.DescribeTenantDatabasesInput, filter tfslices.Predicate[*awstypes.TenantDatabase]) (*awstypes.TenantDatabase, error) {
	output, err := findTenantDatabases(ctx, conn, input, filter)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTenantDatabases(ctx context.Context, conn *rds.Client, input *rds.DescribeTenantDatabasesInput, filter tfslices.Predicate[*awstypes.TenantDatabase]) ([]awstypes.TenantDatabase, error) {
	var output []awstypes.TenantDatabase

	pages := rds.NewDescribeTenantDatabasesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.DBInstanceNotFoundFault](err) || errs.IsA[*awstypes.TenantDatabaseNotFoundFault](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.TenantDatabases {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusTenantDatabase(conn *rds.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTenantDatabaseByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		status := aws.ToString(output.Status)
		if v := output.PendingModifiedValues; status == tenantDatabaseStatusAvailable && v != nil && (v.MasterUserPassword != nil || v.TenantDBName != nil) {
			status = tenantDatabaseStatusPendingModifications
		}

		return output, status, nil
	}
}

func waitTenantDatabaseAvailable(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			tenantDatabaseStatusCreating,
			tenantDatabaseStatusModifying,
			tenantDatabaseStatusPendingModifications,
			tenantDatabaseStatusRenaming,
			tenantDatabaseStatusResettingMasterCredentials,
		},
		Target:                    []string{tenantDatabaseStatusAvailable},
		Refresh:                   statusTenantDatabase(conn, id),
		Timeout:                   timeout,
		Delay:                     30 * time.Second,
		PollInterval:              10 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.TenantDatabase); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTenantDatabaseDeleted(ctx context.Context, conn *rds.Client, id string, timeout time.Duration) (*awstypes.TenantDatabase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			tenantDatabaseStatusAvailable,
			tenantDatabaseStatusDeleting,
			tenantDatabaseStatusModifying,
		},
		Target:       []string{},
		Refresh:      statusTenantDatabase(conn, id),
		Timeout:      timeout,
		Delay:        30 * time.Second,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.TenantDatabase); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// Tenant database operations also modify the DB instance; wait so that subsequent operations on it succeed.
// During these operations the DB instance reports statuses such as "creating-tenant-database" that
// waitDBInstanceAvailable doesn't expect, so only known terminal statuses are treated as errors here.
func waitTenantDatabaseInstanceAvailable(ctx context.Context, conn *rds.Client, tenantDatabase *awstypes.TenantDatabase, timeout time.Duration) error {
	id := aws.ToString(tenantDatabase.DbiResourceId)

	return tfresource.WaitUntil(ctx, timeout, func(ctx context.Context) (bool, error) {
		output, err := findDBInstanceByID(ctx, conn, id)
		if err != nil {
			return false, err
		}

		switch status := aws.ToString(output.DBInstanceStatus); status {
		case instanceStatusAvailable, instanceStatusStorageOptimization:
			return true, nil
		case instanceStatusFailed, instanceStatusIncompatibleOptionGroup, instanceStatusIncompatibleParameters, instanceStatusStopped:
			return false, fmt.Errorf("unexpected RDS DB Instance (%s) status: %s", id, status)
		default:
			return false, nil
		}
	}, tfresource.WaitOpts{
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
		PollInterval:              10 * time.Second,
	})
}

type tenantDatabaseResourceModel struct {
	framework.WithRegionModel
	CharacterSetName            types.String                                           `tfsdk:"character_set_name"`
	DBInstanceIdentifier        types.String                                           `tfsdk:"db_instance_identifier"`
	FinalSnapshotIdentifier     types.String                                           `tfsdk:"final_snapshot_identifier" autoflex:"-"`
	ManageMasterUserPassword    types.Bool                                             `tfsdk:"manage_master_user_password"`
	MasterUserPassword          types.String                                           `tfsdk:"master_user_password"`
	MasterUserPasswordWO        types.String                                           `tfsdk:"master_user_password_wo" autoflex:"-"`
	MasterUserPasswordWOVersion types.Int64                                            `tfsdk:"master_user_password_wo_version" autoflex:"-"`
	MasterUserSecret            fwtypes.ListNestedObjectValueOf[masterUserSecretModel] `tfsdk:"master_user_secret"`
	MasterUserSecretKMSKeyID    types.String                                           `tfsdk:"master_user_secret_kms_key_id"`
	MasterUsername              types.String                                           `tfsdk:"master_username"`
	NcharCharacterSetName       types.String                                           `tfsdk:"nchar_character_set_name"`
	SkipFinalSnapshot           types.Bool                                             `tfsdk:"skip_final_snapshot" autoflex:"-"`
	Tags                        tftags.Map                                             `tfsdk:"tags"`
	TagsAll                     tftags.Map                                             `tfsdk:"tags_all"`
	TenantDBName                types.String                                           `tfsdk:"tenant_db_name"`
	TenantDatabaseARN           types.String                                           `tfsdk:"arn"`
	TenantDatabaseResourceID    types.String                                           `tfsdk:"tenant_database_resource_id"`
	Timeouts                    timeouts.Value                                         `tfsdk:"timeouts"`
}

type masterUserSecretModel struct {
	KMSKeyID     types.String `tfsdk:"kms_key_id"`
	SecretARN    types.String `tfsdk:"secret_arn"`
	SecretStatus types.String `tfsdk:"secret_status"`
}
