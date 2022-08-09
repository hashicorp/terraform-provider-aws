package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceTransitGatewayPrefixListReference() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayPrefixListReferenceCreate,
		Read:   resourceTransitGatewayPrefixListReferenceRead,
		Update: resourceTransitGatewayPrefixListReferenceUpdate,
		Delete: resourceTransitGatewayPrefixListReferenceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix_list_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_route_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayPrefixListReferenceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.CreateTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Prefix List Reference: %s", input)
	output, err := conn.CreateTransitGatewayPrefixListReference(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway Prefix List Reference: %w", err)
	}

	d.SetId(TransitGatewayPrefixListReferenceCreateResourceID(aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)))

	if _, err := WaitTransitGatewayPrefixListReferenceStateCreated(conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Prefix List Reference (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayPrefixListReferenceRead(d, meta)
}

func resourceTransitGatewayPrefixListReferenceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, prefixListID, err := TransitGatewayPrefixListReferenceParseResourceID(d.Id())

	if err != nil {
		return err
	}

	transitGatewayPrefixListReference, err := FindTransitGatewayPrefixListReferenceByTwoPartKey(conn, transitGatewayRouteTableID, prefixListID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Prefix List Reference (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	d.Set("blackhole", transitGatewayPrefixListReference.Blackhole)
	d.Set("prefix_list_id", transitGatewayPrefixListReference.PrefixListId)
	d.Set("prefix_list_owner_id", transitGatewayPrefixListReference.PrefixListOwnerId)
	if transitGatewayPrefixListReference.TransitGatewayAttachment == nil {
		d.Set("transit_gateway_attachment_id", nil)
	} else {
		d.Set("transit_gateway_attachment_id", transitGatewayPrefixListReference.TransitGatewayAttachment.TransitGatewayAttachmentId)
	}
	d.Set("transit_gateway_route_table_id", transitGatewayPrefixListReference.TransitGatewayRouteTableId)

	return nil
}

func resourceTransitGatewayPrefixListReferenceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.ModifyTransitGatewayPrefixListReferenceInput{}

	if v, ok := d.GetOk("blackhole"); ok {
		input.Blackhole = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_attachment_id"); ok {
		input.TransitGatewayAttachmentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("transit_gateway_route_table_id"); ok {
		input.TransitGatewayRouteTableId = aws.String(v.(string))
	}

	output, err := conn.ModifyTransitGatewayPrefixListReference(input)

	if err != nil {
		return fmt.Errorf("updating EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPrefixListReferenceStateUpdated(conn, aws.StringValue(output.TransitGatewayPrefixListReference.TransitGatewayRouteTableId), aws.StringValue(output.TransitGatewayPrefixListReference.PrefixListId)); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Prefix List Reference (%s) update: %w", d.Id(), err)
	}

	return resourceTransitGatewayPrefixListReferenceRead(d, meta)
}

func resourceTransitGatewayPrefixListReferenceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, prefixListID, err := TransitGatewayPrefixListReferenceParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Prefix List Reference: %s", d.Id())
	_, err = conn.DeleteTransitGatewayPrefixListReference(&ec2.DeleteTransitGatewayPrefixListReferenceInput{
		PrefixListId:               aws.String(prefixListID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Prefix List Reference (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPrefixListReferenceStateDeleted(conn, transitGatewayRouteTableID, prefixListID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Prefix List Reference (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const transitGatewayPrefixListReferenceIDSeparator = "_"

func TransitGatewayPrefixListReferenceCreateResourceID(transitGatewayRouteTableID string, prefixListID string) string {
	parts := []string{transitGatewayRouteTableID, prefixListID}
	id := strings.Join(parts, transitGatewayPrefixListReferenceIDSeparator)

	return id
}

func TransitGatewayPrefixListReferenceParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPrefixListReferenceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sPREFIX-LIST-ID", id, transitGatewayPrefixListReferenceIDSeparator)
}
