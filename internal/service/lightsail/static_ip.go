// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_static_ip")
func ResourceStaticIP() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStaticIPCreate,
		ReadWithoutTimeout:   resourceStaticIPRead,
		DeleteWithoutTimeout: resourceStaticIPDelete,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStaticIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	name := d.Get(names.AttrName).(string)
	log.Printf("[INFO] Allocating Lightsail Static IP: %q", name)
	_, err := conn.AllocateStaticIp(ctx, &lightsail.AllocateStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Static IP: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceStaticIPRead(ctx, d, meta)...)
}

func resourceStaticIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	name := d.Get(names.AttrName).(string)
	log.Printf("[INFO] Reading Lightsail Static IP: %q", name)
	out, err := conn.GetStaticIp(ctx, &lightsail.GetStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		if IsANotFoundError(err) {
			log.Printf("[WARN] Lightsail Static IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Static IP (%s):%s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.StaticIp.Arn)
	d.Set(names.AttrIPAddress, out.StaticIp.IpAddress)
	d.Set("support_code", out.StaticIp.SupportCode)

	return diags
}

func resourceStaticIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	name := d.Get(names.AttrName).(string)
	log.Printf("[INFO] Deleting Lightsail Static IP: %q", name)
	_, err := conn.ReleaseStaticIp(ctx, &lightsail.ReleaseStaticIpInput{
		StaticIpName: aws.String(name),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Static IP (%s):%s", d.Id(), err)
	}
	return diags
}
