// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_timestreaminfluxdb_db_instance", name="Db Instance")
// @Tags(identifierAttribute="arn")
func newResourceDBInstance(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDBInstance{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	// If not provided, CreateDbInstance will use the below default values
	// for bucket and organization. These values need to be set in Terraform
	// because GetDbInstance won't return them.
	DefaultBucketValue       = names.AttrBucket
	DefaultOrganizationValue = "organization"
	DefaultUsernameValue     = "admin"
	ResNameDBInstance        = "DB Instance"
)

type resourceDBInstance struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDBInstance) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_timestreaminfluxdb_db_instance"
}

func (r *resourceDBInstance) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAllocatedStorage: schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(20),
					int64validator.AtMost(16384),
				},
				Description: `The amount of storage to allocate for your DB storage type in GiB (gibibytes).`,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAvailabilityZone: schema.StringAttribute{
				Computed:    true,
				Description: `The Availability Zone in which the DB instance resides.`,
			},
			names.AttrBucket: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(DefaultBucketValue),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L768
						regexache.MustCompile("^[^_][^\"]*$"),
						"",
					),
				},
				Description: `The name of the initial InfluxDB bucket. All InfluxDB data is stored in a bucket. 
					A bucket combines the concept of a database and a retention period (the duration of time 
					that each data point persists). A bucket belongs to an organization.`,
			},
			"db_instance_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(awstypes.DbInstanceTypeDbInfluxMedium),
						string(awstypes.DbInstanceTypeDbInfluxLarge),
						string(awstypes.DbInstanceTypeDbInfluxXlarge),
						string(awstypes.DbInstanceTypeDbInflux2xlarge),
						string(awstypes.DbInstanceTypeDbInflux4xlarge),
						string(awstypes.DbInstanceTypeDbInflux8xlarge),
						string(awstypes.DbInstanceTypeDbInflux12xlarge),
						string(awstypes.DbInstanceTypeDbInflux16xlarge),
					),
				},
				Description: `The Timestream for InfluxDB DB instance type to run InfluxDB on.`,
			},
			"db_parameter_group_identifier": schema.StringAttribute{
				Optional: true,
				// Once a parameter group is associated with a DB instance, it cannot be removed.
				// Therefore, if db_parameter_group_identifier is removed, a replace of the DB instance
				// is necessary.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						statementReplaceIf, "Replace db_parameter_group_identifier diff", "Replace db_parameter_group_identifier diff",
					),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L1390
						regexache.MustCompile("^[a-zA-Z0-9]+$"),
						"",
					),
				},
				Description: `The id of the DB parameter group assigned to your DB instance.`,
			},
			"db_storage_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.DbStorageTypeInfluxIoIncludedT1)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(awstypes.DbStorageTypeInfluxIoIncludedT1),
						string(awstypes.DbStorageTypeInfluxIoIncludedT2),
						string(awstypes.DbStorageTypeInfluxIoIncludedT3),
					),
				},
				Description: `The Timestream for InfluxDB DB storage type to read and write InfluxDB data. 
					You can choose between 3 different types of provisioned Influx IOPS included storage according 
					to your workloads requirements: Influx IO Included 3000 IOPS, Influx IO Included 12000 IOPS, 
					Influx IO Included 16000 IOPS.`,
			},
			"deployment_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(awstypes.DeploymentTypeSingleAz)),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(awstypes.DeploymentTypeSingleAz),
						string(awstypes.DeploymentTypeWithMultiazStandby),
					),
				},
				Description: `Specifies whether the DB instance will be deployed as a standalone instance or 
					with a Multi-AZ standby for high availability.`,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed:    true,
				Description: `The endpoint used to connect to InfluxDB. The default InfluxDB port is 8086.`,
			},
			names.AttrID: framework.IDAttribute(),
			"influx_auth_parameters_secret_arn": schema.StringAttribute{
				Computed: true,
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
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(40),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L1215
						regexache.MustCompile("^[a-zA-z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$"),
						"",
					),
				},
				Description: `The name that uniquely identifies the DB instance when interacting with the 
					Amazon Timestream for InfluxDB API and CLI commands. This name will also be a 
					prefix included in the endpoint. DB instance names must be unique per customer 
					and per region.`,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"organization": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(DefaultOrganizationValue),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(64),
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
					stringvalidator.LengthAtLeast(8),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9]+$"), ""),
				},
				Description: `The password of the initial admin user created in InfluxDB. This password will 
					allow you to access the InfluxDB UI to perform various administrative tasks and 
					also use the InfluxDB CLI to create an operator token. These attributes will be 
					stored in a Secret created in AWS SecretManager in your account.`,
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
				Description: `Configures the DB instance with a public IP to facilitate access.`,
			},
			"secondary_availability_zone": schema.StringAttribute{
				Computed: true,
				Description: `The Availability Zone in which the standby instance is located when deploying 
					with a MultiAZ standby instance.`,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: `The status of the DB instance.`,
			},
			names.AttrUsername: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(DefaultUsernameValue),
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
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(5),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexache.MustCompile("^sg-[a-z0-9]+$"), ""),
					),
				},
				Description: `A list of VPC security group IDs to associate with the DB instance.`,
			},
			"vpc_subnet_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(3),
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
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: `Configuration for sending InfluxDB engine logs to a specified S3 bucket.`,
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								names.AttrBucketName: schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(3),
										stringvalidator.LengthAtMost(63),
										stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-z]+[0-9a-z\\.\\-]*[0-9a-z]+$"), ""),
									},
									Description: `The name of the S3 bucket to deliver logs to.`,
								},
								names.AttrEnabled: schema.BoolAttribute{
									Required:    true,
									Description: `Indicates whether log delivery to the S3 bucket is enabled.`,
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

func statementReplaceIf(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	var plan, state resourceDBInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbParameterGroupIdentifierRemoved := (!state.DBParameterGroupIdentifier.IsNull() && plan.DBParameterGroupIdentifier.IsNull())

	resp.RequiresReplace = dbParameterGroupIdentifierRemoved
}

func (r *resourceDBInstance) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var plan resourceDBInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &timestreaminfluxdb.CreateDbInstanceInput{
		AllocatedStorage:    aws.Int32(int32(plan.AllocatedStorage.ValueInt64())),
		DbInstanceType:      awstypes.DbInstanceType(plan.DBInstanceType.ValueString()),
		Name:                aws.String(plan.Name.ValueString()),
		Password:            aws.String(plan.Password.ValueString()),
		VpcSecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, plan.VPCSecurityGroupIDs),
		VpcSubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, plan.VPCSubnetIDs),
		Tags:                getTagsIn(ctx),
	}
	if !plan.Bucket.IsNull() {
		in.Bucket = aws.String(plan.Bucket.ValueString())
	}
	if !plan.DBParameterGroupIdentifier.IsNull() {
		in.DbParameterGroupIdentifier = aws.String(plan.DBParameterGroupIdentifier.ValueString())
	}
	if !plan.DBStorageType.IsNull() {
		in.DbStorageType = awstypes.DbStorageType(plan.DBStorageType.ValueString())
	}
	if !plan.DeploymentType.IsNull() {
		in.DeploymentType = awstypes.DeploymentType(plan.DeploymentType.ValueString())
	}
	if !plan.LogDeliveryConfiguration.IsNull() {
		var tfList []logDeliveryConfigurationData
		resp.Diagnostics.Append(plan.LogDeliveryConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.LogDeliveryConfiguration = expandLogDeliveryConfiguration(tfList)
	}
	if !plan.Organization.IsNull() {
		in.Organization = aws.String(plan.Organization.ValueString())
	}
	if !plan.PubliclyAccessible.IsNull() {
		in.PubliclyAccessible = aws.Bool(plan.PubliclyAccessible.ValueBool())
	}
	if !plan.Username.IsNull() {
		in.Username = aws.String(plan.Username.ValueString())
	}

	out, err := conn.CreateDbInstance(ctx, in)
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

	// Computed attributes
	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = flex.StringToFramework(ctx, out.Id)
	plan.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDBInstanceCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForCreation, ResNameDBInstance, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	readOut, err := findDBInstanceByID(ctx, conn, plan.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Computed attributes only set after resource is finished creating
	plan.Endpoint = flex.StringToFramework(ctx, readOut.Endpoint)
	plan.InfluxAuthParametersSecretARN = flex.StringToFramework(ctx, readOut.InfluxAuthParametersSecretArn)
	plan.Status = flex.StringToFramework(ctx, (*string)(&readOut.Status))
	plan.SecondaryAvailabilityZone = flex.StringToFramework(ctx, readOut.SecondaryAvailabilityZone)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDBInstance) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	var state resourceDBInstanceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDBInstanceByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
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

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AllocatedStorage = flex.Int32ToFramework(ctx, out.AllocatedStorage)
	state.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	state.DBInstanceType = flex.StringToFramework(ctx, (*string)(&out.DbInstanceType))
	state.DBParameterGroupIdentifier = flex.StringToFramework(ctx, out.DbParameterGroupIdentifier)
	state.DBStorageType = flex.StringToFramework(ctx, (*string)(&out.DbStorageType))
	state.DeploymentType = flex.StringToFramework(ctx, (*string)(&out.DeploymentType))
	state.Endpoint = flex.StringToFramework(ctx, out.Endpoint)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.InfluxAuthParametersSecretARN = flex.StringToFramework(ctx, out.InfluxAuthParametersSecretArn)
	logDeliveryConfiguration, d := flattenLogDeliveryConfiguration(ctx, out.LogDeliveryConfiguration)
	resp.Diagnostics.Append(d...)
	state.LogDeliveryConfiguration = logDeliveryConfiguration
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.PubliclyAccessible = flex.BoolToFramework(ctx, out.PubliclyAccessible)
	state.SecondaryAvailabilityZone = flex.StringToFramework(ctx, out.SecondaryAvailabilityZone)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))
	state.VPCSecurityGroupIDs = flex.FlattenFrameworkStringValueSet[string](ctx, out.VpcSecurityGroupIds)
	state.VPCSubnetIDs = flex.FlattenFrameworkStringValueSet[string](ctx, out.VpcSubnetIds)

	// timestreaminfluxdb.GetDbInstance will not return InfluxDB managed attributes, like username,
	// bucket, organization, or password. All of these attributes are stored in a secret indicated by
	// out.InfluxAuthParametersSecretArn. To support importing, these attributes must be read from the
	// secret.
	secretsConn := r.Meta().SecretsManagerClient(ctx)
	secretsOut, err := secretsConn.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: out.InfluxAuthParametersSecretArn,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	secrets := make(map[string]string)
	if err := json.Unmarshal([]byte(aws.ToString(secretsOut.SecretString)), &secrets); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if username, ok := secrets[names.AttrUsername]; ok {
		state.Username = flex.StringValueToFramework[string](ctx, username)
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if password, ok := secrets[names.AttrPassword]; ok {
		state.Password = flex.StringValueToFramework[string](ctx, password)
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if organization, ok := secrets["organization"]; ok {
		state.Organization = flex.StringValueToFramework[string](ctx, organization)
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if bucket, ok := secrets[names.AttrBucket]; ok {
		state.Bucket = flex.StringValueToFramework[string](ctx, bucket)
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	tags, err := listTags(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	setTagsOut(ctx, Tags(tags))

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

	// Only fields without RequireReplace() will cause an update.
	// Any other field changes will cause the resource to be destroyed and recreated.
	// for aws_timestreaminfluxdb_db_instance this is tags, log_delivery_configuration, and
	// db_parameter_group_identifier.
	if !plan.DBParameterGroupIdentifier.Equal(state.DBParameterGroupIdentifier) ||
		!plan.LogDeliveryConfiguration.Equal(state.LogDeliveryConfiguration) {
		in := &timestreaminfluxdb.UpdateDbInstanceInput{
			Identifier: aws.String(plan.ID.ValueString()),
		}

		if !plan.DBParameterGroupIdentifier.IsNull() && !plan.DBParameterGroupIdentifier.Equal(state.DBParameterGroupIdentifier) {
			in.DbParameterGroupIdentifier = aws.String(plan.DBParameterGroupIdentifier.ValueString())
		}

		if !plan.LogDeliveryConfiguration.IsNull() && !plan.LogDeliveryConfiguration.Equal(state.LogDeliveryConfiguration) {
			var tfList []logDeliveryConfigurationData
			resp.Diagnostics.Append(plan.LogDeliveryConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.LogDeliveryConfiguration = expandLogDeliveryConfiguration(tfList)
		}

		out, err := conn.UpdateDbInstance(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDBInstance, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDBInstance, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDBInstanceUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForUpdate, ResNameDBInstance, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Update status to current status
	readOut, err := findDBInstanceByID(ctx, conn, plan.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDBInstance, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	// Setting computed attributes
	plan.ARN = flex.StringToFramework(ctx, readOut.Arn)
	plan.AvailabilityZone = flex.StringToFramework(ctx, readOut.AvailabilityZone)
	plan.DBStorageType = flex.StringToFramework(ctx, (*string)(&readOut.DbStorageType))
	plan.DeploymentType = flex.StringToFramework(ctx, (*string)(&readOut.DeploymentType))
	plan.Endpoint = flex.StringToFramework(ctx, readOut.Endpoint)
	plan.ID = flex.StringToFramework(ctx, readOut.Id)
	plan.InfluxAuthParametersSecretARN = flex.StringToFramework(ctx, readOut.InfluxAuthParametersSecretArn)
	plan.SecondaryAvailabilityZone = flex.StringToFramework(ctx, readOut.SecondaryAvailabilityZone)
	plan.Status = flex.StringToFramework(ctx, (*string)(&readOut.Status))

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
		Identifier: aws.String(state.ID.ValueString()),
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

func (r *resourceDBInstance) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}
func (r *resourceDBInstance) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitDBInstanceCreated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.CreateDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(string(awstypes.StatusCreating), string(awstypes.StatusUpdating), string(awstypes.StatusModifying)),
		Target:                    enum.Slice(awstypes.StatusAvailable),
		Refresh:                   statusDBInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.CreateDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBInstanceUpdated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.UpdateDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(string(awstypes.StatusModifying), string(awstypes.StatusUpdating)),
		Target:                    enum.Slice(string(awstypes.StatusAvailable)),
		Refresh:                   statusDBInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.UpdateDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDBInstanceDeleted(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.DeleteDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(string(awstypes.StatusDeleting), string(awstypes.StatusModifying), string(awstypes.StatusUpdating), string(awstypes.StatusAvailable)),
		Target:       enum.Slice[string](),
		Refresh:      statusDBInstance(ctx, conn, id),
		Timeout:      timeout,
		Delay:        30 * time.Second,
		PollInterval: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.DeleteDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDBInstance(ctx context.Context, conn *timestreaminfluxdb.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
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
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenLogDeliveryConfiguration(ctx context.Context, apiObject *awstypes.LogDeliveryConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: logDeliveryConfigrationAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}
	s3Configuration, d := flattenS3Configuration(ctx, apiObject.S3Configuration)
	diags.Append(d...)
	obj := map[string]attr.Value{
		"s3_configuration": s3Configuration,
	}
	objVal, d := types.ObjectValue(logDeliveryConfigrationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenS3Configuration(ctx context.Context, apiObject *awstypes.S3Configuration) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: s3ConfigurationAttrTypes}

	if apiObject == nil {
		return types.ObjectNull(elemType.AttrTypes), diags
	}

	obj := map[string]attr.Value{
		names.AttrBucketName: flex.StringValueToFramework(ctx, *apiObject.BucketName),
		names.AttrEnabled:    flex.BoolToFramework(ctx, apiObject.Enabled),
	}
	objVal, d := types.ObjectValue(s3ConfigurationAttrTypes, obj)
	diags.Append(d...)

	return objVal, diags
}

func expandLogDeliveryConfiguration(tfList []logDeliveryConfigurationData) *awstypes.LogDeliveryConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.LogDeliveryConfiguration{
		S3Configuration: expandS3Configuration(tfObj.S3Configuration),
	}
	return apiObject
}

func expandS3Configuration(tfObj s3ConfigurationData) *awstypes.S3Configuration {
	apiObject := &awstypes.S3Configuration{
		BucketName: aws.String(tfObj.BucketName.ValueString()),
		Enabled:    aws.Bool(tfObj.Enabled.ValueBool()),
	}
	return apiObject
}

type resourceDBInstanceData struct {
	AllocatedStorage              types.Int64    `tfsdk:"allocated_storage"`
	ARN                           types.String   `tfsdk:"arn"`
	AvailabilityZone              types.String   `tfsdk:"availability_zone"`
	Bucket                        types.String   `tfsdk:"bucket"`
	DBInstanceType                types.String   `tfsdk:"db_instance_type"`
	DBParameterGroupIdentifier    types.String   `tfsdk:"db_parameter_group_identifier"`
	DBStorageType                 types.String   `tfsdk:"db_storage_type"`
	DeploymentType                types.String   `tfsdk:"deployment_type"`
	Endpoint                      types.String   `tfsdk:"endpoint"`
	ID                            types.String   `tfsdk:"id"`
	InfluxAuthParametersSecretARN types.String   `tfsdk:"influx_auth_parameters_secret_arn"`
	LogDeliveryConfiguration      types.List     `tfsdk:"log_delivery_configuration"`
	Name                          types.String   `tfsdk:"name"`
	Organization                  types.String   `tfsdk:"organization"`
	Password                      types.String   `tfsdk:"password"`
	PubliclyAccessible            types.Bool     `tfsdk:"publicly_accessible"`
	SecondaryAvailabilityZone     types.String   `tfsdk:"secondary_availability_zone"`
	Status                        types.String   `tfsdk:"status"`
	Tags                          types.Map      `tfsdk:"tags"`
	TagsAll                       types.Map      `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
	Username                      types.String   `tfsdk:"username"`
	VPCSecurityGroupIDs           types.Set      `tfsdk:"vpc_security_group_ids"`
	VPCSubnetIDs                  types.Set      `tfsdk:"vpc_subnet_ids"`
}

type logDeliveryConfigurationData struct {
	S3Configuration s3ConfigurationData `tfsdk:"s3_configuration"`
}

type s3ConfigurationData struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

var logDeliveryConfigrationAttrTypes = map[string]attr.Type{
	"s3_configuration": types.ObjectType{AttrTypes: s3ConfigurationAttrTypes},
}

var s3ConfigurationAttrTypes = map[string]attr.Type{
	names.AttrBucketName: types.StringType,
	names.AttrEnabled:    types.BoolType,
}
