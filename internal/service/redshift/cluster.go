// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_cluster", name="Cluster")
// @Tags(identifierAttribute="arn")
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterCreate,
		ReadWithoutTimeout:   resourceClusterRead,
		UpdateWithoutTimeout: resourceClusterUpdate,
		DeleteWithoutTimeout: resourceClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Update: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allow_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"aqua_configuration_status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(redshift.AquaConfigurationStatus_Values(), false),
				Deprecated:   "This parameter is no longer supported by the AWS API. It will be removed in the next major version of the provider.",
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return true
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automated_snapshot_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtMost(35),
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone_relocation_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexache.MustCompile(`(?i)^[a-z]`), "first character must be a letter"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"cluster_namespace_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"node_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_public_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_revision_number": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.0",
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z_$]+$`), "must contain only lowercase alphanumeric characters, underscores, and dollar signs"),
					validation.StringMatch(regexache.MustCompile(`(?i)^[a-z_]`), "first character must be a letter or underscore"),
				),
			},
			"default_iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enhanced_vpc_routing": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrFinalSnapshotIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]+$`), "must only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"iam_roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"logging": {
				Type: schema.TypeList,
				Deprecated: "Use the aws_redshift_logging resource instead. " +
					"This argument will be removed in a future major version.",
				MaxItems:         1,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"enable": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"log_destination_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(redshift.LogDestinationType_Values(), false),
						},
						"log_exports": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrS3KeyPrefix: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"maintenance_track_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "current",
			},
			"manage_master_password": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"master_password"},
			},
			"manual_snapshot_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.IntBetween(-1, 3653),
			},
			"master_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(8, 64),
					validation.StringMatch(regexache.MustCompile(`^.*[a-z].*`), "must contain at least one lowercase letter"),
					validation.StringMatch(regexache.MustCompile(`^.*[A-Z].*`), "must contain at least one uppercase letter"),
					validation.StringMatch(regexache.MustCompile(`^.*[0-9].*`), "must contain at least one number"),
					validation.StringMatch(regexache.MustCompile(`^[^\@\/'" ]*$`), "cannot contain [/@\"' ]"),
				),
				ConflictsWith: []string{"manage_master_password"},
			},
			"master_password_secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_password_secret_kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidKMSKeyID,
			},
			"master_username": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexache.MustCompile(`^\w+$`), "must contain only alphanumeric characters"),
					validation.StringMatch(regexache.MustCompile(`(?i)^[a-z_]`), "first character must be a letter"),
				),
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"node_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"number_of_nodes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"owner_account": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrPort: {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5439,
				ValidateFunc: validation.IntBetween(1115, 65535),
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(val interface{}) string {
					if val == nil {
						return ""
					}
					return strings.ToLower(val.(string))
				},
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"snapshot_identifier"},
			},
			"snapshot_cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"snapshot_copy": {
				Type: schema.TypeList,
				Deprecated: "Use the aws_redshift_snapshot_copy resource instead. " +
					"This argument will be removed in a future major version.",
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"grant_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrRetentionPeriod: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  7,
						},
					},
				},
			},
			"snapshot_identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_arn"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				azRelocationEnabled, multiAZ := diff.Get("availability_zone_relocation_enabled").(bool), diff.Get("multi_az").(bool)

				if azRelocationEnabled && multiAZ {
					return errors.New("`availability_zone_relocation_enabled` and `multi_az` cannot be both true")
				}

				if diff.Id() != "" {
					if o, n := diff.GetChange(names.AttrAvailabilityZone); !azRelocationEnabled && o.(string) != n.(string) {
						return errors.New("cannot change `availability_zone` if `availability_zone_relocation_enabled` is not true")
					}
				}

				return nil
			},
		),
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	clusterID := d.Get(names.AttrClusterIdentifier).(string)
	inputR := &redshift.RestoreFromClusterSnapshotInput{
		AllowVersionUpgrade:              aws.Bool(d.Get("allow_version_upgrade").(bool)),
		AutomatedSnapshotRetentionPeriod: aws.Int64(int64(d.Get("automated_snapshot_retention_period").(int))),
		ClusterIdentifier:                aws.String(clusterID),
		Port:                             aws.Int64(int64(d.Get(names.AttrPort).(int))),
		NodeType:                         aws.String(d.Get("node_type").(string)),
		PubliclyAccessible:               aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
	}
	inputC := &redshift.CreateClusterInput{
		AllowVersionUpgrade:              aws.Bool(d.Get("allow_version_upgrade").(bool)),
		AutomatedSnapshotRetentionPeriod: aws.Int64(int64(d.Get("automated_snapshot_retention_period").(int))),
		ClusterIdentifier:                aws.String(clusterID),
		ClusterVersion:                   aws.String(d.Get("cluster_version").(string)),
		DBName:                           aws.String(d.Get(names.AttrDatabaseName).(string)),
		MasterUsername:                   aws.String(d.Get("master_username").(string)),
		NodeType:                         aws.String(d.Get("node_type").(string)),
		Port:                             aws.Int64(int64(d.Get(names.AttrPort).(int))),
		PubliclyAccessible:               aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
		Tags:                             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("aqua_configuration_status"); ok {
		inputR.AquaConfigurationStatus = aws.String(v.(string))
		inputC.AquaConfigurationStatus = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		inputR.AvailabilityZone = aws.String(v.(string))
		inputC.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_relocation_enabled"); ok {
		inputR.AvailabilityZoneRelocation = aws.Bool(v.(bool))
		inputC.AvailabilityZoneRelocation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("cluster_parameter_group_name"); ok {
		inputR.ClusterParameterGroupName = aws.String(v.(string))
		inputC.ClusterParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cluster_subnet_group_name"); ok {
		inputR.ClusterSubnetGroupName = aws.String(v.(string))
		inputC.ClusterSubnetGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_iam_role_arn"); ok {
		inputR.DefaultIamRoleArn = aws.String(v.(string))
		inputC.DefaultIamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("elastic_ip"); ok {
		inputR.ElasticIp = aws.String(v.(string))
		inputC.ElasticIp = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enhanced_vpc_routing"); ok {
		inputR.EnhancedVpcRouting = aws.Bool(v.(bool))
		inputC.EnhancedVpcRouting = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("iam_roles"); ok {
		inputR.IamRoles = flex.ExpandStringSet(v.(*schema.Set))
		inputC.IamRoles = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		inputR.KmsKeyId = aws.String(v.(string))
		inputC.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("maintenance_track_name"); ok {
		inputR.MaintenanceTrackName = aws.String(v.(string))
		inputC.MaintenanceTrackName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("manage_master_password"); ok {
		inputR.ManageMasterPassword = aws.Bool(v.(bool))
		inputC.ManageMasterPassword = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("manual_snapshot_retention_period"); ok {
		inputR.ManualSnapshotRetentionPeriod = aws.Int64(int64(v.(int)))
		inputC.ManualSnapshotRetentionPeriod = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("master_password"); ok {
		inputC.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_password_secret_kms_key_id"); ok {
		inputR.MasterPasswordSecretKmsKeyId = aws.String(v.(string))
		inputC.MasterPasswordSecretKmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("multi_az"); ok {
		inputR.MultiAZ = aws.Bool(v.(bool))
		inputC.MultiAZ = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("number_of_nodes"); ok {
		inputR.NumberOfNodes = aws.Int64(int64(v.(int)))
		// NumberOfNodes set below for CreateCluster.
	}

	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		inputR.PreferredMaintenanceWindow = aws.String(v.(string))
		inputC.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v := d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set); v.Len() > 0 {
		inputR.VpcSecurityGroupIds = flex.ExpandStringSet(v)
		inputC.VpcSecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := d.GetOk("snapshot_identifier"); ok {
		inputR.SnapshotIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("snapshot_arn"); ok {
		inputR.SnapshotArn = aws.String(v.(string))
	}

	if inputR.SnapshotArn != nil || inputR.SnapshotIdentifier != nil {
		if v, ok := d.GetOk("owner_account"); ok {
			inputR.OwnerAccount = aws.String(v.(string))
		}

		if v, ok := d.GetOk("snapshot_cluster_identifier"); ok {
			inputR.SnapshotClusterIdentifier = aws.String(v.(string))
		}

		output, err := conn.RestoreFromClusterSnapshotWithContext(ctx, inputR)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "restoring Redshift Cluster (%s) from snapshot: %s", clusterID, err)
		}

		d.SetId(aws.StringValue(output.Cluster.ClusterIdentifier))
	} else {
		if _, ok := d.GetOk("master_password"); !ok {
			if _, ok := d.GetOk("manage_master_password"); !ok {
				return sdkdiag.AppendErrorf(diags, `provider.aws: aws_redshift_cluster: %s: one of "manage_master_password" or "master_password" is required`, d.Get(names.AttrClusterIdentifier).(string))
			}
		}

		if _, ok := d.GetOk("master_username"); !ok {
			return sdkdiag.AppendErrorf(diags, `provider.aws: aws_redshift_cluster: %s: "master_username": required field is not set`, d.Get(names.AttrClusterIdentifier).(string))
		}

		if v, ok := d.GetOk(names.AttrEncrypted); ok {
			inputC.Encrypted = aws.Bool(v.(bool))
		}

		if v := d.Get("number_of_nodes").(int); v > 1 {
			inputC.ClusterType = aws.String(clusterTypeMultiNode)
			inputC.NumberOfNodes = aws.Int64(int64(d.Get("number_of_nodes").(int)))
		} else {
			inputC.ClusterType = aws.String(clusterTypeSingleNode)
		}

		output, err := conn.CreateClusterWithContext(ctx, inputC)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster (%s): %s", clusterID, err)
		}

		d.SetId(aws.StringValue(output.Cluster.ClusterIdentifier))
	}

	if _, err := waitClusterCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster (%s): waiting for completion: %s", d.Id(), err)
	}

	if _, err := waitClusterRelocationStatusResolved(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster (%s): waiting for relocation: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("snapshot_copy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if err := enableSnapshotCopy(ctx, conn, d.Id(), v.([]interface{})[0].(map[string]interface{})); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster (%s): %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		if v, ok := tfMap["enable"].(bool); ok && v {
			if err := enableLogging(ctx, conn, d.Id(), tfMap); err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Redshift Cluster (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	rsc, err := findClusterByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster (%s): %s", d.Id(), err)
	}

	loggingStatus, err := conn.DescribeLoggingStatusWithContext(ctx, &redshift.DescribeLoggingStatusInput{
		ClusterIdentifier: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Cluster (%s) logging status: %s", d.Id(), err)
	}

	d.Set("allow_version_upgrade", rsc.AllowVersionUpgrade)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   redshift.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("cluster:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if rsc.AquaConfiguration != nil {
		d.Set("aqua_configuration_status", rsc.AquaConfiguration.AquaConfigurationStatus)
	}
	d.Set("automated_snapshot_retention_period", rsc.AutomatedSnapshotRetentionPeriod)
	d.Set(names.AttrAvailabilityZone, rsc.AvailabilityZone)
	if v, err := clusterAvailabilityZoneRelocationStatus(rsc); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("availability_zone_relocation_enabled", v)
	}
	d.Set(names.AttrClusterIdentifier, rsc.ClusterIdentifier)
	d.Set("cluster_namespace_arn", rsc.ClusterNamespaceArn)
	if err := d.Set("cluster_nodes", flattenClusterNodes(rsc.ClusterNodes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_nodes: %s", err)
	}
	d.Set("cluster_parameter_group_name", rsc.ClusterParameterGroups[0].ParameterGroupName)
	d.Set("cluster_public_key", rsc.ClusterPublicKey)
	d.Set("cluster_revision_number", rsc.ClusterRevisionNumber)
	d.Set("cluster_subnet_group_name", rsc.ClusterSubnetGroupName)
	if len(rsc.ClusterNodes) > 1 {
		d.Set("cluster_type", clusterTypeMultiNode)
	} else {
		d.Set("cluster_type", clusterTypeSingleNode)
	}
	d.Set("cluster_version", rsc.ClusterVersion)
	d.Set(names.AttrDatabaseName, rsc.DBName)
	d.Set("default_iam_role_arn", rsc.DefaultIamRoleArn)
	d.Set(names.AttrEncrypted, rsc.Encrypted)
	d.Set("enhanced_vpc_routing", rsc.EnhancedVpcRouting)
	d.Set("iam_roles", tfslices.ApplyToAll(rsc.IamRoles, func(v *redshift.ClusterIamRole) string {
		return aws.StringValue(v.IamRoleArn)
	}))
	d.Set(names.AttrKMSKeyID, rsc.KmsKeyId)
	if err := d.Set("logging", flattenLogging(loggingStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging: %s", err)
	}
	d.Set("maintenance_track_name", rsc.MaintenanceTrackName)
	d.Set("manual_snapshot_retention_period", rsc.ManualSnapshotRetentionPeriod)
	d.Set("master_username", rsc.MasterUsername)
	d.Set("master_password_secret_arn", rsc.MasterPasswordSecretArn)
	d.Set("master_password_secret_kms_key_id", rsc.MasterPasswordSecretKmsKeyId)
	if v, err := clusterMultiAZStatus(rsc); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	} else {
		d.Set("multi_az", v)
	}
	d.Set("node_type", rsc.NodeType)
	d.Set("number_of_nodes", rsc.NumberOfNodes)
	d.Set(names.AttrPreferredMaintenanceWindow, rsc.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, rsc.PubliclyAccessible)
	if err := d.Set("snapshot_copy", flattenSnapshotCopy(rsc.ClusterSnapshotCopyStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_copy: %s", err)
	}
	d.Set(names.AttrVPCSecurityGroupIDs, tfslices.ApplyToAll(rsc.VpcSecurityGroups, func(v *redshift.VpcSecurityGroupMembership) string {
		return aws.StringValue(v.VpcSecurityGroupId)
	}))

	d.Set(names.AttrDNSName, nil)
	d.Set(names.AttrEndpoint, nil)
	d.Set(names.AttrPort, nil)
	if endpoint := rsc.Endpoint; endpoint != nil {
		if address := aws.StringValue(endpoint.Address); address != "" {
			d.Set(names.AttrDNSName, address)
			if port := aws.Int64Value(endpoint.Port); port != 0 {
				d.Set(names.AttrEndpoint, fmt.Sprintf("%s:%d", address, port))
				d.Set(names.AttrPort, port)
			} else {
				d.Set(names.AttrEndpoint, address)
			}
		}
	}

	setTagsOut(ctx, rsc.Tags)

	return diags
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChangesExcept("aqua_configuration_status", names.AttrAvailabilityZone, "iam_roles", "logging", "multi_az", "snapshot_copy", names.AttrTags, names.AttrTagsAll, "skip_final_snapshot") {
		input := &redshift.ModifyClusterInput{
			ClusterIdentifier: aws.String(d.Id()),
		}

		if d.HasChange("allow_version_upgrade") {
			input.AllowVersionUpgrade = aws.Bool(d.Get("allow_version_upgrade").(bool))
		}

		if d.HasChange("automated_snapshot_retention_period") {
			input.AutomatedSnapshotRetentionPeriod = aws.Int64(int64(d.Get("automated_snapshot_retention_period").(int)))
		}

		if d.HasChange("availability_zone_relocation_enabled") {
			input.AvailabilityZoneRelocation = aws.Bool(d.Get("availability_zone_relocation_enabled").(bool))
		}

		if d.HasChange("cluster_parameter_group_name") {
			input.ClusterParameterGroupName = aws.String(d.Get("cluster_parameter_group_name").(string))
		}

		if d.HasChange("maintenance_track_name") {
			input.MaintenanceTrackName = aws.String(d.Get("maintenance_track_name").(string))
		}

		if d.HasChange("manual_snapshot_retention_period") {
			input.ManualSnapshotRetentionPeriod = aws.Int64(int64(d.Get("manual_snapshot_retention_period").(int)))
		}

		// If the cluster type, node type, or number of nodes changed, then the AWS API expects all three
		// items to be sent over.
		if d.HasChanges("cluster_type", "node_type", "number_of_nodes") {
			input.NodeType = aws.String(d.Get("node_type").(string))

			if v := d.Get("number_of_nodes").(int); v > 1 {
				input.ClusterType = aws.String(clusterTypeMultiNode)
				input.NumberOfNodes = aws.Int64(int64(d.Get("number_of_nodes").(int)))
			} else {
				input.ClusterType = aws.String(clusterTypeSingleNode)
			}
		}

		if d.HasChange("cluster_version") {
			input.ClusterVersion = aws.String(d.Get("cluster_version").(string))
		}

		if d.HasChange(names.AttrEncrypted) {
			input.Encrypted = aws.Bool(d.Get(names.AttrEncrypted).(bool))
		}

		if d.HasChange("enhanced_vpc_routing") {
			input.EnhancedVpcRouting = aws.Bool(d.Get("enhanced_vpc_routing").(bool))
		}

		if d.Get(names.AttrEncrypted).(bool) && d.HasChange(names.AttrKMSKeyID) {
			input.KmsKeyId = aws.String(d.Get(names.AttrKMSKeyID).(string))
		}

		if d.HasChange("master_password") {
			input.MasterUserPassword = aws.String(d.Get("master_password").(string))
		}

		if d.HasChange("master_password_secret_kms_key_id") {
			input.MasterPasswordSecretKmsKeyId = aws.String(d.Get("master_password_secret_kms_key_id").(string))
		}

		if d.HasChange("manage_master_password") {
			input.ManageMasterPassword = aws.Bool(d.Get("manage_master_password").(bool))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange(names.AttrPubliclyAccessible) {
			input.PubliclyAccessible = aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set))
		}

		_, err := conn.ModifyClusterWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) update: %s", d.Id(), err)
		}

		if _, err := waitClusterRelocationStatusResolved(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) Availability Zone Relocation Status resolution: %s", d.Id(), err)
		}
	}

	if d.HasChanges("default_iam_role_arn", "iam_roles") {
		o, n := d.GetChange("iam_roles")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)

		input := &redshift.ModifyClusterIamRolesInput{
			AddIamRoles:       flex.ExpandStringSet(add),
			ClusterIdentifier: aws.String(d.Id()),
			RemoveIamRoles:    flex.ExpandStringSet(del),
			DefaultIamRoleArn: aws.String(d.Get("default_iam_role_arn").(string)),
		}

		_, err := conn.ModifyClusterIamRolesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Cluster (%s) IAM roles: %s", d.Id(), err)
		}

		if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("aqua_configuration_status") {
		input := &redshift.ModifyAquaConfigurationInput{
			AquaConfigurationStatus: aws.String(d.Get("aqua_configuration_status").(string)),
			ClusterIdentifier:       aws.String(d.Id()),
		}

		_, err := conn.ModifyAquaConfigurationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Cluster (%s) Aqua Configuration: %s", d.Id(), err)
		}

		if d.Get(names.AttrApplyImmediately).(bool) {
			input := &redshift.RebootClusterInput{
				ClusterIdentifier: aws.String(d.Id()),
			}

			_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, clusterInvalidClusterStateFaultTimeout,
				func() (interface{}, error) {
					return conn.RebootClusterWithContext(ctx, input)
				},
				redshift.ErrCodeInvalidClusterStateFault,
			)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "rebooting Redshift Cluster (%s): %s", d.Id(), err)
			}

			if _, err := waitClusterRebooted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) reboot: %s", d.Id(), err)
			}

			if _, err := waitClusterAquaApplied(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) Aqua Configuration update: %s", d.Id(), err)
			}
		}
	}

	// Availability Zone cannot be changed at the same time as other settings
	if d.HasChange(names.AttrAvailabilityZone) {
		input := &redshift.ModifyClusterInput{
			AvailabilityZone:  aws.String(d.Get(names.AttrAvailabilityZone).(string)),
			ClusterIdentifier: aws.String(d.Id()),
		}

		_, err := conn.ModifyClusterWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "relocating Redshift Cluster (%s): %s", d.Id(), err)
		}

		if _, err := waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("snapshot_copy") {
		if v, ok := d.GetOk("snapshot_copy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			if err := enableSnapshotCopy(ctx, conn, d.Id(), v.([]interface{})[0].(map[string]interface{})); err != nil {
				if !tfawserr.ErrCodeEquals(err, redshift.ErrCodeSnapshotCopyAlreadyEnabledFault) {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster (%s) snapshot_copy: %s", d.Id(), err)
				}
				if err := toggleSnapshotCopy(ctx, conn, d.Id(), v.([]interface{})[0].(map[string]interface{})); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster (%s) snapshot_copy: %s", d.Id(), err)
				}
			}
		} else {
			if err := disableSnapshotCopy(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster (%s) snapshot_copy: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("logging") {
		if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if v, ok := tfMap["enable"].(bool); ok && v {
				if err := enableLogging(ctx, conn, d.Id(), tfMap); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster (%s): %s", d.Id(), err)
				}
			} else {
				if err := disableLogging(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Redshift Cluster (%s): %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("multi_az") {
		azRelocationEnabled, multiAZ := d.Get("availability_zone_relocation_enabled").(bool), d.Get("multi_az").(bool)
		input := &redshift.ModifyClusterInput{
			ClusterIdentifier: aws.String(d.Id()),
			MultiAZ:           aws.Bool(multiAZ),
		}

		_, err := conn.ModifyClusterWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Cluster (%s) multi-AZ: %s", d.Id(), err)
		}

		if _, err = waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) update: %s", d.Id(), err)
		}

		if !multiAZ {
			// Disabling MultiAZ, Redshift automatically enables AZ Relocation.
			// For that reason is necessary to align it with the current configuration.
			input = &redshift.ModifyClusterInput{
				AvailabilityZoneRelocation: aws.Bool(azRelocationEnabled),
				ClusterIdentifier:          aws.String(d.Id()),
			}

			_, err = conn.ModifyClusterWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying Redshift Cluster (%s) AZ relocation: %s", d.Id(), err)
			}

			if _, err = waitClusterUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) update: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterRead(ctx, d, meta)...)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)
	input := &redshift.DeleteClusterInput{
		ClusterIdentifier:        aws.String(d.Id()),
		SkipFinalClusterSnapshot: aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if v, ok := d.GetOk(names.AttrFinalSnapshotIdentifier); ok {
			input.FinalClusterSnapshotIdentifier = aws.String(v.(string))
		} else {
			return sdkdiag.AppendErrorf(diags, "Redshift Cluster Instance FinalSnapshotIdentifier is required when a final snapshot is required")
		}
	}

	log.Printf("[DEBUG] Deleting Redshift Cluster: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, clusterInvalidClusterStateFaultTimeout,
		func() (interface{}, error) {
			return conn.DeleteClusterWithContext(ctx, input)
		},
		redshift.ErrCodeInvalidClusterStateFault,
	)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Cluster (%s): %s", d.Id(), err)
	}

	if _, err := waitClusterDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Cluster (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required.
	d.Set("skip_final_snapshot", true)

	return []*schema.ResourceData{d}, nil
}

func enableLogging(ctx context.Context, conn *redshift.Redshift, clusterID string, tfMap map[string]interface{}) error {
	input := &redshift.EnableLoggingInput{
		ClusterIdentifier: aws.String(clusterID),
	}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		input.BucketName = aws.String(v)
	}

	if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
		input.LogDestinationType = aws.String(v)
	}

	if v, ok := tfMap["log_exports"].(*schema.Set); ok && v.Len() > 0 {
		input.LogExports = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap[names.AttrS3KeyPrefix].(string); ok && v != "" {
		input.S3KeyPrefix = aws.String(v)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, clusterInvalidClusterStateFaultTimeout,
		func() (interface{}, error) {
			return conn.EnableLoggingWithContext(ctx, input)
		},
		redshift.ErrCodeInvalidClusterStateFault,
	)

	if err != nil {
		return fmt.Errorf("enabling logging: %w", err)
	}

	return nil
}

func disableLogging(ctx context.Context, conn *redshift.Redshift, clusterID string) error {
	input := &redshift.DisableLoggingInput{
		ClusterIdentifier: aws.String(clusterID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, clusterInvalidClusterStateFaultTimeout,
		func() (interface{}, error) {
			return conn.DisableLoggingWithContext(ctx, input)
		},
		redshift.ErrCodeInvalidClusterStateFault,
	)

	if err != nil {
		return fmt.Errorf("disabling logging: %w", err)
	}

	return nil
}

func enableSnapshotCopy(ctx context.Context, conn *redshift.Redshift, clusterID string, tfMap map[string]interface{}) error {
	input := &redshift.EnableSnapshotCopyInput{
		ClusterIdentifier: aws.String(clusterID),
		DestinationRegion: aws.String(tfMap["destination_region"].(string)),
	}

	if v, ok := tfMap[names.AttrRetentionPeriod]; ok {
		input.RetentionPeriod = aws.Int64(int64(v.(int)))
	}

	if v, ok := tfMap["grant_name"]; ok {
		input.SnapshotCopyGrantName = aws.String(v.(string))
	}

	_, err := conn.EnableSnapshotCopyWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("enabling snapshot copy: %w", err)
	}

	return nil
}

func disableSnapshotCopy(ctx context.Context, conn *redshift.Redshift, clusterID string) error {
	input := &redshift.DisableSnapshotCopyInput{
		ClusterIdentifier: aws.String(clusterID),
	}

	_, err := conn.DisableSnapshotCopyWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("disabling snapshot copy: %w", err)
	}

	return nil
}

// toggleSnapshotCopy calls disableSnapshotCopy followed by enableSnapshotCopy
//
// This workflow is necessary in cases where the existing snapshot copy configuration
// needs to be updated. Once enabled, `EnableSnapshotCopy` cannot be called to update existing
// settings. While the `ModifySnapshotCopyRetentionPeriod` API is available to update the
// `retention_period` argument, there is no mechanism to update other arguments such
// as `destination_region` or `snapshot_copy_grant_name` without disabling first.
func toggleSnapshotCopy(ctx context.Context, conn *redshift.Redshift, clusterID string, tfMap map[string]interface{}) error {
	if err := disableSnapshotCopy(ctx, conn, clusterID); err != nil {
		return err
	}
	if err := enableSnapshotCopy(ctx, conn, clusterID, tfMap); err != nil {
		return err
	}
	return nil
}

func flattenClusterNode(apiObject *redshift.ClusterNode) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.NodeRole; v != nil {
		tfMap["node_role"] = aws.StringValue(v)
	}

	if v := apiObject.PrivateIPAddress; v != nil {
		tfMap["private_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.PublicIPAddress; v != nil {
		tfMap["public_ip_address"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenClusterNodes(apiObjects []*redshift.ClusterNode) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenClusterNode(apiObject))
	}

	return tfList
}

func clusterAvailabilityZoneRelocationStatus(cluster *redshift.Cluster) (bool, error) {
	// AvailabilityZoneRelocation is not returned by the API, and AvailabilityZoneRelocationStatus is not implemented as Const at this time.
	switch availabilityZoneRelocationStatus := aws.StringValue(cluster.AvailabilityZoneRelocationStatus); availabilityZoneRelocationStatus {
	case "enabled":
		return true, nil
	case "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected AvailabilityZoneRelocationStatus value %q returned by API", availabilityZoneRelocationStatus)
	}
}

func clusterMultiAZStatus(cluster *redshift.Cluster) (bool, error) {
	// MultiAZ is returned as string from the API but is implemented as bool to keep consistency with other parameters.
	switch multiAZStatus := aws.StringValue(cluster.MultiAZ); strings.ToLower(multiAZStatus) {
	case "enabled":
		return true, nil
	case "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected MultiAZ value %q returned by API", multiAZStatus)
	}
}
