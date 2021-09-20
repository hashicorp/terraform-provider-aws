package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
	conn := meta.(*conns.AWSClient).LightsailConn

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Attaching Lightsail Static IP: %q", staticIpName)
	out, err := conn.AttachStaticIp(&lightsail.AttachStaticIpInput{
		StaticIpName: aws.String(staticIpName),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Lightsail Static IP attached: %s", *out)

	d.SetId(staticIpName)

	return resourceStaticIPAttachmentRead(d, meta)
}

func resourceStaticIPAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Reading Lightsail Static IP: %q", staticIpName)
	out, err := conn.GetStaticIp(&lightsail.GetStaticIpInput{
		StaticIpName: aws.String(staticIpName),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail Static IP (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}
	if !*out.StaticIp.IsAttached {
		log.Printf("[WARN] Lightsail Static IP (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	log.Printf("[INFO] Received Lightsail Static IP: %s", *out)

	d.Set("instance_name", out.StaticIp.AttachedTo)
	d.Set("ip_address", out.StaticIp.IpAddress)

	return nil
}

func resourceStaticIPAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	name := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Detaching Lightsail Static IP: %q", name)
	out, err := conn.DetachStaticIp(&lightsail.DetachStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Detached Lightsail Static IP: %s", *out)
	return nil
}
