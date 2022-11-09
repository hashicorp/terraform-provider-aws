package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayPeeringAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayPeeringAttachmentCreate,
		Read:   resourceTransitGatewayPeeringAttachmentRead,
		Update: resourceTransitGatewayPeeringAttachmentUpdate,
		Delete: resourceTransitGatewayPeeringAttachmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"peer_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTransitGatewayPeeringAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	peerAccountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("peer_account_id"); ok {
		peerAccountID = v.(string)
	}
	input := &ec2.CreateTransitGatewayPeeringAttachmentInput{
		PeerAccountId:        aws.String(peerAccountID),
		PeerRegion:           aws.String(d.Get("peer_region").(string)),
		PeerTransitGatewayId: aws.String(d.Get("peer_transit_gateway_id").(string)),
		TagSpecifications:    tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
		TransitGatewayId:     aws.String(d.Get("transit_gateway_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Peering Attachment: %s", input)
	output, err := conn.CreateTransitGatewayPeeringAttachment(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway Peering Attachment: %w", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if _, err := WaitTransitGatewayPeeringAttachmentCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Peering Attachment (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayPeeringAttachmentRead(d, meta)
}

func resourceTransitGatewayPeeringAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayPeeringAttachment, err := FindTransitGatewayPeeringAttachmentByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Peering Attachment (%s): %w", d.Id(), err)
	}

	d.Set("peer_account_id", transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.AccepterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	d.Set("transit_gateway_id", transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)

	tags := KeyValueTags(transitGatewayPeeringAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayPeeringAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Transit Gateway Peering Attachment (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayPeeringAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayPeeringAttachment(&ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Peering Attachment (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPeeringAttachmentDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Peering Attachment (%s) delete: %w", d.Id(), err)
	}

	return nil
}
