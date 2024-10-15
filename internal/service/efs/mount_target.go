// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_mount_target", name="Mount Target")
func resourceMountTarget() *schema.Resource {
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
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrIPAddress: {
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
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceMountTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	// CreateMountTarget would return the same Mount Target ID
	// to parallel requests if they both include the same AZ
	// and we would end up managing the same MT as 2 resources.
	// So we make it fail by calling 1 request per AZ at a time.
	subnetID := d.Get(names.AttrSubnetID).(string)
	az, err := getAZFromSubnetID(ctx, meta.(*conns.AWSClient).EC2Client(ctx), subnetID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnet (%s): %s", subnetID, err)
	}

	fsID := d.Get(names.AttrFileSystemID).(string)
	mtKey := "efs-mt-" + fsID + "-" + az
	conns.GlobalMutexKV.Lock(mtKey)
	defer conns.GlobalMutexKV.Unlock(mtKey)

	input := &efs.CreateMountTargetInput{
		FileSystemId: aws.String(fsID),
		SubnetId:     aws.String(subnetID),
	}

	if v, ok := d.GetOk(names.AttrIPAddress); ok {
		input.IpAddress = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecurityGroups); ok {
		input.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	mt, err := conn.CreateMountTarget(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Mount Target (%s): %s", fsID, err)
	}

	d.SetId(aws.ToString(mt.MountTargetId))

	if _, err := waitMountTargetCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Mount Target (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	mt, err := findMountTargetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Mount Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s): %s", d.Id(), err)
	}

	fsID := aws.ToString(mt.FileSystemId)
	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  "file-system/" + fsID,
		Service:   "elasticfilesystem",
	}.String()
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set(names.AttrDNSName, meta.(*conns.AWSClient).RegionalHostname(ctx, fsID+".efs"))
	d.Set("file_system_arn", fsARN)
	d.Set(names.AttrFileSystemID, fsID)
	d.Set(names.AttrIPAddress, mt.IpAddress)
	d.Set("mount_target_dns_name", meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s.%s.efs", aws.ToString(mt.AvailabilityZoneName), aws.ToString(mt.FileSystemId))))
	d.Set(names.AttrNetworkInterfaceID, mt.NetworkInterfaceId)
	d.Set(names.AttrOwnerID, mt.OwnerId)
	d.Set(names.AttrSubnetID, mt.SubnetId)

	output, err := conn.DescribeMountTargetSecurityGroups(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s) security groups: %s", d.Id(), err)
	}

	d.Set(names.AttrSecurityGroups, output.SecurityGroups)

	return diags
}

func resourceMountTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	if d.HasChange(names.AttrSecurityGroups) {
		input := &efs.ModifyMountTargetSecurityGroupsInput{
			MountTargetId:  aws.String(d.Id()),
			SecurityGroups: flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		}

		_, err := conn.ModifyMountTargetSecurityGroups(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EFS Mount Target (%s) security groups: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	log.Printf("[DEBUG] Deleting EFS Mount Target: %s", d.Id())
	_, err := conn.DeleteMountTarget(ctx, &efs.DeleteMountTargetInput{
		MountTargetId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.MountTargetNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Mount Target (%s): %s", d.Id(), err)
	}

	if _, err := waitMountTargetDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Mount Target (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func getAZFromSubnetID(ctx context.Context, conn *ec2.Client, subnetID string) (string, error) {
	subnet, err := tfec2.FindSubnetByID(ctx, conn, subnetID)

	if err != nil {
		return "", err
	}

	return aws.ToString(subnet.AvailabilityZone), nil
}

func findMountTarget(ctx context.Context, conn *efs.Client, input *efs.DescribeMountTargetsInput, filter tfslices.Predicate[*awstypes.MountTargetDescription]) (*awstypes.MountTargetDescription, error) {
	output, err := findMountTargets(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMountTargets(ctx context.Context, conn *efs.Client, input *efs.DescribeMountTargetsInput, filter tfslices.Predicate[*awstypes.MountTargetDescription]) ([]awstypes.MountTargetDescription, error) {
	var output []awstypes.MountTargetDescription

	pages := efs.NewDescribeMountTargetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.MountTargetNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.MountTargets {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findMountTargetByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.MountTargetDescription, error) {
	input := &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(id),
	}

	output, err := findMountTarget(ctx, conn, input, tfslices.PredicateTrue[*awstypes.MountTargetDescription]())

	if err != nil {
		return nil, err
	}

	if state := output.LifeCycleState; state == awstypes.LifeCycleStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusMountTargetLifeCycleState(ctx context.Context, conn *efs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findMountTargetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LifeCycleState), nil
	}
}

func waitMountTargetCreated(ctx context.Context, conn *efs.Client, id string, timeout time.Duration) (*awstypes.MountTargetDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.LifeCycleStateCreating),
		Target:     enum.Slice(awstypes.LifeCycleStateAvailable),
		Refresh:    statusMountTargetLifeCycleState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MountTargetDescription); ok {
		return output, err
	}

	return nil, err
}

func waitMountTargetDeleted(ctx context.Context, conn *efs.Client, id string, timeout time.Duration) (*awstypes.MountTargetDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting, awstypes.LifeCycleStateDeleted),
		Target:     []string{},
		Refresh:    statusMountTargetLifeCycleState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MountTargetDescription); ok {
		return output, err
	}

	return nil, err
}
