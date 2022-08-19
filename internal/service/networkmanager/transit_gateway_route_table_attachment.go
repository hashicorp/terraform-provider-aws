package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayRouteTableAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentCreate,
		ReadWithoutTimeout:   resourceTransitGatewayRouteTableAttachmentRead,
		UpdateWithoutTimeout: resourceTransitGatewayRouteTableAttachmentUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayRouteTableAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"peering_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"transit_gateway_route_table_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceTransitGatewayRouteTableAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	peeringID := d.Get("peering_id").(string)
	transitGatewayRouteTableARN := d.Get("transit_gateway_route_table_arn").(string)
	input := &networkmanager.CreateTransitGatewayRouteTableAttachmentInput{
		PeeringId:                   aws.String(peeringID),
		TransitGatewayRouteTableArn: aws.String(transitGatewayRouteTableARN),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Network Manager Transit Gateway Route Table Attachment: %s", input)
	output, err := conn.CreateTransitGatewayRouteTableAttachmentWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Network Manager Transit Gateway (%s) Route Table (%s) Attachment: %s", peeringID, transitGatewayRouteTableARN, err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayRouteTableAttachment.Attachment.AttachmentId))

	// TODO: Waiter.

	return resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)
}

func resourceTransitGatewayRouteTableAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	return nil
}

func resourceTransitGatewayRouteTableAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTagsWithContext(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Network Manager Transit Gateway Route Table Attachment (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceTransitGatewayRouteTableAttachmentRead(ctx, d, meta)
}

func resourceTransitGatewayRouteTableAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
