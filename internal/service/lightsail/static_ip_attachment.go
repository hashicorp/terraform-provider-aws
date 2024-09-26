// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_static_ip_attachment")
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
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStaticIPAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Creating Lightsail Static IP Attachment: %q", staticIpName)
	_, err := conn.AttachStaticIp(ctx, &lightsail.AttachStaticIpInput{
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
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	staticIpName := d.Get("static_ip_name").(string)
	log.Printf("[INFO] Reading Lightsail Static IP Attachment: %q", staticIpName)
	out, err := conn.GetStaticIp(ctx, &lightsail.GetStaticIpInput{
		StaticIpName: aws.String(staticIpName),
	})
	if err != nil {
		if IsANotFoundError(err) {
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
	d.Set(names.AttrIPAddress, out.StaticIp.IpAddress)

	return diags
}

func resourceStaticIPAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	name := d.Get("static_ip_name").(string)
	_, err := conn.DetachStaticIp(ctx, &lightsail.DetachStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Static IP Attachment (%s):%s", d.Id(), err)
	}
	return diags
}
