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

func ResourceTransitGatewayPolicyTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayPolicyTableAssociationCreate,
		Read:   resourceTransitGatewayPolicyTableAssociationRead,
		Delete: resourceTransitGatewayPolicyTableAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_policy_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayPolicyTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	// If the TGW attachment is already associated with a TGW route table, disassociate it to prevent errors like
	// "IncorrectState: Cannot have both PolicyTableAssociation and RouteTableAssociation on the same TransitGateway Attachment".
	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	transitGatewayAttachment, err := FindTransitGatewayAttachmentByID(conn, transitGatewayAttachmentID)

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Attachment (%s): %w", transitGatewayAttachmentID, err)
	}

	if v := transitGatewayAttachment.Association; v != nil {
		if transitGatewayRouteTableID := aws.StringValue(v.TransitGatewayRouteTableId); transitGatewayRouteTableID != "" && aws.StringValue(v.State) == ec2.TransitGatewayAssociationStateAssociated {
			id := TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
			input := &ec2.DisassociateTransitGatewayRouteTableInput{
				TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
				TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
			}

			if _, err := conn.DisassociateTransitGatewayRouteTable(input); err != nil {
				return fmt.Errorf("deleting EC2 Transit Gateway Route Table Association (%s): %w", id, err)
			}

			if _, err := WaitTransitGatewayRouteTableAssociationDeleted(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
				return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Association (%s) delete: %w", id, err)
			}
		}
	}

	transitGatewayPolicyTableID := d.Get("transit_gateway_policy_table_id").(string)
	id := TransitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID)
	input := &ec2.AssociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	_, err = conn.AssociateTransitGatewayPolicyTable(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway Policy Table Association (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitTransitGatewayPolicyTableAssociationCreated(conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Policy Table Association (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayPolicyTableAssociationRead(d, meta)
}

func resourceTransitGatewayPolicyTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := TransitGatewayPolicyTableAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	transitGatewayPolicyTableAssociation, err := FindTransitGatewayPolicyTableAssociationByTwoPartKey(conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Policy Table Association %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Transit Gateway Policy Table Association (%s): %w", d.Id(), err)
	}

	d.Set("resource_id", transitGatewayPolicyTableAssociation.ResourceId)
	d.Set("resource_type", transitGatewayPolicyTableAssociation.ResourceType)
	d.Set("transit_gateway_attachment_id", transitGatewayPolicyTableAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_policy_table_id", transitGatewayPolicyTableAssociation.TransitGatewayPolicyTableId)

	return nil
}

func resourceTransitGatewayPolicyTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayPolicyTableID, transitGatewayAttachmentID, err := TransitGatewayPolicyTableAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Policy Table Association: %s", d.Id())
	_, err = conn.DisassociateTransitGatewayPolicyTable(&ec2.DisassociateTransitGatewayPolicyTableInput{
		TransitGatewayAttachmentId:  aws.String(transitGatewayAttachmentID),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Policy Table Association (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayPolicyTableAssociationDeleted(conn, transitGatewayPolicyTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Policy Table Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const transitGatewayPolicyTableAssociationIDSeparator = "_"

func TransitGatewayPolicyTableAssociationCreateResourceID(transitGatewayPolicyTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayPolicyTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayPolicyTableAssociationIDSeparator)

	return id
}

func TransitGatewayPolicyTableAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayPolicyTableAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-POLICY-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayPolicyTableAssociationIDSeparator)
}
