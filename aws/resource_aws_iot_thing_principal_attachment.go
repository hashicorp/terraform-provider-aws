package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceThingPrincipalAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceThingPrincipalAttachmentCreate,
		Read:   resourceThingPrincipalAttachmentRead,
		Delete: resourceThingPrincipalAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"thing": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingPrincipalAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	_, err := conn.AttachThingPrincipal(&iot.AttachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if err != nil {
		return fmt.Errorf("error attaching principal %s to thing %s: %s", principal, thing, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", thing, principal))
	return resourceThingPrincipalAttachmentRead(d, meta)
}

func getIoTThingPricipalAttachment(conn *iot.IoT, thing, principal string) (bool, error) {
	out, err := conn.ListThingPrincipals(&iot.ListThingPrincipalsInput{
		ThingName: aws.String(thing),
	})
	if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
		return false, nil
	} else if err != nil {
		return false, err
	}
	found := false
	for _, name := range out.Principals {
		if principal == aws.StringValue(name) {
			found = true
			break
		}
	}
	return found, nil
}

func resourceThingPrincipalAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	found, err := getIoTThingPricipalAttachment(conn, thing, principal)

	if err != nil {
		return fmt.Errorf("error listing principals for thing %s: %s", thing, err)
	}

	if !found {
		log.Printf("[WARN] IoT Thing Principal Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	return nil
}

func resourceThingPrincipalAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	_, err := conn.DetachThingPrincipal(&iot.DetachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if tfawserr.ErrMessageContains(err, iot.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] IoT Principal %s or Thing %s not found, removing from state", principal, thing)
	} else if err != nil {
		return fmt.Errorf("error detaching principal %s from thing %s: %s", principal, thing, err)
	}

	return nil
}
