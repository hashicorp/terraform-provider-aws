// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_efs_mount_target")
func ResourceMountTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMountTargetCreate,
		ReadWithoutTimeout:   resourceMountTargetRead,
		UpdateWithoutTimeout: resourceMountTargetUpdate,
		DeleteWithoutTimeout: resourceMountTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"mount_target_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceMountTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	// CreateMountTarget would return the same Mount Target ID
	// to parallel requests if they both include the same AZ
	// and we would end up managing the same MT as 2 resources.
	// So we make it fail by calling 1 request per AZ at a time.
	subnetID := d.Get("subnet_id").(string)
	az, err := getAZFromSubnetID(ctx, meta.(*conns.AWSClient).EC2Conn(ctx), subnetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnet (%s): %s", subnetID, err)
	}

	fsID := d.Get("file_system_id").(string)
	mtKey := "efs-mt-" + fsID + "-" + az
	conns.GlobalMutexKV.Lock(mtKey)
	defer conns.GlobalMutexKV.Unlock(mtKey)

	input := &efs.CreateMountTargetInput{
		FileSystemId: aws.String(fsID),
		SubnetId:     aws.String(subnetID),
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_groups"); ok {
		input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	mt, err := conn.CreateMountTargetWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Mount Target (%s): %s", fsID, err)
	}

	d.SetId(aws.StringValue(mt.MountTargetId))

	if _, err := waitMountTargetCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Mount Target (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	mt, err := FindMountTargetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Mount Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(mt.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set("dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(mt.FileSystemId))))
	d.Set("file_system_arn", arn)
	d.Set("file_system_id", mt.FileSystemId)
	d.Set("ip_address", mt.IpAddress)
	d.Set("mount_target_dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.%s.efs", aws.StringValue(mt.AvailabilityZoneName), aws.StringValue(mt.FileSystemId))))
	d.Set("network_interface_id", mt.NetworkInterfaceId)
	d.Set("owner_id", mt.OwnerId)
	d.Set("subnet_id", mt.SubnetId)

	output, err := conn.DescribeMountTargetSecurityGroupsWithContext(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s) security groups: %s", d.Id(), err)
	}

	d.Set("security_groups", aws.StringValueSlice(output.SecurityGroups))

	return diags
}

func resourceMountTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	if d.HasChange("security_groups") {
		input := &efs.ModifyMountTargetSecurityGroupsInput{
			MountTargetId:  aws.String(d.Id()),
			SecurityGroups: flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		_, err := conn.ModifyMountTargetSecurityGroupsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EFS Mount Target (%s) security groups: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn(ctx)

	log.Printf("[DEBUG] Deleting EFS Mount Target: %s", d.Id())
	_, err := conn.DeleteMountTargetWithContext(ctx, &efs.DeleteMountTargetInput{
		MountTargetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Mount Target (%s): %s", d.Id(), err)
	}

	if _, err := waitMountTargetDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Mount Target (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func getAZFromSubnetID(ctx context.Context, conn *ec2.EC2, subnetID string) (string, error) {
	subnet, err := tfec2.FindSubnetByID(ctx, conn, subnetID)

	if err != nil {
		return "", err
	}

	return aws.StringValue(subnet.AvailabilityZone), nil
}

func FindMountTargetByID(ctx context.Context, conn *efs.EFS, id string) (*efs.MountTargetDescription, error) {
	input := &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(id),
	}

	output, err := conn.DescribeMountTargetsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeMountTargetNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.MountTargets) == 0 || output.MountTargets[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.MountTargets); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.MountTargets[0], nil
}

func statusMountTargetLifeCycleState(ctx context.Context, conn *efs.EFS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMountTargetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.LifeCycleState), nil
	}
}

func waitMountTargetCreated(ctx context.Context, conn *efs.EFS, id string, timeout time.Duration) (*efs.MountTargetDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{efs.LifeCycleStateCreating},
		Target:     []string{efs.LifeCycleStateAvailable},
		Refresh:    statusMountTargetLifeCycleState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.MountTargetDescription); ok {
		return output, err
	}

	return nil, err
}

func waitMountTargetDeleted(ctx context.Context, conn *efs.EFS, id string, timeout time.Duration) (*efs.MountTargetDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:     []string{},
		Refresh:    statusMountTargetLifeCycleState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.MountTargetDescription); ok {
		return output, err
	}

	return nil, err
}
