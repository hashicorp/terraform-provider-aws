// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_timestreaminfluxdb_db_backup", name="DB Backup")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb;timestreaminfluxdb.GetDbBackupOutput")
// @Testing(hasNoPreExistingResource=true)
func newDBBackupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &dbBackupResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type dbBackupResource struct {
	framework.ResourceWithModel[dbBackupResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *dbBackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAllocatedStorage: schema.Int32Attribute{
				Computed:    true,
				Description: `The allocated storage of the resource at the time of backup, in GiB.`,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: `The time when the backup was created.`,
			},
			"db_instance_type": schema.StringAttribute{
				Computed:    true,
				Description: `The DB instance type of the resource at the time of backup.`,
			},
			"db_parameter_group_id": schema.StringAttribute{
				Computed:    true,
				Description: `The identifier of the DB parameter group associated with the backup.`,
			},
			"db_resource_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `The id of the DB instance or DB cluster to back up.`,
			},
			"db_storage_type": schema.StringAttribute{
				Computed:    true,
				Description: `The storage type of the resource at the time of backup.`,
			},
			"deployment_type": schema.StringAttribute{
				Computed:    true,
				Description: `The deployment type of the resource that the backup was created from.`,
			},
			"engine_type": schema.StringAttribute{
				Computed:    true,
				Description: `The engine type of the resource that the backup was created from.`,
			},
			"expires_after": schema.StringAttribute{
				Computed:    true,
				Description: `The date after which the backup will be automatically deleted.`,
			},
			"failover_mode": schema.StringAttribute{
				Computed:    true,
				Description: `The failover mode of the resource at the time of backup.`,
			},
			names.AttrID: framework.IDAttribute(),
			"influx_auth_parameters_secret_arn": schema.StringAttribute{
				Computed:    true,
				Description: `The ARN of the Secrets Manager secret containing the InfluxDB auth parameters.`,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Computed:    true,
				Description: `The AWS KMS key ARN used for encryption of the resource at the time of backup.`,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: `The name of the backup. Must be unique within the account and Region.`,
			},
			"network_type": schema.StringAttribute{
				Computed:    true,
				Description: `The network type of the resource at the time of backup.`,
			},
			names.AttrPort: schema.Int32Attribute{
				Computed:    true,
				Description: `The port number of the resource at the time of backup.`,
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Computed:    true,
				Description: `Indicates whether the resource was publicly accessible at the time of backup.`,
			},
			"retention_days": schema.Int32Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int32{
					int32validator.Between(1, 3650),
				},
				Description: `The number of days to retain the backup. Valid values are 1 to 3650.`,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: `The current status of the backup.`,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Computed:    true,
				Description: `The type of backup.`,
			},
			names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Computed:    true,
				Description: `The VPC security group IDs associated with the resource at the time of backup.`,
			},
			"vpc_subnet_ids": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Computed:    true,
				Description: `The VPC subnet IDs associated with the resource at the time of backup.`,
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

func (r *dbBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dbBackupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var input timestreaminfluxdb.CreateDbBackupInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDbBackup(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.Name))
		return
	}

	backupID := aws.ToString(out.Id)
	plan.ID = fwflex.StringValueToFramework(ctx, backupID)

	output, err := waitDBBackupCreated(ctx, conn, backupID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), backupID)) // Set 'id' so as to taint the resource.
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, backupID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *dbBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dbBackupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	backupID := fwflex.StringValueFromFramework(ctx, state.ID)
	output, err := findDBBackupByID(ctx, conn, backupID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, backupID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *dbBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dbBackupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	backupID := fwflex.StringValueFromFramework(ctx, state.ID)
	input := timestreaminfluxdb.DeleteDbBackupInput{
		Identifier: aws.String(backupID),
	}

	_, err := conn.DeleteDbBackup(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, backupID)
		return
	}

	if _, err := waitDBBackupDeleted(ctx, conn, backupID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, backupID)
		return
	}
}

func waitDBBackupCreated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbBackupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DbBackupStatusInProgress),
		Target:                    enum.Slice(awstypes.DbBackupStatusCompleted),
		Refresh:                   statusDBBackup(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbBackupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBBackupDeleted(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbBackupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DbBackupStatusDeleting),
		Target:  []string{},
		Refresh: statusDBBackup(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbBackupOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDBBackup(conn *timestreaminfluxdb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findDBBackupByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findDBBackupByID(ctx context.Context, conn *timestreaminfluxdb.Client, id string) (*timestreaminfluxdb.GetDbBackupOutput, error) {
	in := timestreaminfluxdb.GetDbBackupInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetDbBackup(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type dbBackupResourceModel struct {
	framework.WithRegionModel
	AllocatedStorage              types.Int32         `tfsdk:"allocated_storage"`
	ARN                           types.String        `tfsdk:"arn"`
	CreatedAt                     timetypes.RFC3339   `tfsdk:"created_at"`
	DBInstanceType                types.String        `tfsdk:"db_instance_type"`
	DBParameterGroupID            types.String        `tfsdk:"db_parameter_group_id"`
	DBResourceID                  types.String        `tfsdk:"db_resource_id"`
	DBStorageType                 types.String        `tfsdk:"db_storage_type"`
	DeploymentType                types.String        `tfsdk:"deployment_type"`
	EngineType                    types.String        `tfsdk:"engine_type"`
	ExpiresAfter                  types.String        `tfsdk:"expires_after"`
	FailoverMode                  types.String        `tfsdk:"failover_mode"`
	ID                            types.String        `tfsdk:"id"`
	InfluxAuthParametersSecretARN types.String        `tfsdk:"influx_auth_parameters_secret_arn"`
	KMSKeyID                      types.String        `tfsdk:"kms_key_id"`
	Name                          types.String        `tfsdk:"name"`
	NetworkType                   types.String        `tfsdk:"network_type"`
	Port                          types.Int32         `tfsdk:"port"`
	PubliclyAccessible            types.Bool          `tfsdk:"publicly_accessible"`
	RetentionDays                 types.Int32         `tfsdk:"retention_days"`
	Status                        types.String        `tfsdk:"status"`
	Tags                          tftags.Map          `tfsdk:"tags"`
	TagsAll                       tftags.Map          `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value      `tfsdk:"timeouts"`
	Type                          types.String        `tfsdk:"type"`
	VPCSecurityGroupIDs           fwtypes.SetOfString `tfsdk:"vpc_security_group_ids"`
	VPCSubnetIDs                  fwtypes.SetOfString `tfsdk:"vpc_subnet_ids"`
}
