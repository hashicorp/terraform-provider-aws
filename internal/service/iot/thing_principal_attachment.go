package iot

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceThingPrincipalAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingPrincipalAttachmentCreate,
		ReadWithoutTimeout:   resourceThingPrincipalAttachmentRead,
		DeleteWithoutTimeout: resourceThingPrincipalAttachmentDelete,

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

func resourceThingPrincipalAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	_, err := conn.AttachThingPrincipalWithContext(ctx, &iot.AttachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching principal %s to thing %s: %s", principal, thing, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", thing, principal))
	return append(diags, resourceThingPrincipalAttachmentRead(ctx, d, meta)...)
}

func GetThingPricipalAttachment(ctx context.Context, conn *iot.IoT, thing, principal string) (bool, error) {
	out, err := conn.ListThingPrincipalsWithContext(ctx, &iot.ListThingPrincipalsInput{
		ThingName: aws.String(thing),
	})
	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
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

func resourceThingPrincipalAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	found, err := GetThingPricipalAttachment(ctx, conn, thing, principal)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing principals for thing %s: %s", thing, err)
	}

	if !found {
		log.Printf("[WARN] IoT Thing Principal Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	return diags
}

func resourceThingPrincipalAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	principal := d.Get("principal").(string)
	thing := d.Get("thing").(string)

	_, err := conn.DetachThingPrincipalWithContext(ctx, &iot.DetachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] IoT Principal %s or Thing %s not found, removing from state", principal, thing)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "detaching principal %s from thing %s: %s", principal, thing, err)
	}

	return diags
}
