package ec2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayConnectPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayConnectPeerCreate,
		ReadWithoutTimeout:   resourceTransitGatewayConnectPeerRead,
		UpdateWithoutTimeout: resourceTransitGatewayConnectPeerUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayConnectPeerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: valid4ByteASN,
			},
			"inside_cidr_blocks": {
				Type:     schema.TypeSet,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPAddress,
			},
			"transit_gateway_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTransitGatewayConnectPeerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateTransitGatewayConnectPeerInput{
		InsideCidrBlocks:           flex.ExpandStringSet(d.Get("inside_cidr_blocks").(*schema.Set)),
		PeerAddress:                aws.String(d.Get("peer_address").(string)),
		TagSpecifications:          ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayConnectPeer),
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
	}

	if v, ok := d.GetOk("bgp_asn"); ok {
		v, err := strconv.ParseInt(v.(string), 10, 64)

		if err != nil {
			return diag.FromErr(err)
		}

		input.BgpOptions = &ec2.TransitGatewayConnectRequestBgpOptions{
			PeerAsn: aws.Int64(v),
		}
	}

	if v, ok := d.GetOk("transit_gateway_address"); ok {
		input.TransitGatewayAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Connect Peer: %s", input)
	output, err := conn.CreateTransitGatewayConnectPeerWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error creating EC2 Transit Gateway Connect Peer: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayConnectPeer.TransitGatewayConnectPeerId))

	if _, err := WaitTransitGatewayConnectPeerCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) create: %s", d.Id(), err)
	}

	return resourceTransitGatewayConnectPeerRead(ctx, d, meta)
}

func resourceTransitGatewayConnectPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayConnectPeer, err := FindTransitGatewayConnectPeerByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Connect Peer %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EC2 Transit Gateway Connect Peer (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-connect-peer/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("bgp_asn", strconv.FormatInt(aws.Int64Value(transitGatewayConnectPeer.ConnectPeerConfiguration.BgpConfigurations[0].PeerAsn), 10))
	d.Set("inside_cidr_blocks", aws.StringValueSlice(transitGatewayConnectPeer.ConnectPeerConfiguration.InsideCidrBlocks))
	d.Set("peer_address", transitGatewayConnectPeer.ConnectPeerConfiguration.PeerAddress)
	d.Set("transit_gateway_address", transitGatewayConnectPeer.ConnectPeerConfiguration.TransitGatewayAddress)
	d.Set("transit_gateway_attachment_id", transitGatewayConnectPeer.TransitGatewayAttachmentId)

	tags := KeyValueTags(transitGatewayConnectPeer.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceTransitGatewayConnectPeerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return diag.Errorf("error updating EC2 Transit Gateway Connect Peer (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceTransitGatewayConnectPeerRead(ctx, d, meta)
}

func resourceTransitGatewayConnectPeerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Connect Peer: %s", d.Id())
	_, err := conn.DeleteTransitGatewayConnectPeerWithContext(ctx, &ec2.DeleteTransitGatewayConnectPeerInput{
		TransitGatewayConnectPeerId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidTransitGatewayConnectPeerIDNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting EC2 Transit Gateway Connect Peer: %s", err)
	}

	if _, err := WaitTransitGatewayConnectPeerDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for EC2 Transit Gateway Connect Peer (%s) delete: %s", d.Id(), err)
	}

	return nil
}
