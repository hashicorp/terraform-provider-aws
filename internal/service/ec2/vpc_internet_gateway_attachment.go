package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceInternetGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceInternetGatewayAttachmentCreate,
		Read:   resourceInternetGatewayAttachmentRead,
		Delete: resourceInternetGatewayAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceInternetGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwID := d.Get("internet_gateway_id").(string)
	vpcID := d.Get("vpc_id").(string)

	if err := attachInternetGateway(conn, igwID, vpcID); err != nil {
		return err
	}

	d.SetId(InternetGatewayAttachmentCreateResourceID(igwID, vpcID))

	return resourceInternetGatewayAttachmentRead(d, meta)
}

func resourceInternetGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwID, vpcID, err := InternetGatewayAttachmentParseResourceID(d.Id())

	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindInternetGatewayAttachment(conn, igwID, vpcID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Internet Gateway Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Internet Gateway Attachment (%s): %w", d.Id(), err)
	}

	igw := outputRaw.(*ec2.InternetGatewayAttachment)

	d.Set("internet_gateway_id", igwID)
	d.Set("vpc_id", igw.VpcId)

	return nil
}

func resourceInternetGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwID, vpcID, err := InternetGatewayAttachmentParseResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := detachInternetGateway(conn, igwID, vpcID); err != nil {
		return err
	}

	return nil
}

const internetGatewayAttachmentIDSeparator = ":"

func InternetGatewayAttachmentCreateResourceID(igwID, vpcID string) string {
	parts := []string{igwID, vpcID}
	id := strings.Join(parts, internetGatewayAttachmentIDSeparator)

	return id
}

func InternetGatewayAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, internetGatewayAttachmentIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected INTERNET-GATEWAY-ID%[2]sVPC-ID", id, internetGatewayAttachmentIDSeparator)
}
