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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iot_thing_principal_attachment", name="Thing Principal Attachment")
func resourceThingPrincipalAttachment() *schema.Resource {
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
	id := fmt.Sprintf("%s|%s", thing, principal)
	input := &iot.AttachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	}

	_, err := conn.AttachThingPrincipal(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Principal Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceThingPrincipalAttachmentRead(ctx, d, meta)...)
}

func resourceThingPrincipalAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	principal := d.Get(names.AttrPrincipal).(string)
	thing := d.Get("thing").(string)

	_, err := findThingPrincipalAttachmentByTwoPartKey(ctx, conn, thing, principal)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IoT Thing Principal Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Principal Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceThingPrincipalAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	log.Printf("[DEBUG] Deleting IoT Thing Principal Attachment: %s", d.Id())
	_, err := conn.DetachThingPrincipal(ctx, &iot.DetachThingPrincipalInput{
		Principal: aws.String(d.Get(names.AttrPrincipal).(string)),
		ThingName: aws.String(d.Get("thing").(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Thing Principal Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func findThingPrincipalAttachmentByTwoPartKey(ctx context.Context, conn *iot.Client, thing, principal string) (*string, error) {
	input := &iot.ListThingPrincipalsInput{
		ThingName: aws.String(thing),
	}

	return findThingPrincipal(ctx, conn, input, func(v string) bool {
		return principal == v
	})
}

func findThingPrincipal(ctx context.Context, conn *iot.Client, input *iot.ListThingPrincipalsInput, filter tfslices.Predicate[string]) (*string, error) {
	output, err := findThingPrincipals(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findThingPrincipals(ctx context.Context, conn *iot.Client, input *iot.ListThingPrincipalsInput, filter tfslices.Predicate[string]) ([]string, error) {
	var output []string

	pages := iot.NewListThingPrincipalsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Principals {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
