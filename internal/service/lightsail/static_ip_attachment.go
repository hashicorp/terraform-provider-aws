package lightsail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceStaticIPAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceStaticIPAttachmentCreate,
		Read:   resourceStaticIPAttachmentRead,
		Delete: resourceStaticIPAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"static_ip_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStaticIPAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn()

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Creating Lightsail Static IP Attachment: %q", staticIpName)
	_, err := conn.AttachStaticIp(&lightsail.AttachStaticIpInput{
		StaticIpName: aws.String(staticIpName),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})
	if err != nil {
		return fmt.Errorf("creating Lightsail Static IP Attachment: %w", err)
	}

	d.SetId(staticIpName)

	return resourceStaticIPAttachmentRead(d, meta)
}

func resourceStaticIPAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn()

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Reading Lightsail Static IP Attachment: %q", staticIpName)
	out, err := conn.GetStaticIp(&lightsail.GetStaticIpInput{
		StaticIpName: aws.String(staticIpName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			log.Printf("[WARN] Lightsail Static IP Attachment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Lightsail Static IP Attachment (%s):%w", d.Id(), err)
	}
	if !*out.StaticIp.IsAttached {
		log.Printf("[WARN] Lightsail Static IP Attachment (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("instance_name", out.StaticIp.AttachedTo)
	d.Set("ip_address", out.StaticIp.IpAddress)

	return nil
}

func resourceStaticIPAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn()

	name := d.Get("static_ip_name").(string)
	_, err := conn.DetachStaticIp(&lightsail.DetachStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return fmt.Errorf("deleting Lightsail Static IP Attachment (%s):%w", d.Id(), err)
	}
	return nil
}
