// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	dms "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
func resourceReplicationInstance() *schema.Resource {
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
			names.AttrAllocatedStorage: {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(5, 6144),
			},
			names.AttrAllowMajorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			names.AttrKMSKeyARN: {
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
			names.AttrPreferredMaintenanceWindow: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			names.AttrPubliclyAccessible: {
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
			names.AttrVPCSecurityGroupIDs: {
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
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	replicationInstanceID := d.Get("replication_instance_id").(string)
	input := &dms.CreateReplicationInstanceInput{
		AutoMinorVersionUpgrade:       aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool)),
		PubliclyAccessible:            aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool)),
		MultiAZ:                       aws.Bool(d.Get("multi_az").(bool)),
		ReplicationInstanceClass:      aws.String(d.Get("replication_instance_class").(string)),
		ReplicationInstanceIdentifier: aws.String(replicationInstanceID),
		Tags:                          getTagsIn(ctx),
	}

	// WARNING: GetOk returns the zero value for the type if the key is omitted in config. This means for optional
	// keys that the zero value is valid we cannot know if the zero value was in the config and cannot allow the API
	// to set the default value. See GitHub Issue #5694 https://github.com/hashicorp/terraform/issues/5694

	if v, ok := d.GetOk(names.AttrAllocatedStorage); ok {
		input.AllocatedStorage = aws.Int32(int32(v.(int)))
	}
	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("network_type"); ok {
		input.NetworkType = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}
	if v, ok := d.GetOk("replication_subnet_group_id"); ok {
		input.ReplicationSubnetGroupIdentifier = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok {
		input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.CreateReplicationInstance(ctx, input)

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
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	instance, err := findReplicationInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Instance (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAllocatedStorage, instance.AllocatedStorage)
	d.Set(names.AttrAutoMinorVersionUpgrade, instance.AutoMinorVersionUpgrade)
	d.Set(names.AttrAvailabilityZone, instance.AvailabilityZone)
	d.Set(names.AttrEngineVersion, instance.EngineVersion)
	d.Set(names.AttrKMSKeyARN, instance.KmsKeyId)
	d.Set("multi_az", instance.MultiAZ)
	d.Set("network_type", instance.NetworkType)
	d.Set(names.AttrPreferredMaintenanceWindow, instance.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, instance.PubliclyAccessible)
	d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
	d.Set("replication_instance_class", instance.ReplicationInstanceClass)
	d.Set("replication_instance_id", instance.ReplicationInstanceIdentifier)
	d.Set("replication_instance_private_ips", instance.ReplicationInstancePrivateIpAddresses)
	d.Set("replication_instance_public_ips", instance.ReplicationInstancePublicIpAddresses)
	d.Set("replication_subnet_group_id", instance.ReplicationSubnetGroup.ReplicationSubnetGroupIdentifier)
	vpcSecurityGroupIDs := tfslices.ApplyToAll(instance.VpcSecurityGroups, func(v awstypes.VpcSecurityGroupMembership) string {
		return aws.ToString(v.VpcSecurityGroupId)
	})
	d.Set(names.AttrVPCSecurityGroupIDs, vpcSecurityGroupIDs)

	return diags
}

func resourceReplicationInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, names.AttrAllowMajorVersionUpgrade) {
		// Having allowing_major_version_upgrade by itself should not trigger ModifyReplicationInstance
		// as it results in InvalidParameterCombination: No modifications were requested
		input := &dms.ModifyReplicationInstanceInput{
			AllowMajorVersionUpgrade: d.Get(names.AttrAllowMajorVersionUpgrade).(bool),
			ApplyImmediately:         d.Get(names.AttrApplyImmediately).(bool),
			ReplicationInstanceArn:   aws.String(d.Get("replication_instance_arn").(string)),
		}

		if d.HasChange(names.AttrAllocatedStorage) {
			input.AllocatedStorage = aws.Int32(int32(d.Get(names.AttrAllocatedStorage).(int)))
		}

		if d.HasChange(names.AttrAutoMinorVersionUpgrade) {
			input.AutoMinorVersionUpgrade = aws.Bool(d.Get(names.AttrAutoMinorVersionUpgrade).(bool))
		}

		if d.HasChange(names.AttrEngineVersion) {
			input.EngineVersion = aws.String(d.Get(names.AttrEngineVersion).(string))
		}

		if d.HasChange("multi_az") {
			input.MultiAZ = aws.Bool(d.Get("multi_az").(bool))
		}

		if d.HasChange("network_type") {
			input.NetworkType = aws.String(d.Get("network_type").(string))
		}

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange("replication_instance_class") {
			input.ReplicationInstanceClass = aws.String(d.Get("replication_instance_class").(string))
		}

		if d.HasChange(names.AttrVPCSecurityGroupIDs) {
			input.VpcSecurityGroupIds = flex.ExpandStringValueSet(d.Get(names.AttrVPCSecurityGroupIDs).(*schema.Set))
		}

		_, err := conn.ModifyReplicationInstance(ctx, input)

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
	conn := meta.(*conns.AWSClient).DMSClient(ctx)

	log.Printf("[DEBUG] Deleting DMS Replication Instance: %s", d.Id())
	_, err := conn.DeleteReplicationInstance(ctx, &dms.DeleteReplicationInstanceInput{
		ReplicationInstanceArn: aws.String(d.Get("replication_instance_arn").(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
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

func findReplicationInstanceByID(ctx context.Context, conn *dms.Client, id string) (*awstypes.ReplicationInstance, error) {
	input := &dms.DescribeReplicationInstancesInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("replication-instance-id"),
				Values: []string{id},
			},
		},
	}

	return findReplicationInstance(ctx, conn, input)
}

func findReplicationInstance(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationInstancesInput) (*awstypes.ReplicationInstance, error) {
	output, err := findReplicationInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findReplicationInstances(ctx context.Context, conn *dms.Client, input *dms.DescribeReplicationInstancesInput) ([]awstypes.ReplicationInstance, error) {
	var output []awstypes.ReplicationInstance

	pages := dms.NewDescribeReplicationInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ReplicationInstances...)
	}

	return output, nil
}

func statusReplicationInstance(ctx context.Context, conn *dms.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findReplicationInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.ReplicationInstanceStatus), nil
	}
}

func waitReplicationInstanceCreated(ctx context.Context, conn *dms.Client, id string, timeout time.Duration) (*awstypes.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusCreating, replicationInstanceStatusModifying},
		Target:     []string{replicationInstanceStatusAvailable},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationInstanceUpdated(ctx context.Context, conn *dms.Client, id string, timeout time.Duration) (*awstypes.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusModifying, replicationInstanceStatusUpgrading},
		Target:     []string{replicationInstanceStatusAvailable},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationInstanceDeleted(ctx context.Context, conn *dms.Client, id string, timeout time.Duration) (*awstypes.ReplicationInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationInstanceStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationInstance(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ReplicationInstance); ok {
		return output, err
	}

	return nil, err
}
