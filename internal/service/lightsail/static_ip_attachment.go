package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceStaticIPAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStaticIPAttachmentCreate,
		ReadWithoutTimeout:   resourceStaticIPAttachmentRead,
		DeleteWithoutTimeout: resourceStaticIPAttachmentDelete,

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

func resourceStaticIPAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Creating Lightsail Static IP Attachment: %q", staticIpName)
	_, err := conn.AttachStaticIpWithContext(ctx, &lightsail.AttachStaticIpInput{
		StaticIpName: aws.String(staticIpName),
		InstanceName: aws.String(d.Get("instance_name").(string)),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Static IP Attachment: %s", err)
	}

	d.SetId(staticIpName)

	return append(diags, resourceStaticIPAttachmentRead(ctx, d, meta)...)
}

func resourceStaticIPAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Reading Lightsail Static IP Attachment: %q", staticIpName)
	out, err := conn.GetStaticIpWithContext(ctx, &lightsail.GetStaticIpInput{
		StaticIpName: aws.String(staticIpName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			log.Printf("[WARN] Lightsail Static IP Attachment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Static IP Attachment (%s):%s", d.Id(), err)
	}
	if !*out.StaticIp.IsAttached {
		log.Printf("[WARN] Lightsail Static IP Attachment (%s) is not attached, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("instance_name", out.StaticIp.AttachedTo)
	d.Set("ip_address", out.StaticIp.IpAddress)

	return diags
}

func resourceStaticIPAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

	name := d.Get("static_ip_name").(string)
	_, err := conn.DetachStaticIpWithContext(ctx, &lightsail.DetachStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Static IP Attachment (%s):%s", d.Id(), err)
	}
	return diags
}
