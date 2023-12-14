// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_dms_replication_instance", name="Replication Instance")
// @Tags(identifierAttribute="replication_instance_arn")
func ResourceReplicationInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationInstanceCreate,
		ReadWithoutTimeout:   resourceReplicationInstanceRead,
		UpdateWithoutTimeout: resourceReplicationInstanceUpdate,
		DeleteWithoutTimeout: resourceReplicationInstanceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"allocated_storage": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(5, 6144),
			},
			"allow_major_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"multi_az": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"network_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(networkType_Values(), false),
			},
			"preferred_maintenance_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"replication_instance_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_instance_class": {
				Type:     schema.TypeString,
				Required: true,
				// Valid Values: dms.t2.micro | dms.t2.small | dms.t2.medium | dms.t2.large | dms.c4.large |
				// dms.c4.xlarge | dms.c4.2xlarge | dms.c4.4xlarge
			},
			"replication_instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validReplicationInstanceID,
			},
			"replication_instance_private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replication_instance_public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"replication_subnet_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	replicationInstanceID := d.Get("replication_instance_id").(string)
	input := &dms.CreateReplicationInstanceInput{
		AutoMinorVersionUpgrade:       aws.Bool(d.Get("auto_minor_version_upgrade").(bool)),
		PubliclyAccessible:            aws.Bool(d.Get("publicly_accessible").(bool)),
		MultiAZ:                       aws.Bool(d.Get("multi_az").(bool)),
		ReplicationInstanceClass:      aws.String(d.Get("replication_instance_class").(string)),
		ReplicationInstanceIdentifier: aws.String(replicationInstanceID),
		Tags:                          getTagsIn(ctx),
	}

	// WARNING: GetOk returns the zero value for the type if the key is omitted in config. This means for optional
	// keys that the zero value is valid we cannot know if the zero value was in the config and cannot allow the API
	// to set the default value. See GitHub Issue #5694 https://github.com/hashicorp/terraform/issues/5694

	if v, ok := d.GetOk("allocated_storage"); ok {
		input.AllocatedStorage = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}
	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("kms_key_arn"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("network_type"); ok {
		input.NetworkType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}
	if v, ok := d.GetOk("replication_subnet_group_id"); ok {
		input.ReplicationSubnetGroupIdentifier = aws.String(v.(string))
	}
	if v, ok := d.GetOk("vpc_security_group_ids"); ok {
		input.VpcSecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateReplicationInstanceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Instance (%s): %s", replicationInstanceID, err)
	}

	d.SetId(replicationInstanceID)

	if _, err := waitReplicationInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceReplicationInstanceRead(ctx, d, meta)...)
}

func resourceReplicationInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	instance, err := FindReplicationInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Instance (%s): %s", d.Id(), err)
	}

	d.Set("allocated_storage", instance.AllocatedStorage)
	d.Set("auto_minor_version_upgrade", instance.AutoMinorVersionUpgrade)
	d.Set("availability_zone", instance.AvailabilityZone)
	d.Set("engine_version", instance.EngineVersion)
	d.Set("kms_key_arn", instance.KmsKeyId)
	d.Set("multi_az", instance.MultiAZ)
	d.Set("network_type", instance.NetworkType)
	d.Set("preferred_maintenance_window", instance.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", instance.PubliclyAccessible)
	d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
	d.Set("replication_instance_class", instance.ReplicationInstanceClass)
	d.Set("replication_instance_id", instance.ReplicationInstanceIdentifier)
	d.Set("replication_instance_private_ips", aws.StringValueSlice(instance.ReplicationInstancePrivateIpAddresses))
	d.Set("replication_instance_public_ips", aws.StringValueSlice(instance.ReplicationInstancePublicIpAddresses))
	d.Set("replication_subnet_group_id", instance.ReplicationSubnetGroup.ReplicationSubnetGroupIdentifier)
	vpcSecurityGroupIDs := tfslices.ApplyToAll(instance.VpcSecurityGroups, func(sg *dms.VpcSecurityGroupMembership) string {
		return aws.StringValue(sg.VpcSecurityGroupId)
	})
	d.Set("vpc_security_group_ids", vpcSecurityGroupIDs)

	return diags
}

func resourceReplicationInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "allow_major_version_upgrade") {
		// Having allowing_major_version_upgrade by itself should not trigger ModifyReplicationInstance
		// as it results in InvalidParameterCombination: No modifications were requested
		input := &dms.ModifyReplicationInstanceInput{
			AllowMajorVersionUpgrade: aws.Bool(d.Get("allow_major_version_upgrade").(bool)),
			ApplyImmediately:         aws.Bool(d.Get("apply_immediately").(bool)),
			ReplicationInstanceArn:   aws.String(d.Get("replication_instance_arn").(string)),
		}

		if d.HasChange("allocated_storage") {
			input.AllocatedStorage = aws.Int64(int64(d.Get("allocated_storage").(int)))
		}

		if d.HasChange("auto_minor_version_upgrade") {
			input.AutoMinorVersionUpgrade = aws.Bool(d.Get("auto_minor_version_upgrade").(bool))
		}

		if d.HasChange("engine_version") {
			input.EngineVersion = aws.String(d.Get("engine_version").(string))
		}

		if d.HasChange("multi_az") {
			input.MultiAZ = aws.Bool(d.Get("multi_az").(bool))
		}

		if d.HasChange("network_type") {
			input.NetworkType = aws.String(d.Get("network_type").(string))
		}

		if d.HasChange("preferred_maintenance_window") {
			input.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		}

		if d.HasChange("replication_instance_class") {
			input.ReplicationInstanceClass = aws.String(d.Get("replication_instance_class").(string))
		}

		if d.HasChange("vpc_security_group_ids") {
			input.VpcSecurityGroupIds = flex.ExpandStringSet(d.Get("vpc_security_group_ids").(*schema.Set))
		}

		_, err := conn.ModifyReplicationInstanceWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Replication Instance (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicationInstanceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Instance (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReplicationInstanceRead(ctx, d, meta)...)
}

func resourceReplicationInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	log.Printf("[DEBUG] Deleting DMS Replication Instance: %s", d.Id())
	_, err := conn.DeleteReplicationInstanceWithContext(ctx, &dms.DeleteReplicationInstanceInput{
		ReplicationInstanceArn: aws.String(d.Get("replication_instance_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicationInstanceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindReplicationInstanceByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationInstance, error) {
	input := &dms.DescribeReplicationInstancesInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-instance-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	return findReplicationInstance(ctx, conn, input)
}

func findReplicationInstance(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationInstancesInput) (*dms.ReplicationInstance, error) {
	output, err := findReplicationInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findReplicationInstances(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationInstancesInput) ([]*dms.ReplicationInstance, error) {
	var output []*dms.ReplicationInstance

	err := conn.DescribeReplicationInstancesPagesWithContext(ctx, input, func(page *dms.DescribeReplicationInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationInstances {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
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

func statusReplicationInstance(ctx context.Context, conn *dms.DatabaseMigrationService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReplicationInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ReplicationInstanceStatus), nil
	}
}

func waitReplicationInstanceCreated(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusCreating, replicationInstanceStatusModifying},
		Target:     []string{replicationInstanceStatusAvailable},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationInstanceUpdated(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusModifying, replicationInstanceStatusUpgrading},
		Target:     []string{replicationInstanceStatusAvailable},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationInstanceDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}
