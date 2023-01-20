package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPNGateway() *schema.Resource {
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
				ValidateFunc: validAmazonSideASN,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
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
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVpnGatewayInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpnGateway),
		Type:              aws.String(ec2.GatewayTypeIpsec1),
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		v, err := strconv.ParseInt(v.(string), 10, 64)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Gateway: %s", err)
		}

		input.AmazonSideAsn = aws.Int64(v)
	}

	log.Printf("[DEBUG] Creating EC2 VPN Gateway: %s", input)
	output, err := conn.CreateVpnGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.VpnGateway.VpnGatewayId))

	if v, ok := d.GetOk("vpc_id"); ok {
		if err := attachVPNGatewayToVPC(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Gateway: %s", err)
		}
	}

	return append(diags, resourceVPNGatewayRead(ctx, d, meta)...)
}

func resourceVPNGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindVPNGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Gateway (%s): %s", d.Id(), err)
	}

	vpnGateway := outputRaw.(*ec2.VpnGateway)

	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(vpnGateway.AmazonSideAsn), 10))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpn-gateway/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	if aws.StringValue(vpnGateway.AvailabilityZone) != "" {
		d.Set("availability_zone", vpnGateway.AvailabilityZone)
	}

	d.Set("vpc_id", nil)
	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusAttached {
			d.Set("vpc_id", vpcAttachment.VpcId)
		}
	}

	tags := KeyValueTags(vpnGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceVPNGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("vpc_id") {
		o, n := d.GetChange("vpc_id")

		if vpcID, ok := o.(string); ok && vpcID != "" {
			if err := detachVPNGatewayFromVPC(ctx, conn, d.Id(), vpcID); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 VPN Gateway (%s): %s", d.Id(), err)
			}
		}

		if vpcID, ok := n.(string); ok && vpcID != "" {
			if err := attachVPNGatewayToVPC(ctx, conn, d.Id(), vpcID); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 VPN Gateway (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 VPN Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPNGatewayRead(ctx, d, meta)...)
}

func resourceVPNGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if v, ok := d.GetOk("vpc_id"); ok {
		if err := detachVPNGatewayFromVPC(ctx, conn, d.Id(), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EC2 VPN Gateway (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting EC2 VPN Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, VPNGatewayDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteVpnGatewayWithContext(ctx, &ec2.DeleteVpnGatewayInput{
			VpnGatewayId: aws.String(d.Id()),
		})
	}, errCodeIncorrectState)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPN Gateway (%s): %s", d.Id(), err)
	}

	return diags
}

func attachVPNGatewayToVPC(ctx context.Context, conn *ec2.EC2, vpnGatewayID, vpcID string) error {
	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[INFO] Attaching EC2 VPN Gateway (%s) to VPC (%s)", vpnGatewayID, vpcID)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AttachVpnGatewayWithContext(ctx, input)
	}, errCodeInvalidVPNGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("attaching to VPC (%s): %w", vpcID, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentAttached(ctx, conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("attaching to VPC (%s): waiting for completion: %w", vpcID, err)
	}

	return nil
}

func detachVPNGatewayFromVPC(ctx context.Context, conn *ec2.EC2, vpnGatewayID, vpcID string) error {
	input := &ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[INFO] Detaching EC2 VPN Gateway (%s) from VPC (%s)", vpnGatewayID, vpcID)
	_, err := conn.DetachVpnGatewayWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayAttachmentNotFound, errCodeInvalidVPNGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching from VPC (%s): %w", vpcID, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentDetached(ctx, conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("detaching from VPC (%s): waiting for completion: %w", vpcID, err)
	}

	return nil
}
