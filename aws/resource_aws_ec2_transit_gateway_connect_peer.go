package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TransitGatewayConnectPeerCreate,
		Read:   resourceAwsEc2TransitGatewayConnectPeerRead,
		Update: resourceAwsEc2TransitGatewayConnectPeerUpdate,
		Delete: resourceAwsEc2TransitGatewayConnectPeerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"peer_asn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"peer_address": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsEc2TransitGatewayConnectPeerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTransitGatewayConnectPeerInput{
		InsideCidrBlocks:           expandStringSet(d.Get("inside_cidr_blocks").(*schema.Set)),
		PeerAddress:                aws.String(d.Get("peer_address").(string)),
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
		TagSpecifications:          ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayConnectPeer),
	}

	if v, ok := d.GetOk("peer_asn"); ok {
		input.BgpOptions = &ec2.TransitGatewayConnectRequestBgpOptions{
			PeerAsn: aws.Int64(int64(v.(int))),
		}
	}

	if v, ok := d.GetOk("transit_gateway_address"); ok {
		input.TransitGatewayAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Peer: %s", input)
	output, err := conn.CreateTransitGatewayConnectPeer(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Connect Peer: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnectPeer.TransitGatewayConnectPeerId))

	if err := waitForEc2TransitGatewayConnectPeerCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsEc2TransitGatewayConnectPeerRead(d, meta)
}

func resourceAwsEc2TransitGatewayConnectPeerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	transitGatewayConnectPeer, err := ec2DescribeTransitGatewayConnectPeer(conn, d.Id())

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Connect Peer: %s", err)
	}

	if transitGatewayConnectPeer == nil {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayConnectPeer.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayConnectPeer.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayConnectPeer.State))
		d.SetId("")
		return nil
	}

	tags := keyvaluetags.Ec2KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("transit_gateway_attachment_id", aws.StringValue(transitGatewayConnectPeer.TransitGatewayAttachmentId))
	d.Set("peer_asn", aws.Int64Value(transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations[0].PeerAsn))
	d.Set("peer_address", aws.StringValue(transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress))
	d.Set("transit_gateway_address", aws.StringValue(transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress))
	if err := d.Set("inside_cidr_blocks", flattenStringList(transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks)); err != nil {
		return fmt.Errorf("error setting inside_cidr_blocks: %s", err)
	}

	return nil
}

func resourceAwsEc2TransitGatewayConnectPeerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Connect Peer (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEc2TransitGatewayConnectPeerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect Peer (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayConnectPeer(input)

	if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Connect Peer: %s", err)
	}

	if err := waitForEc2TransitGatewayConnectPeerDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
