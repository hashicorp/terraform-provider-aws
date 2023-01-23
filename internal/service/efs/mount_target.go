package efs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

const (
	mountTargetDeleteTimeout = 10 * time.Minute
)

func ResourceMountTarget() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMountTargetCreate,
		ReadWithoutTimeout:   resourceMountTargetRead,
		UpdateWithoutTimeout: resourceMountTargetUpdate,
		DeleteWithoutTimeout: resourceMountTargetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Optional: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_target_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMountTargetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	fsId := d.Get("file_system_id").(string)
	subnetId := d.Get("subnet_id").(string)

	// CreateMountTarget would return the same Mount Target ID
	// to parallel requests if they both include the same AZ
	// and we would end up managing the same MT as 2 resources.
	// So we make it fail by calling 1 request per AZ at a time.
	az, err := getAzFromSubnetId(ctx, subnetId, meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Availability Zone from subnet ID (%s): %s", subnetId, err)
	}
	mtKey := "efs-mt-" + fsId + "-" + az
	conns.GlobalMutexKV.Lock(mtKey)
	defer conns.GlobalMutexKV.Unlock(mtKey)

	input := efs.CreateMountTargetInput{
		FileSystemId: aws.String(fsId),
		SubnetId:     aws.String(subnetId),
	}

	if v, ok := d.GetOk("ip_address"); ok {
		input.IpAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("security_groups"); ok {
		input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	mt, err := conn.CreateMountTargetWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Mount Target (%s): %s", fsId, err)
	}

	d.SetId(aws.StringValue(mt.MountTargetId))
	log.Printf("[INFO] EFS mount target ID: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeMountTargetsWithContext(ctx, &efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(d.Id()),
			})
			if err != nil {
				return nil, "error", err
			}

			if HasEmptyMountTargets(resp) {
				return nil, "error", fmt.Errorf("EFS mount target %q could not be found.", d.Id())
			}

			mt := resp.MountTargets[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.MountTargetId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    30 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS mount target (%s) to create: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS mount target created: %s", aws.StringValue(mt.MountTargetId))

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	if d.HasChange("security_groups") {
		input := efs.ModifyMountTargetSecurityGroupsInput{
			MountTargetId:  aws.String(d.Id()),
			SecurityGroups: flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}
		_, err := conn.ModifyMountTargetSecurityGroupsWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EFS Mount Target (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMountTargetRead(ctx, d, meta)...)
}

func resourceMountTargetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()
	resp, err := conn.DescribeMountTargetsWithContext(ctx, &efs.DescribeMountTargetsInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, efs.ErrCodeMountTargetNotFound) {
			// The EFS mount target could not be found,
			// which would indicate that it might be
			// already deleted.
			log.Printf("[WARN] EFS mount target %q could not be found.", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading EFS mount target %s: %s", d.Id(), err)
	}

	if HasEmptyMountTargets(resp) {
		return sdkdiag.AppendErrorf(diags, "EFS mount target %q could not be found.", d.Id())
	}

	mt := resp.MountTargets[0]

	log.Printf("[DEBUG] Found EFS mount target: %#v", mt)

	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(mt.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("file_system_arn", fsARN)
	d.Set("file_system_id", mt.FileSystemId)
	d.Set("ip_address", mt.IpAddress)
	d.Set("subnet_id", mt.SubnetId)
	d.Set("network_interface_id", mt.NetworkInterfaceId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("owner_id", mt.OwnerId)

	sgResp, err := conn.DescribeMountTargetSecurityGroupsWithContext(ctx, &efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s): %s", d.Id(), err)
	}

	err = d.Set("security_groups", flex.FlattenStringSet(sgResp.SecurityGroups))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s): %s", d.Id(), err)
	}

	d.Set("dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(mt.FileSystemId))))
	d.Set("mount_target_dns_name", meta.(*conns.AWSClient).RegionalHostname(fmt.Sprintf("%s.%s.efs", aws.StringValue(mt.AvailabilityZoneName), aws.StringValue(mt.FileSystemId))))

	return diags
}

func getAzFromSubnetId(ctx context.Context, subnetId string, meta interface{}) (string, error) {
	conn := meta.(*conns.AWSClient).EC2Conn()
	subnet, err := ec2.FindSubnetByID(ctx, conn, subnetId)
	if err != nil {
		return "", err
	}

	return aws.StringValue(subnet.AvailabilityZone), nil
}

func resourceMountTargetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSConn()

	_, err := conn.DeleteMountTargetWithContext(ctx, &efs.DeleteMountTargetInput{
		MountTargetId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Mount Target (%s): %s", d.Id(), err)
	}

	err = WaitForDeleteMountTarget(ctx, conn, d.Id(), mountTargetDeleteTimeout)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Mount Target (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func WaitForDeleteMountTarget(ctx context.Context, conn *efs.EFS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeMountTargetsWithContext(ctx, &efs.DescribeMountTargetsInput{
				MountTargetId: aws.String(id),
			})
			if err != nil {
				if tfawserr.ErrCodeEquals(err, efs.ErrCodeMountTargetNotFound) {
					return nil, "", nil
				}

				return nil, "error", err
			}

			if HasEmptyMountTargets(resp) {
				return nil, "", nil
			}

			mt := resp.MountTargets[0]

			log.Printf("[DEBUG] Current status of %q: %q", aws.StringValue(mt.MountTargetId), aws.StringValue(mt.LifeCycleState))
			return mt, aws.StringValue(mt.LifeCycleState), nil
		},
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func HasEmptyMountTargets(mto *efs.DescribeMountTargetsOutput) bool {
	if mto != nil && len(mto.MountTargets) > 0 {
		return false
	}
	return true
}
