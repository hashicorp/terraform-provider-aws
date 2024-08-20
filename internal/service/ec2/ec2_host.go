// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_host", name="Host")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceHost() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostCreate,
		ReadWithoutTimeout:   resourceHostRead,
		UpdateWithoutTimeout: resourceHostUpdate,
		DeleteWithoutTimeout: resourceHostDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"outpost_arn"},
				Computed:     true,
			},
			"auto_placement": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AutoPlacementOn,
				ValidateDiagFunc: enum.Validate[awstypes.AutoPlacement](),
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_recovery": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.HostRecoveryOff,
				ValidateDiagFunc: enum.Validate[awstypes.HostRecovery](),
			},
			"instance_family": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", names.AttrInstanceType},
			},
			names.AttrInstanceType: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"instance_family", names.AttrInstanceType},
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.AllocateHostsInput{
		AutoPlacement:     awstypes.AutoPlacement(d.Get("auto_placement").(string)),
		AvailabilityZone:  aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		ClientToken:       aws.String(id.UniqueId()),
		HostRecovery:      awstypes.HostRecovery(d.Get("host_recovery").(string)),
		Quantity:          aws.Int32(1),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeDedicatedHost),
	}

	if v, ok := d.GetOk("asset_id"); ok {
		input.AssetIds = []string{v.(string)}
	}

	if v, ok := d.GetOk("instance_family"); ok {
		input.InstanceFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		input.InstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		input.OutpostArn = aws.String(v.(string))
	}

	output, err := conn.AllocateHosts(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "allocating EC2 Host: %s", err)
	}

	d.SetId(output.HostIds[0])

	if _, err := waitHostCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	host, err := findHostByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Host %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Host (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.ToString(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("asset_id", host.AssetId)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set(names.AttrAvailabilityZone, host.AvailabilityZone)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set(names.AttrInstanceType, host.HostProperties.InstanceType)
	d.Set("outpost_arn", host.OutpostArn)
	d.Set(names.AttrOwnerID, host.OwnerId)

	setTagsOut(ctx, host.Tags)

	return diags
}

func resourceHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyHostsInput{
			HostIds: []string{d.Id()},
		}

		if d.HasChange("auto_placement") {
			input.AutoPlacement = awstypes.AutoPlacement(d.Get("auto_placement").(string))
		}

		if d.HasChange("host_recovery") {
			input.HostRecovery = awstypes.HostRecovery(d.Get("host_recovery").(string))
		}

		if hasChange, v := d.HasChange("instance_family"), d.Get("instance_family").(string); hasChange && v != "" {
			input.InstanceFamily = aws.String(v)
		}

		if hasChange, v := d.HasChange(names.AttrInstanceType), d.Get(names.AttrInstanceType).(string); hasChange && v != "" {
			input.InstanceType = aws.String(v)
		}

		output, err := conn.ModifyHosts(ctx, input)

		if err == nil && output != nil {
			err = unsuccessfulItemsError(output.Unsuccessful)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Host (%s): %s", d.Id(), err)
		}

		if _, err := waitHostUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceHostRead(ctx, d, meta)...)
}

func resourceHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 Host: %s", d.Id())
	output, err := conn.ReleaseHosts(ctx, &ec2.ReleaseHostsInput{
		HostIds: []string{d.Id()},
	})

	if err == nil && output != nil {
		err = unsuccessfulItemsError(output.Unsuccessful)
	}

	if tfawserr.ErrCodeEquals(err, errCodeClientInvalidHostIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "releasing EC2 Host (%s): %s", d.Id(), err)
	}

	if _, err := waitHostDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Host (%s) delete: %s", d.Id(), err)
	}

	return diags
}
