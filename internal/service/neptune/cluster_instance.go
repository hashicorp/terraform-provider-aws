// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_cluster_instance", name="Cluster")
// @Tags(identifierAttribute="arn")
func ResourceClusterInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterInstanceCreate,
		ReadWithoutTimeout:   resourceClusterInstanceRead,
		UpdateWithoutTimeout: resourceClusterInstanceUpdate,
		DeleteWithoutTimeout: resourceClusterInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dbi_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngine: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      engineNeptune,
				ValidateFunc: validation.StringInSlice(engine_Values(), false),
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIdentifier: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
				ValidateFunc:  validIdentifier,
			},
			"identifier_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrIdentifier},
				ValidateFunc:  validIdentifierPrefix,
			},
			"instance_class": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"neptune_parameter_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default.neptune1",
			},
			"neptune_subnet_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  DefaultPort,
			},
			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
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
			"promotion_tier": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrStorageEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"writer": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	instanceID := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrIdentifier).(string)),
		create.WithConfiguredPrefix(d.Get("identifier_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	input := &neptune.CreateDBInstanceInput{
		AutoMinorVersionUpgrade: aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
		DBClusterIdentifier:     aws.String(d.Get(names.AttrClusterIdentifier).(string)),
		DBInstanceClass:         aws.String(d.Get("instance_class").(string)),
		DBInstanceIdentifier:    aws.String(instanceID),
		Engine:                  aws.String(d.Get(names.AttrEngine).(string)),
		PromotionTier:           aws.Int64(int64(d.Get("promotion_tier").(int))),
		PubliclyAccessible:      aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("neptune_parameter_group_name"); ok {
		input.DBParameterGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("neptune_subnet_group_name"); ok {
		input.DBSubnetGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		input.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateDBInstanceWithContext(ctx, input)
	}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster Instance (%s): %s", instanceID, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*neptune.CreateDBInstanceOutput).DBInstance.DBInstanceIdentifier))

	if _, err := waitDBInstanceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	db, err := FindDBInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Cluster Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Instance (%s): %s", d.Id(), err)
	}

	clusterID := aws.StringValue(db.DBClusterIdentifier)
	d.Set(names.AttrARN, db.DBInstanceArn)
	d.Set(names.AttrAutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, db.AvailabilityZone)
	d.Set(names.AttrClusterIdentifier, clusterID)
	d.Set("dbi_resource_id", db.DbiResourceId)
	d.Set(names.AttrEngineVersion, db.EngineVersion)
	d.Set(names.AttrEngine, db.Engine)
	d.Set(names.AttrIdentifier, db.DBInstanceIdentifier)
	d.Set("identifier_prefix", create.NamePrefixFromName(aws.StringValue(db.DBInstanceIdentifier)))
	d.Set("instance_class", db.DBInstanceClass)
	d.Set(names.AttrKMSKeyARN, db.KmsKeyId)
	if len(db.DBParameterGroups) > 0 {
		d.Set("neptune_parameter_group_name", db.DBParameterGroups[0].DBParameterGroupName)
	}
	if db.DBSubnetGroup != nil {
		d.Set("neptune_subnet_group_name", db.DBSubnetGroup.DBSubnetGroupName)
	}
	d.Set("preferred_backup_window", db.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	d.Set("promotion_tier", db.PromotionTier)
	d.Set(names.AttrPubliclyAccessible, db.PubliclyAccessible)
	d.Set(names.AttrStorageEncrypted, db.StorageEncrypted)
	d.Set(names.AttrStorageType, db.StorageType)

	if db.Endpoint != nil {
		address := aws.StringValue(db.Endpoint.Address)
		port := int(aws.Int64Value(db.Endpoint.Port))

		d.Set(names.AttrAddress, address)
		d.Set(names.AttrEndpoint, fmt.Sprintf("%s:%d", address, port))
		d.Set(names.AttrPort, port)
	}

	m, err := findClusterMemberByInstanceByTwoPartKey(ctx, conn, clusterID, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster (%s) member (%s): %s", clusterID, d.Id(), err)
	}

	d.Set("writer", m.IsClusterWriter)

	return diags
}

func resourceClusterInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &neptune.ModifyDBInstanceInput{
			ApplyImmediately:     aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
			DBInstanceIdentifier: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAutoMinorVersionUpgrade) {
			input.AutoMinorVersionUpgrade = aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool))
		}

		if d.HasChange("instance_class") {
			input.DBInstanceClass = aws.String(d.Get("instance_class").(string))
		}

		if d.HasChange("neptune_parameter_group_name") {
			input.DBParameterGroupName = aws.String(d.Get("neptune_parameter_group_name").(string))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange("promotion_tier") {
			input.PromotionTier = aws.Int64(int64(d.Get("promotion_tier").(int)))
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.ModifyDBInstanceWithContext(ctx, input)
		}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitDBInstanceAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterInstanceRead(ctx, d, meta)...)
}

func resourceClusterInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	log.Printf("[DEBUG] Deleting Neptune Cluster Instance: %s", d.Id())
	input := neptune.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(d.Id()),
	}

	if d.Get("skip_final_snapshot").(bool) {
		input.SkipFinalSnapshot = aws.Bool(true)
	}

	_, err := conn.DeleteDBInstanceWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBInstanceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitDBInstanceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindDBInstanceByID(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBInstance, error) {
	input := &neptune.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(id),
	}
	output, err := findDBInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBInstanceIdentifier) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBInstance(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBInstancesInput) (*neptune.DBInstance, error) {
	output, err := findDBInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBInstances(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBInstancesInput) ([]*neptune.DBInstance, error) {
	var output []*neptune.DBInstance

	err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *neptune.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBInstanceNotFoundFault) {
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

func findClusterMemberByInstanceByTwoPartKey(ctx context.Context, conn *neptune.Neptune, clusterID, instanceID string) (*neptune.DBClusterMember, error) {
	output, err := FindDBClusterByID(ctx, conn, clusterID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(tfslices.Filter(output.DBClusterMembers, func(v *neptune.DBClusterMember) bool {
		return aws.StringValue(v.DBInstanceIdentifier) == instanceID
	}))
}

func statusDBInstance(ctx context.Context, conn *neptune.Neptune, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDBInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DBInstanceStatus), nil
	}
}

func waitDBInstanceAvailable(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBInstance, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			dbInstanceStatusBackingUp,
			dbInstanceStatusConfiguringEnhancedMonitoring,
			dbInstanceStatusConfiguringIAMDatabaseAuth,
			dbInstanceStatusConfiguringLogExports,
			dbInstanceStatusCreating,
			dbInstanceStatusMaintenance,
			dbInstanceStatusModifying,
			dbInstanceStatusRebooting,
			dbInstanceStatusRenaming,
			dbInstanceStatusResettingMasterCredentials,
			dbInstanceStatusStarting,
			dbInstanceStatusStorageOptimization,
			dbInstanceStatusUpgrading,
		},
		Target:     []string{dbInstanceStatusAvailable},
		Refresh:    statusDBInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceDeleted(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			dbInstanceStatusModifying,
			dbInstanceStatusDeleting,
		},
		Target:     []string{},
		Refresh:    statusDBInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptune.DBInstance); ok {
		return output, err
	}

	return nil, err
}
