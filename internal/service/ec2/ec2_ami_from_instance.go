// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ami_from_instance", name="AMI")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceAMIFromInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAMIFromInstanceCreate,
		// The remaining operations are shared with the generic aws_ami resource,
		// since the aws_ami_from_instance resource only differs in how it's created.
		ReadWithoutTimeout:   resourceAMIRead,
		UpdateWithoutTimeout: resourceAMIUpdate,
		DeleteWithoutTimeout: resourceAMIDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(amiRetryTimeout),
			Update: schema.DefaultTimeout(amiRetryTimeout),
			Delete: schema.DefaultTimeout(amiDeleteTimeout),
		},

		// Keep in sync with aws_ami's schema.
		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"boot_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deprecation_time": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.IsRFC3339Time,
				DiffSuppressFunc:      verify.SuppressEquivalentRoundedTime(time.RFC3339, time.Minute),
				DiffSuppressOnRefresh: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			// The following block device attributes intentionally mimick the
			// corresponding attributes on aws_instance, since they have the
			// same meaning.
			// However, we don't use root_block_device here because the constraint
			// on which root device attributes can be overridden for an instance to
			// not apply when registering an AMI.
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEncrypted: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"outpost_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSnapshotID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrThroughput: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrVolumeType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrSnapshotID].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"ena_support": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m[names.AttrVirtualName].(string)))
					return create.StringHashcode(buf.String())
				},
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"imds_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Not a public attribute; used to let the aws_ami_copy and aws_ami_from_instance
			// resources record that they implicitly created new EBS snapshots that we should
			// now manage. Not set by aws_ami, since the snapshots used there are presumed to
			// be independently managed.
			"manage_ebs_snapshots": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_details": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ramdisk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_device_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"snapshot_without_reboot": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"sriov_net_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"tpm_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_operation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtualization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAMIFromInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	instanceID := d.Get("source_instance_id").(string)
	name := d.Get(names.AttrName).(string)
	input := &ec2.CreateImageInput{
		Description:       aws.String(d.Get(names.AttrDescription).(string)),
		InstanceId:        aws.String(instanceID),
		Name:              aws.String(name),
		NoReboot:          aws.Bool(d.Get("snapshot_without_reboot").(bool)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeImage),
	}

	output, err := conn.CreateImage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s) from EC2 Instance (%s): %s", name, instanceID, err)
	}

	d.SetId(aws.ToString(output.ImageId))
	d.Set("manage_ebs_snapshots", true)

	if _, err := waitImageAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s) from EC2 Instance (%s): waiting for completion: %s", name, instanceID, err)
	}

	if v, ok := d.GetOk("deprecation_time"); ok {
		if err := enableImageDeprecation(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 AMI (%s) from EC2 Instance (%s): %s", name, instanceID, err)
		}
	}

	return append(diags, resourceAMIRead(ctx, d, meta)...)
}
