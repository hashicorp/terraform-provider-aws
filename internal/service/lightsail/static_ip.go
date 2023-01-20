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

func ResourceStaticIP() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStaticIPCreate,
		ReadWithoutTimeout:   resourceStaticIPRead,
		DeleteWithoutTimeout: resourceStaticIPDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
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
	conn := meta.(*conns.AWSClient).LightsailConn()

	name := d.Get("name").(string)
	log.Printf("[INFO] Allocating Lightsail Static IP: %q", name)
	_, err := conn.AllocateStaticIpWithContext(ctx, &lightsail.AllocateStaticIpInput{
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
	conn := meta.(*conns.AWSClient).LightsailConn()

	name := d.Get("name").(string)
	log.Printf("[INFO] Reading Lightsail Static IP: %q", name)
	out, err := conn.GetStaticIpWithContext(ctx, &lightsail.GetStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
			log.Printf("[WARN] Lightsail Static IP (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Static IP (%s):%s", d.Id(), err)
	}

	d.Set("arn", out.StaticIp.Arn)
	d.Set("ip_address", out.StaticIp.IpAddress)
	d.Set("support_code", out.StaticIp.SupportCode)

	return diags
}

func resourceStaticIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LightsailConn()

	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting Lightsail Static IP: %q", name)
	_, err := conn.ReleaseStaticIpWithContext(ctx, &lightsail.ReleaseStaticIpInput{
		StaticIpName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Static IP (%s):%s", d.Id(), err)
	}
	return diags
}
