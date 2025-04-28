// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_timestreaminfluxdb_db_instance", name="DB Instance")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb;timestreaminfluxdb.GetDbInstanceOutput")
// @Testing(importIgnore="bucket;username;organization;password")
func newResourceDBInstance(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDBInstance{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDBInstance = "DB Instance"
)

type resourceDBInstance struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (r *resourceDBInstance) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAllocatedStorage: schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(20, 16384),
				},
				Description: `The amount of storage to allocate for your DB storage type in GiB (gibibytes).`,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed:    true,
				Description: `The Availability Zone in which the DB instance resides.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrBucket: schema.StringAttribute{
				Required: true,
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
				Description: `The name of the initial InfluxDB bucket. All InfluxDB data is stored in a bucket. 
					A bucket combines the concept of a database and a retention period (the duration of time 
					that each data point persists). A bucket belongs to an organization.`,
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
						dbParameterGroupIdentifierReplaceIf, "Replace db_parameter_group_identifier diff", "Replace db_parameter_group_identifier diff",
					),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-Z0-9]+$"),
						"",
					),
				},
				Description: `The id of the DB parameter group assigned to your DB instance.`,
			},
			"db_storage_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DbStorageType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The Timestream for InfluxDB DB storage type to read and write InfluxDB data. 
					You can choose between 3 different types of provisioned Influx IOPS included storage according 
					to your workloads requirements: Influx IO Included 3000 IOPS, Influx IO Included 12000 IOPS, 
					Influx IO Included 16000 IOPS.`,
			},
			"deployment_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DeploymentType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `Specifies whether the DB instance will be deployed as a standalone instance or 
					with a Multi-AZ standby for high availability.`,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The endpoint used to connect to InfluxDB. The default InfluxDB port is 8086.`,
			},
			names.AttrID: framework.IDAttribute(),
			"influx_auth_parameters_secret_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: `The Amazon Resource Name (ARN) of the AWS Secrets Manager secret containing the 
					initial InfluxDB authorization parameters. The secret value is a JSON formatted 
					key-value pair holding InfluxDB authorization values: organization, bucket, 
					username, and password.`,
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
				Description: `The name that uniquely identifies the DB instance when interacting with the 
					Amazon Timestream for InfluxDB API and CLI commands. This name will also be a 
					prefix included in the endpoint. DB instance names must be unique per customer 
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
				Description: `Specifies whether the networkType of the Timestream for InfluxDB instance is 
					IPV4, which can communicate over IPv4 protocol only, or DUAL, which can communicate 
					over both IPv4 and IPv6 protocols.`,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"organization": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				Description: `The name of the initial organization for the initial admin user in InfluxDB. An 
					InfluxDB organization is a workspace for a group of users.`,
			},
			names.AttrPassword: schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(8, 64),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9]+$"), ""),
				},
				Description: `The password of the initial admin user created in InfluxDB. This password will 
					allow you to access the InfluxDB UI to perform various administrative tasks and 
					also use the InfluxDB CLI to create an operator token. These attributes will be 
					stored in a Secret created in AWS SecretManager in your account.`,
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
				Description: `Configures the DB instance with a public IP to facilitate access.`,
			},
			"secondary_availability_zone": schema.StringAttribute{
				Computed: true,
				Description: `The Availability Zone in which the standby instance is located when deploying 
					with a MultiAZ standby instance.`,
			},
			names.AttrUsername: schema.StringAttribute{
				Required: true,
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
				Description: `The username of the initial admin user created in InfluxDB. 
					Must start with a letter and can't end with a hyphen or contain two 
					consecutive hyphens. For example, my-user1. This username will allow 
					you to access the InfluxDB UI to perform various administrative tasks 
					and also use the InfluxDB CLI to create an operator token. These 
					attributes will be stored in a Secret created in Amazon Secrets 
					Manager in your account`,
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
				Description: `A list of VPC security group IDs to associate with the DB instance.`,
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
				Description: `A list of VPC subnet IDs to associate with the DB instance. Provide at least 
					two VPC subnet IDs in different availability zones when deploying with a Multi-AZ standby.`,
			},
		},
		Blocks: map[string]schema.Block{
			"log_delivery_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logDeliveryConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: `Configuration for sending InfluxDB engine logs to a specified S3 bucket.`,
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationData](ctx),
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

func dbParameterGroupIdentifierReplaceIf(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var plan, state resourceDBInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbParameterGroupIdentifierRemoved := !state.DBParameterGroupIdentifier.IsNull() && plan.DBParameterGroupIdentifier.IsNull()

	resp.RequiresReplace = dbParameterGroupIdentifierRemoved
}

func (r *resourceDBInstance) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var plan resourceDBInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := timestreaminfluxdb.CreateDbInstanceInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateDbInstance(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDBInstance, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDBInstance, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.ID = fwflex.StringToFramework(ctx, out.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitDBInstanceCreated(ctx, conn, state.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), out.Id)...)
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForCreation, ResNameDBInstance, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// flatten using legacy since this computed output may be null
	state.SecondaryAvailabilityZone = fwflex.StringToFrameworkLegacy(ctx, output.SecondaryAvailabilityZone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDBInstance) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var state resourceDBInstanceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findDBInstanceByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// flatten using legacy since this computed output may be null
	state.SecondaryAvailabilityZone = fwflex.StringToFrameworkLegacy(ctx, output.SecondaryAvailabilityZone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDBInstance) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var plan, state resourceDBInstanceData
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
		input := timestreaminfluxdb.UpdateDbInstanceInput{
			Identifier: plan.ID.ValueStringPointer(),
		}
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDbInstance(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDBInstance, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		output, err := waitDBInstanceUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForUpdate, ResNameDBInstance, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		plan.SecondaryAvailabilityZone = fwflex.StringToFrameworkLegacy(ctx, output.SecondaryAvailabilityZone)
	} else {
		plan.SecondaryAvailabilityZone = state.SecondaryAvailabilityZone
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDBInstance) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var state resourceDBInstanceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &timestreaminfluxdb.DeleteDbInstanceInput{
		Identifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteDbInstance(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionDeleting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDBInstanceDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForDeletion, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDBInstance) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var allocatedStorage types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(names.AttrAllocatedStorage), &allocatedStorage)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if allocatedStorage.IsNull() || allocatedStorage.IsUnknown() {
		return
	}

	if allocatedStorage.ValueInt64() > math.MaxInt32 {
		resp.Diagnostics.AddError(
			"Invalid value for allocated_storage",
			"allocated_storage was greater than the maximum allowed value for int32",
		)
		return
	}

	if allocatedStorage.ValueInt64() < math.MinInt32 {
		resp.Diagnostics.AddError(
			"Invalid value for allocated_storage",
			"allocated_storage was less than the minimum allowed value for int32",
		)
		return
	}
}

func waitDBInstanceCreated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusCreating),
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusDBInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBInstanceUpdated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusModifying, awstypes.StatusUpdating, awstypes.StatusUpdatingInstanceType, awstypes.StatusUpdatingDeploymentType),
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusDBInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBInstanceDeleted(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.GetDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDeleting, awstypes.StatusDeleted),
		Target:  []string{},
		Refresh: statusDBInstance(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.GetDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDBInstance(ctx context.Context, conn *timestreaminfluxdb.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findDBInstanceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func findDBInstanceByID(ctx context.Context, conn *timestreaminfluxdb.Client, id string) (*timestreaminfluxdb.GetDbInstanceOutput, error) {
	in := &timestreaminfluxdb.GetDbInstanceInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetDbInstance(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceDBInstanceData struct {
	AllocatedStorage              types.Int64                                                   `tfsdk:"allocated_storage"`
	ARN                           types.String                                                  `tfsdk:"arn"`
	AvailabilityZone              types.String                                                  `tfsdk:"availability_zone"`
	Bucket                        types.String                                                  `tfsdk:"bucket"`
	DBInstanceType                fwtypes.StringEnum[awstypes.DbInstanceType]                   `tfsdk:"db_instance_type"`
	DBParameterGroupIdentifier    types.String                                                  `tfsdk:"db_parameter_group_identifier"`
	DBStorageType                 fwtypes.StringEnum[awstypes.DbStorageType]                    `tfsdk:"db_storage_type"`
	DeploymentType                fwtypes.StringEnum[awstypes.DeploymentType]                   `tfsdk:"deployment_type"`
	Endpoint                      types.String                                                  `tfsdk:"endpoint"`
	ID                            types.String                                                  `tfsdk:"id"`
	InfluxAuthParametersSecretARN types.String                                                  `tfsdk:"influx_auth_parameters_secret_arn"`
	LogDeliveryConfiguration      fwtypes.ListNestedObjectValueOf[logDeliveryConfigurationData] `tfsdk:"log_delivery_configuration"`
	Name                          types.String                                                  `tfsdk:"name"`
	NetworkType                   fwtypes.StringEnum[awstypes.NetworkType]                      `tfsdk:"network_type"`
	Organization                  types.String                                                  `tfsdk:"organization"`
	Password                      types.String                                                  `tfsdk:"password"`
	Port                          types.Int32                                                   `tfsdk:"port"`
	PubliclyAccessible            types.Bool                                                    `tfsdk:"publicly_accessible"`
	SecondaryAvailabilityZone     types.String                                                  `tfsdk:"secondary_availability_zone"`
	Tags                          tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                       tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value                                                `tfsdk:"timeouts"`
	Username                      types.String                                                  `tfsdk:"username"`
	VPCSecurityGroupIDs           fwtypes.SetValueOf[types.String]                              `tfsdk:"vpc_security_group_ids"`
	VPCSubnetIDs                  fwtypes.SetValueOf[types.String]                              `tfsdk:"vpc_subnet_ids"`
}

type logDeliveryConfigurationData struct {
	S3Configuration fwtypes.ListNestedObjectValueOf[s3ConfigurationData] `tfsdk:"s3_configuration"`
}

type s3ConfigurationData struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}
