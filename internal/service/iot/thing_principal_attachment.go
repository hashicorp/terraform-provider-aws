// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing_principal_attachment")
func ResourceThingPrincipalAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingPrincipalAttachmentCreate,
		ReadWithoutTimeout:   resourceThingPrincipalAttachmentRead,
		DeleteWithoutTimeout: resourceThingPrincipalAttachmentDelete,

		Schema: map[string]*schema.Schema{
			names.AttrPrincipal: {
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
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	principal := d.Get(names.AttrPrincipal).(string)
	thing := d.Get("thing").(string)

	_, err := conn.AttachThingPrincipal(ctx, &iot.AttachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching principal %s to thing %s: %s", principal, thing, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", thing, principal))
	return append(diags, resourceThingPrincipalAttachmentRead(ctx, d, meta)...)
}

func GetThingPricipalAttachment(ctx context.Context, conn *iot.Client, thing, principal string) (bool, error) {
	out, err := conn.ListThingPrincipals(ctx, &iot.ListThingPrincipalsInput{
		ThingName: aws.String(thing),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	found := false
	for _, name := range out.Principals {
		if principal == name {
			found = true
			break
		}
	}
	return found, nil
}

func resourceThingPrincipalAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	principal := d.Get(names.AttrPrincipal).(string)
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
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	principal := d.Get(names.AttrPrincipal).(string)
	thing := d.Get("thing").(string)

	_, err := conn.DetachThingPrincipal(ctx, &iot.DetachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] IoT Principal %s or Thing %s not found, removing from state", principal, thing)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "detaching principal %s from thing %s: %s", principal, thing, err)
	}

	return diags
}
