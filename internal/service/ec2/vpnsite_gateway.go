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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpn_gateway", name="VPN Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNGatewayCreate,
		ReadWithoutTimeout:   resourceVPNGatewayRead,
		UpdateWithoutTimeout: resourceVPNGatewayUpdate,
		DeleteWithoutTimeout: resourceVPNGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAmazonSideASN,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPNGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpnGatewayInput{
		AvailabilityZone:  aws.String(d.Get(names.AttrAvailabilityZone).(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeVpnGateway),
		Type:              awstypes.GatewayTypeIpsec1,
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		input.AmazonSideAsn = flex.StringValueToInt64(v.(string))
	}

	output, err := conn.CreateVpnGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.VpnGateway.VpnGatewayId))

	if _, err := waitVPNGatewayCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Gateway (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		if err := attachVPNGatewayToVPC(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceVPNGatewayRead(ctx, d, meta)...)
}

func resourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findVPNGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Gateway (%s): %s", d.Id(), err)
	}

	vpnGateway := outputRaw.(*awstypes.VpnGateway)

	d.Set("amazon_side_asn", flex.Int64ToStringValue(vpnGateway.AmazonSideAsn))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-gateway/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if aws.ToString(vpnGateway.AvailabilityZone) != "" {
		d.Set(names.AttrAvailabilityZone, vpnGateway.AvailabilityZone)
	}
	d.Set(names.AttrVPCID, nil)
	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if vpcAttachment.State == awstypes.AttachmentStatusAttached {
			d.Set(names.AttrVPCID, vpcAttachment.VpcId)
		}
	}

	setTagsOut(ctx, vpnGateway.Tags)

	return diags
}

func resourceVPNGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange(names.AttrVPCID) {
		o, n := d.GetChange(names.AttrVPCID)

		if vpcID, ok := o.(string); ok && vpcID != "" {
			if err := detachVPNGatewayFromVPC(ctx, conn, d.Id(), vpcID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if vpcID, ok := n.(string); ok && vpcID != "" {
			if err := attachVPNGatewayToVPC(ctx, conn, d.Id(), vpcID); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceVPNGatewayRead(ctx, d, meta)...)
}

func resourceVPNGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		if err := detachVPNGatewayFromVPC(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting EC2 VPN Gateway: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DeleteVpnGateway(ctx, &ec2.DeleteVpnGatewayInput{
			VpnGatewayId: aws.String(d.Id()),
		})
	}, errCodeIncorrectState)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPN Gateway (%s): %s", d.Id(), err)
	}

	if _, err := waitVPNGatewayDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Gateway (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func attachVPNGatewayToVPC(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) error {
	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return conn.AttachVpnGateway(ctx, input)
	}, errCodeInvalidVPNGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("attaching EC2 VPN Gateway (%s) to VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	if _, err := waitVPNGatewayVPCAttachmentAttached(ctx, conn, vpnGatewayID, vpcID); err != nil {
		return fmt.Errorf("waiting for EC2 VPN Gateway (%s) VPC (%s) attachment create: %w", vpnGatewayID, vpcID, err)
	}

	return nil
}

func detachVPNGatewayFromVPC(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) error {
	input := &ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	_, err := conn.DetachVpnGateway(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayAttachmentNotFound, errCodeInvalidVPNGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching EC2 VPN Gateway (%s) from VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	if _, err := waitVPNGatewayVPCAttachmentDetached(ctx, conn, vpnGatewayID, vpcID); err != nil {
		return fmt.Errorf("waiting for EC2 VPN Gateway (%s) VPC (%s) attachment delete: %w", vpnGatewayID, vpcID, err)
	}

	return nil
}
