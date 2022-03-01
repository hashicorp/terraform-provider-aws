package ec2

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayConnectPeerCreate,
		Read:   resourceTransitGatewayConnectPeerRead,
		Update: resourceTransitGatewayConnectPeerUpdate,
		Delete: resourceTransitGatewayConnectPeerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"bgp_asn": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: valid4ByteASN,
			},
			"inside_cidr_blocks": {
				// TODO: TypeSet
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: verify.IsIPv4CIDRBlockOrIPv6CIDRBlock(
						validation.All(
							validation.IsCIDRNetwork(29, 29),
							validation.StringMatch(regexp.MustCompile(`^169\.254\.`), "IPv4 range must be from range 169.254.0.0/16"),
							validation.StringDoesNotMatch(regexp.MustCompile(`^169\.254\.([0-5]\.0|169\.248)/29`), "IPv4 range must not be 169.254.([0-5].0|169.248)/29"),
						),
						validation.All(
							validation.IsCIDRNetwork(125, 125),
							validation.StringMatch(regexp.MustCompile(`^[fF][dD]`), "IPv6 range must be from fd00::/8"),
						),
					),
				},
			},
			"peer_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.IsIPv4Address,
					validation.IsIPv6Address,
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.IsIPv4Address,
					validation.IsIPv6Address,
				),
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayConnectPeerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	insideCidrBlocks := d.Get("inside_cidr_blocks").([]interface{})
	peerAddress := d.Get("peer_address").(string)

	input := &ec2.CreateTransitGatewayConnectPeerInput{
		InsideCidrBlocks:           flex.ExpandStringList(insideCidrBlocks),
		PeerAddress:                aws.String(peerAddress),
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TagSpecifications:          ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayConnectPeer),
	}

	if v, ok := d.GetOk("bgp_asn"); ok {
		bgpOptions := ec2.TransitGatewayConnectRequestBgpOptions{
			PeerAsn: aws.Int64(int64(v.(int))),
		}

		input.BgpOptions = &bgpOptions
	}

	transitGatewayAddress, transitGatewayAddressOk := d.GetOk("transit_gateway_address")
	if transitGatewayAddressOk {
		input.TransitGatewayAddress = aws.String(transitGatewayAddress.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Peer: %s", input)
	output, err := conn.CreateTransitGatewayConnectPeer(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Connect Peer: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnectPeer.TransitGatewayConnectPeerId))

	if err := waitForTransitGatewayConnectPeerCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) availability: %s", d.Id(), err)
	}

	return resourceTransitGatewayConnectPeerRead(d, meta)
}

func resourceTransitGatewayConnectPeerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayConnectPeer, err := DescribeTransitGatewayConnectPeer(conn, d.Id())

	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
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

	if aws.StringValue(transitGatewayConnectPeer.State) == ec2.TransitGatewayConnectPeerStateDeleting || aws.StringValue(transitGatewayConnectPeer.State) == ec2.TransitGatewayConnectPeerStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayConnectPeer.State))
		d.SetId("")
		return nil
	}

	transitGatewayAttachmentID := aws.StringValue(transitGatewayConnectPeer.TransitGatewayAttachmentId)
	tags := KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("bgp_asn", transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations[0].PeerAsn)
	d.Set("inside_cidr_blocks", transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks)
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set("transit_gateway_attachment_id", transitGatewayAttachmentID)

	return nil
}

func resourceTransitGatewayConnectPeerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Connect Peer (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayConnectPeerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect Peer (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayConnectPeer(input)

	if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayConnectPeerID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Connect Peer: %s", err)
	}

	if err := WaitForTransitGatewayConnectPeerDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
