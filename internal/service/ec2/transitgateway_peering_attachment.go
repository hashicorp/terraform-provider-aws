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

	peerAccountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("peer_account_id"); ok {
		peerAccountId = v.(string)
	}
	input := &ec2.CreateTransitGatewayPeeringAttachmentInput{
		PeerAccountId:        aws.String(peerAccountId),
		PeerRegion:           aws.String(d.Get("peer_region").(string)),
		PeerTransitGatewayId: aws.String(d.Get("peer_transit_gateway_id").(string)),
		TagSpecifications:    tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeTransitGatewayAttachment),
		TransitGatewayId:     aws.String(d.Get("transit_gateway_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Peering Attachment: %s", input)
	output, err := conn.CreateTransitGatewayPeeringAttachment(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Peering Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if err := waitForTransitGatewayPeeringAttachmentCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) availability: %s", d.Id(), err)
	}

	return resourceTransitGatewayPeeringAttachmentRead(d, meta)
}

func resourceTransitGatewayPeeringAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	transitGatewayPeeringAttachment, err := DescribeTransitGatewayPeeringAttachment(conn, d.Id())

	if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if transitGatewayPeeringAttachment == nil {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleting || aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(transitGatewayPeeringAttachment.State))
		d.SetId("")
		return nil
	}

	d.Set("peer_account_id", transitGatewayPeeringAttachment.AccepterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.AccepterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	d.Set("transit_gateway_id", transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)

	tags := KeyValueTags(transitGatewayPeeringAttachment.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceTransitGatewayPeeringAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceTransitGatewayPeeringAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), input)
	_, err := conn.DeleteTransitGatewayPeeringAttachment(input)

	if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment: %s", err)
	}

	if err := WaitForTransitGatewayPeeringAttachmentDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %s", d.Id(), err)
	}

	return nil
}
