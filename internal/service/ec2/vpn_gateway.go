package ec2

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	if _, ok := d.GetOk("vpc_id"); ok {
		if err := resourceVPNGatewayAttach(d, meta); err != nil {
			return fmt.Errorf("error attaching EC2 VPN Gateway (%s) to VPC: %s", d.Id(), err)
		}
	}

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	vpnGateway, err := FindVPNGatewayByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 VPN Gateway (%s): %w", d.Id(), err)
	}

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
	if d.HasChange("vpc_id") {
		// If we're already attached, detach it first
		if err := resourceVPNGatewayDetach(d, meta); err != nil {
			return err
		}

		// Attach the VPN gateway to the new vpc
		if err := resourceVPNGatewayAttach(d, meta); err != nil {
			return err
		}
	}

	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPN Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceVPNGatewayRead(d, meta)
}

func resourceVPNGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Detach if it is attached
	if err := resourceVPNGatewayDetach(d, meta); err != nil {
		return err
	}

	log.Printf("[INFO] Deleting EC2 VPN Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(VPNGatewayDeletedTimeout, func() (interface{}, error) {
		return conn.DeleteVpnGateway(&ec2.DeleteVpnGatewayInput{
			VpnGatewayId: aws.String(d.Id()),
		})
	}, ErrCodeIncorrectState, ErrCodeInvalidVpnGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("error deleting EC2 VPN Gateway (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceVPNGatewayAttach(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpcId := d.Get("vpc_id").(string)

	if vpcId == "" {
		log.Printf("[DEBUG] Not attaching VPN Gateway '%s' as no VPC ID is set", d.Id())
		return nil
	}

	log.Printf(
		"[INFO] Attaching VPN Gateway '%s' to VPC '%s'",
		d.Id(),
		vpcId)

	req := &ec2.AttachVpnGatewayInput{
		VpnGatewayId: aws.String(d.Id()),
		VpcId:        aws.String(vpcId),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.AttachVpnGateway(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, ErrCodeInvalidVpnGatewayIDNotFound, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.AttachVpnGateway(req)
	}

	if err != nil {
		return fmt.Errorf("Error attaching VPN gateway: %s", err)
	}

	// Wait for it to be fully attached before continuing
	log.Printf("[DEBUG] Waiting for VPN gateway (%s) to attach", d.Id())
	_, err = WaitVPNGatewayVPCAttachmentAttached(conn, d.Id(), vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become attached: %w", d.Id(), vpcId, err)
	}

	return nil
}

func resourceVPNGatewayDetach(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// Get the old VPC ID to detach from
	vpcIdRaw, _ := d.GetChange("vpc_id")
	vpcId := vpcIdRaw.(string)

	if vpcId == "" {
		log.Printf(
			"[DEBUG] Not detaching VPN Gateway '%s' as no VPC ID is set",
			d.Id())
		return nil
	}

	log.Printf(
		"[INFO] Detaching VPN Gateway '%s' from VPC '%s'",
		d.Id(),
		vpcId)

	_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
		VpnGatewayId: aws.String(d.Id()),
		VpcId:        aws.String(vpcId),
	})

	if tfawserr.ErrMessageContains(err, ErrCodeInvalidVpnGatewayAttachmentNotFound, "") || tfawserr.ErrMessageContains(err, ErrCodeInvalidVpnGatewayIDNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPN Gateway (%s) Attachment (%s): %w", d.Id(), vpcId, err)
	}

	// Wait for it to be fully detached before continuing
	_, err = WaitVPNGatewayVPCAttachmentDetached(conn, d.Id(), vpcId)

	if err != nil {
		return fmt.Errorf("error waiting for VPN Gateway (%s) Attachment (%s) to become detached: %w", d.Id(), vpcId, err)
	}

	return nil
}
