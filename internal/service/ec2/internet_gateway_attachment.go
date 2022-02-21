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
		Update: resourceInternetGatewayAttachmentUpdate,
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
			},
		},
	}
}

func resourceInternetGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwId := d.Get("internet_gateway_id").(string)
	vpcId := d.Get("vpc_id").(string)
	id := fmt.Sprintf("%s:%s", igwId, vpcId)

	if err := attachInternetGateway(conn, igwId, vpcId); err != nil {
		return fmt.Errorf("error Creating EC2 Internet Gateway Attachment (%s): %w", id, err)
	}

	d.SetId(id)

	return resourceInternetGatewayAttachmentRead(d, meta)
}

func resourceInternetGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwId, vpcId, err := InternetGatewayAttachmentResourceID(d.Id())
	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(PropagationTimeout, func() (interface{}, error) {
		return FindInternetGatewayAttachment(conn, igwId, vpcId)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Internet Gateway Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Internet Gateway Attachment (%s): %w", d.Id(), err)
	}

	ig := outputRaw.(*ec2.InternetGatewayAttachment)

	d.Set("internet_gateway_id", igwId)
	d.Set("vpc_id", ig.VpcId)

	return nil
}

func resourceInternetGatewayAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	igwId := d.Get("internet_gateway_id").(string)
	if d.HasChange("vpc_id") {
		o, n := d.GetChange("vpc_id")

		if v := o.(string); v != "" {
			if err := detachInternetGateway(conn, igwId, v); err != nil {
				return fmt.Errorf("error detaching EC2 Internet Gateway Attachment (%s): %w", d.Id(), err)
			}
		}

		if v := n.(string); v != "" {
			if err := attachInternetGateway(conn, igwId, v); err != nil {
				return fmt.Errorf("error attaching EC2 Internet Gateway Attachment (%s): %w", d.Id(), err)
			}
			d.SetId(fmt.Sprintf("%s:%s", igwId, v))
		}
	}

	return resourceInternetGatewayAttachmentRead(d, meta)
}

func resourceInternetGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if err := detachInternetGateway(conn, d.Get("internet_gateway_id").(string), d.Get("vpc_id").(string)); err != nil {
		return fmt.Errorf("error detaching EC2 Internet Gateway Attachment (%s): %w", d.Id(), err)
	}

	return nil
}

func InternetGatewayAttachmentResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected InternetGatewayId:VpcId", id)
}
