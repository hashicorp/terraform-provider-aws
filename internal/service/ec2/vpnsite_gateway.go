package ec2

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPNGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPNGatewayCreate,
		Read:   resourceVPNGatewayRead,
		Update: resourceVPNGatewayUpdate,
		Delete: resourceVPNGatewayDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourceVPNGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateVpnGatewayInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVpnGateway),
		Type:              aws.String(ec2.GatewayTypeIpsec1),
	}

	if v, ok := d.GetOk("amazon_side_asn"); ok {
		v, err := strconv.ParseInt(v.(string), 10, 64)

		if err != nil {
			return err
		}

		input.AmazonSideAsn = aws.Int64(v)
	}

	log.Printf("[DEBUG] Creating EC2 VPN Gateway: %s", input)
	output, err := conn.CreateVpnGateway(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 VPN Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.VpnGateway.VpnGatewayId))

	if v, ok := d.GetOk("vpc_id"); ok {
		if err := attachVPNGatewayToVPC(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindVPNGatewayByID(conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Gateway (%s): %w", d.Id(), err)
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
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceVPNGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("vpc_id") {
		o, n := d.GetChange("vpc_id")

		if vpcID, ok := o.(string); ok && vpcID != "" {
			if err := detachVPNGatewayFromVPC(conn, d.Id(), vpcID); err != nil {
				return err
			}
		}

		if vpcID, ok := n.(string); ok && vpcID != "" {
			if err := attachVPNGatewayToVPC(conn, d.Id(), vpcID); err != nil {
				return err
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPN Gateway (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if v, ok := d.GetOk("vpc_id"); ok {
		if err := detachVPNGatewayFromVPC(conn, d.Id(), v.(string)); err != nil {
			return err
		}
	}

	log.Printf("[INFO] Deleting EC2 VPN Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(VPNGatewayDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteVpnGateway(&ec2.DeleteVpnGatewayInput{
			VpnGatewayId: aws.String(d.Id()),
		})
	}, ErrCodeIncorrectState)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpnGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPN Gateway (%s): %w", d.Id(), err)
	}

	return nil
}

func attachVPNGatewayToVPC(conn *ec2.EC2, vpnGatewayID, vpcID string) error {
	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[INFO] Attaching EC2 VPN Gateway (%s) to VPC (%s)", vpnGatewayID, vpcID)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, func() (interface{}, error) {
		return conn.AttachVpnGateway(input)
	}, ErrCodeInvalidVpnGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("error attaching EC2 VPN Gateway (%s) to VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentAttached(conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Gateway (%s) to attach to VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	return nil
}

func detachVPNGatewayFromVPC(conn *ec2.EC2, vpnGatewayID, vpcID string) error {
	input := &ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[INFO] Detaching EC2 VPN Gateway (%s) from VPC (%s)", vpnGatewayID, vpcID)
	_, err := conn.DetachVpnGateway(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpnGatewayAttachmentNotFound, ErrCodeInvalidVpnGatewayIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error detaching EC2 VPN Gateway (%s) from VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	_, err = WaitVPNGatewayVPCAttachmentDetached(conn, vpnGatewayID, vpcID)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 VPN Gateway (%s) to detach from VPC (%s): %w", vpnGatewayID, vpcID, err)
	}

	return nil
}
