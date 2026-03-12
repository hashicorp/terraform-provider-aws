// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package timestreaminfluxdb

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_timestreaminfluxdb_db_cluster", name="DB Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb;timestreaminfluxdb.GetDbClusterOutput")
// @Testing(importIgnore="bucket;username;organization;password")
func newDBClusterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &dbClusterResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDBCluster = "DB Cluster"
)

type dbClusterResource struct {
	framework.ResourceWithModel[dbClusterResourceModel]
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *dbClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAllocatedStorage: schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(20, 16384),
				},
				Description: `The amount of storage to allocate for your DB storage type in GiB (gibibytes).
					This field is forbidden for InfluxDB V3 clusters (when using an InfluxDB V3 db parameter group).`,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrBucket: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[^_][^\"]*$"),
						"",
					),
				},
				Description: `Name of the initial InfluxDB bucket. All InfluxDB data is stored in a bucket.
					A bucket combines the concept of a database and a retention period (the duration of time
					that each data point persists). A bucket belongs to an organization. Along with organization,
					username, and password, this argument will be stored in the secret referred to by the
					influx_auth_parameters_secret_arn attribute. This field is forbidden for InfluxDB V3 clusters
					(when using an InfluxDB V3 db parameter group).`,
			},
			"db_instance_type": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.DbInstanceType](),
				Required:    true,
				Description: `The Timestream for InfluxDB DB instance type to run InfluxDB on.`,
			},
			"db_parameter_group_identifier": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						dbClusterDBParameterGroupIdentifierReplaceIf, "Replace db_parameter_group_identifier diff", "Replace db_parameter_group_identifier diff",
					),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-Z0-9]+$"),
						"",
					),
				},
				Description: `The ID of the DB parameter group to assign to your DB cluster. 
					DB parameter groups specify how the database is configured. For example, DB parameter groups 
					can specify the limit for query concurrency.`,
			},
			"db_storage_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DbStorageType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The Timestream for InfluxDB DB storage type to read and write InfluxDB data. 
					You can choose between 3 different types of provisioned Influx IOPS included storage according 
					to your workloads requirements: Influx IO Included 3000 IOPS, Influx IO Included 12000 IOPS, 
					Influx IO Included 16000 IOPS.`,
			},
			"deployment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ClusterDeploymentType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `Specifies the type of cluster to create. This field is forbidden for InfluxDB V3 clusters
					(when using an InfluxDB V3 db parameter group).`,
			},
			"engine_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The database engine type of the DB cluster.`,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The endpoint used to connect to InfluxDB. The default InfluxDB port is 8086.`,
			},
			"failover_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.FailoverMode](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `Specifies the behavior of failure recovery when the primary node of the cluster
					fails.`,
			},
			names.AttrID: framework.IDAttribute(),
			"influx_auth_parameters_secret_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The Amazon Resource Name (ARN) of the AWS Secrets Manager secret containing the 
					initial InfluxDB authorization parameters. For InfluxDB V2 clusters, the secret value is a JSON
					formatted key-value pair holding InfluxDB authorization values: organization, bucket,
					username, and password. For InfluxDB V3 clusters, the secret contains the InfluxDB admin token.`,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 40),
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$"),
						"",
					),
				},
				Description: `The name that uniquely identifies the DB cluster when interacting with the 
					Amazon Timestream for InfluxDB API and CLI commands. This name will also be a 
					prefix included in the endpoint. DB cluster names must be unique per customer 
					and per region.`,
			},
			"network_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.NetworkType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `Specifies whether the networkType of the Timestream for InfluxDB cluster is 
					IPV4, which can communicate over IPv4 protocol only, or DUAL, which can communicate 
					over both IPv4 and IPv6 protocols.`,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"organization": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				Description: `Name of the initial organization for the initial admin user in InfluxDB. An
					InfluxDB organization is a workspace for a group of users. Along with bucket, username,
					and password, this argument will be stored in the secret referred to by the
					influx_auth_parameters_secret_arn attribute. This field is forbidden for InfluxDB V3 clusters
					(when using an InfluxDB V3 db parameter group).`,
			},
			names.AttrPassword: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 64),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9]+$"), ""),
				},
				Description: `Password of the initial admin user created in InfluxDB. This password will
					allow you to access the InfluxDB UI to perform various administrative tasks and
					also use the InfluxDB CLI to create an operator token. Along with bucket, username,
					and organization, this argument will be stored in the secret referred to by the
					influx_auth_parameters_secret_arn attribute. This field is forbidden for InfluxDB V3 clusters
					(when using an InfluxDB V3 db parameter group) as the AWS API rejects it.`,
			},
			names.AttrPort: schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.Between(1024, 65535),
					int32validator.NoneOf(2375, 2376, 7788, 7789, 7790, 7791, 7792, 7793, 7794, 7795, 7796, 7797, 7798, 7799, 8090, 51678, 51679, 51680),
				},
				Description: `The port number on which InfluxDB accepts connections.`,
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: `Configures the Timestream for InfluxDB cluster with a public IP to facilitate access.`,
			},
			"reader_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The endpoint used to connect to the Timestream for InfluxDB cluster for 
					read-only operations.`,
			},
			names.AttrUsername: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-Z]([a-zA-Z0-9]*(-[a-zA-Z0-9]+)*)?$"),
						`Must start with a letter and can't end with a hyphen or contain two 
						consecutive hyphens`,
					),
				},
				Description: `Username of the initial admin user created in InfluxDB. Must start with a letter
					and can't end with a hyphen or contain two consecutive hyphens. This username will allow
					you to access the InfluxDB UI to perform various administrative tasks and also use the
					InfluxDB CLI to create an operator token. Along with bucket, organization, and password,
					this argument will be stored in the secret referred to by the influx_auth_parameters_secret_arn
					attribute. This field is forbidden for InfluxDB V3 clusters (when using an InfluxDB V3 db parameter group).`,
			},
			names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 5),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexache.MustCompile("^sg-[a-z0-9]+$"), ""),
					),
				},
				Description: `A list of VPC security group IDs to associate with the Timestream for InfluxDB cluster.`,
			},
			"vpc_subnet_ids": schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 3),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexache.MustCompile("^subnet-[a-z0-9]+$"), ""),
					),
				},
				Description: `A list of VPC subnet IDs to associate with the DB cluster. Provide at least 
					two VPC subnet IDs in different availability zones when deploying with a Multi-AZ standby.`,
			},
		},
		Blocks: map[string]schema.Block{
			"log_delivery_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dbClusterLogDeliveryConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: `Configuration for sending InfluxDB engine logs to a specified S3 bucket.`,
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dbClusterS3ConfigurationData](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucketName: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(3, 63),
											stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-z]+[0-9a-z\\.\\-]*[0-9a-z]+$"), ""),
										},
										Description: `The name of the S3 bucket to deliver logs to.`,
									},
									names.AttrEnabled: schema.BoolAttribute{
										Required:    true,
										Description: `Indicates whether log delivery to the S3 bucket is enabled.`,
									},
								},
							},
							Description: `Configuration for S3 bucket log delivery.`,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func dbClusterDBParameterGroupIdentifierReplaceIf(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var plan, state dbClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the DBParameterGroupIdentifier is being removed, the cluster must be recreated.
	dbParameterGroupIdentifierRemoved := !state.DBParameterGroupIdentifier.IsNull() && plan.DBParameterGroupIdentifier.IsNull()

	resp.RequiresReplace = dbParameterGroupIdentifierRemoved
}

func (r *dbClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var plan dbClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := timestreaminfluxdb.CreateDbClusterInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDbCluster(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDBCluster, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDBCluster, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	if out.DbClusterId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDBCluster, plan.Name.String(), nil),
			errors.New("received response with nil DbClusterId").Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringToFramework(ctx, out.DbClusterId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitDBClusterCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), out.DbClusterId)...)
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForCreation, ResNameDBCluster, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dbClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var state dbClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findDBClusterByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionReading, ResNameDBCluster, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dbClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var plan, state dbClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := timestreaminfluxdb.UpdateDbClusterInput{
			DbClusterId: plan.ID.ValueStringPointer(),
		}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDbCluster(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDBCluster, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		output, err := waitDBClusterUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForUpdate, ResNameDBCluster, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dbClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var state dbClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &timestreaminfluxdb.DeleteDbClusterInput{
		DbClusterId: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteDbCluster(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionDeleting, ResNameDBCluster, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDBClusterDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForDeletion, ResNameDBCluster, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func isParameterGroupV3(ctx context.Context, conn *timestreaminfluxdb.Client, parameterGroupID string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	out, err := findDBParameterGroupByID(ctx, conn, parameterGroupID)

	if retry.NotFound(err) {
		return false, diags
	}

	if err != nil {
		diags.AddWarning(
			"Unable to query parameter group",
			"Could not determine parameter group type. Validation will be skipped.",
		)
		return false, diags
	}

	switch out.Parameters.(type) {
	case *awstypes.ParametersMemberInfluxDBv3Core:
		return true, diags
	case *awstypes.ParametersMemberInfluxDBv3Enterprise:
		return true, diags
	default:
		return false, diags
	}
}

func (r *dbClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data dbClusterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isNullOrUnknown := func(val attr.Value) bool {
		return val.IsNull() || val.IsUnknown()
	}

	if !isNullOrUnknown(data.AllocatedStorage) {
		switch v := data.AllocatedStorage.ValueInt64(); {
		case v > math.MaxInt32:
			resp.Diagnostics.AddError(
				"Invalid value for allocated_storage",
				"allocated_storage was greater than the maximum allowed value for int32",
			)
			return
		case v < math.MinInt32:
			resp.Diagnostics.AddError(
				"Invalid value for allocated_storage",
				"allocated_storage was less than the minimum allowed value for int32",
			)
			return
		}
	}

	hasV2Fields := !isNullOrUnknown(data.AllocatedStorage) ||
		!isNullOrUnknown(data.Bucket) ||
		!isNullOrUnknown(data.DeploymentType) ||
		!isNullOrUnknown(data.Organization) ||
		!isNullOrUnknown(data.Password) ||
		!isNullOrUnknown(data.Username)

	var isV3Cluster bool
	if !isNullOrUnknown(data.DBParameterGroupIdentifier) {
		meta := r.Meta()
		if meta == nil {
			return
		}
		paramGroupID := data.DBParameterGroupIdentifier.ValueString()
		isV3, diags := isParameterGroupV3(ctx, meta.TimestreamInfluxDBClient(ctx), paramGroupID)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		isV3Cluster = isV3

		if !hasV2Fields && !isV3Cluster {
			resp.Diagnostics.AddAttributeError(
				path.Root("db_parameter_group_identifier"),
				"Invalid Parameter Group Type",
				"An InfluxDB V2 parameter group requires InfluxDB V2 fields (allocated_storage, bucket, deployment_type, organization, password, username). Use an InfluxDB V3 parameter group or provide the V2 fields.",
			)
		}
	}

	if isV3Cluster {
		for _, v := range []struct {
			val  attr.Value
			path string
		}{
			{data.AllocatedStorage, names.AttrAllocatedStorage},
			{data.Bucket, names.AttrBucket},
			{data.DeploymentType, "deployment_type"},
			{data.Organization, "organization"},
			{data.Password, names.AttrPassword},
			{data.Username, names.AttrUsername},
		} {
			if !isNullOrUnknown(v.val) {
				resp.Diagnostics.AddAttributeError(
					path.Root(v.path),
					"Invalid Configuration for InfluxDB V3",
					v.path+" must not be set when using an InfluxDB V3 db parameter group",
				)
			}
		}
	} else {
		for _, v := range []struct {
			val  attr.Value
			path string
		}{
			{data.AllocatedStorage, names.AttrAllocatedStorage},
			{data.Bucket, names.AttrBucket},
			{data.Organization, "organization"},
			{data.Password, names.AttrPassword},
			{data.Username, names.AttrUsername},
		} {
			if isNullOrUnknown(v.val) {
				resp.Diagnostics.AddAttributeError(
					path.Root(v.path),
					"Missing Required Configuration for InfluxDB V2",
					v.path+" is required for InfluxDB V2 clusters",
				)
			}
		}
	}
}

func (r *dbClusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.Plan.Raw.IsNull() {
		var data dbClusterResourceModel
		resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var isV3Cluster bool
		if !data.DBParameterGroupIdentifier.IsNull() {
			isV3, diags := isParameterGroupV3(ctx, r.Meta().TimestreamInfluxDBClient(ctx), data.DBParameterGroupIdentifier.ValueString())
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			isV3Cluster = isV3
		}

		if !isV3Cluster && data.DeploymentType.IsUnknown() {
			resp.Plan.SetAttribute(ctx, path.Root("deployment_type"), fwtypes.StringEnumValue(awstypes.ClusterDeploymentTypeMultiNodeReadReplicas))
		}
	}
}

func waitDBClusterCreated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ClusterStatusCreating),
		Target:                    enum.Slice(awstypes.ClusterStatusAvailable),
		Refresh:                   statusDBCluster(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBClusterUpdated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(string(awstypes.ClusterStatusUpdating), string(awstypes.StatusUpdatingInstanceType)),
		Target:                    enum.Slice(awstypes.ClusterStatusAvailable),
		Refresh:                   statusDBCluster(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBClusterDeleted(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbClusterOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClusterStatusDeleting, awstypes.ClusterStatusDeleted),
		Target:  []string{},
		Refresh: statusDBCluster(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbClusterOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDBCluster(conn *timestreaminfluxdb.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findDBClusterByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func findDBClusterByID(ctx context.Context, conn *timestreaminfluxdb.Client, id string) (*timestreaminfluxdb.GetDbClusterOutput, error) {
	in := timestreaminfluxdb.GetDbClusterInput{
		DbClusterId: aws.String(id),
	}

	return findDBCluster(ctx, conn, &in)
}

func findDBCluster(ctx context.Context, conn *timestreaminfluxdb.Client, in *timestreaminfluxdb.GetDbClusterInput) (*timestreaminfluxdb.GetDbClusterOutput, error) {
	out, err := conn.GetDbCluster(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func findDBParameterGroupByID(ctx context.Context, conn *timestreaminfluxdb.Client, id string) (*timestreaminfluxdb.GetDbParameterGroupOutput, error) {
	in := timestreaminfluxdb.GetDbParameterGroupInput{
		Identifier: aws.String(id),
	}

	return findDBParameterGroup(ctx, conn, &in)
}

func findDBParameterGroup(ctx context.Context, conn *timestreaminfluxdb.Client, in *timestreaminfluxdb.GetDbParameterGroupInput) (*timestreaminfluxdb.GetDbParameterGroupOutput, error) {
	out, err := conn.GetDbParameterGroup(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type dbClusterResourceModel struct {
	framework.WithRegionModel
	AllocatedStorage              types.Int64                                                            `tfsdk:"allocated_storage"`
	ARN                           types.String                                                           `tfsdk:"arn"`
	Bucket                        types.String                                                           `tfsdk:"bucket"`
	DBInstanceType                fwtypes.StringEnum[awstypes.DbInstanceType]                            `tfsdk:"db_instance_type"`
	DBParameterGroupIdentifier    types.String                                                           `tfsdk:"db_parameter_group_identifier"`
	DBStorageType                 fwtypes.StringEnum[awstypes.DbStorageType]                             `tfsdk:"db_storage_type"`
	DeploymentType                fwtypes.StringEnum[awstypes.ClusterDeploymentType]                     `tfsdk:"deployment_type"`
	EngineType                    types.String                                                           `tfsdk:"engine_type"`
	Endpoint                      types.String                                                           `tfsdk:"endpoint"`
	FailoverMode                  fwtypes.StringEnum[awstypes.FailoverMode]                              `tfsdk:"failover_mode"`
	ID                            types.String                                                           `tfsdk:"id"`
	InfluxAuthParametersSecretARN types.String                                                           `tfsdk:"influx_auth_parameters_secret_arn"`
	LogDeliveryConfiguration      fwtypes.ListNestedObjectValueOf[dbClusterLogDeliveryConfigurationData] `tfsdk:"log_delivery_configuration"`
	Name                          types.String                                                           `tfsdk:"name"`
	NetworkType                   fwtypes.StringEnum[awstypes.NetworkType]                               `tfsdk:"network_type"`
	Organization                  types.String                                                           `tfsdk:"organization"`
	Password                      types.String                                                           `tfsdk:"password"`
	Port                          types.Int32                                                            `tfsdk:"port"`
	PubliclyAccessible            types.Bool                                                             `tfsdk:"publicly_accessible"`
	ReaderEndpoint                types.String                                                           `tfsdk:"reader_endpoint"`
	Tags                          tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value                                                         `tfsdk:"timeouts"`
	Username                      types.String                                                           `tfsdk:"username"`
	VPCSecurityGroupIDs           fwtypes.SetOfString                                                    `tfsdk:"vpc_security_group_ids"`
	VPCSubnetIDs                  fwtypes.SetOfString                                                    `tfsdk:"vpc_subnet_ids"`
}

type dbClusterLogDeliveryConfigurationData struct {
	S3Configuration fwtypes.ListNestedObjectValueOf[dbClusterS3ConfigurationData] `tfsdk:"s3_configuration"`
}

type dbClusterS3ConfigurationData struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}
