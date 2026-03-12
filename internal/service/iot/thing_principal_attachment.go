// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
			"thing_principal_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ThingPrincipalType](),
			},
		},
	}
}

func resourceThingPrincipalAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	principal := d.Get(names.AttrPrincipal).(string)
	thing := d.Get("thing").(string)
	id := fmt.Sprintf("%s|%s", thing, principal)
	input := &iot.AttachThingPrincipalInput{
		Principal: aws.String(principal),
		ThingName: aws.String(thing),
	}

	if v, ok := d.Get("thing_principal_type").(string); ok {
		input.ThingPrincipalType = awstypes.ThingPrincipalType(v)
	}

	_, err := conn.AttachThingPrincipal(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Thing Principal Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceThingPrincipalAttachmentRead(ctx, d, meta)...)
}

func resourceThingPrincipalAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTClient(ctx)

	id := d.Id()
	parts := strings.Split(id, "|")
	if len(parts) != 2 {
		return sdkdiag.AppendErrorf(diags, "unexpected format for ID (%s), expected thing|principal", id)
	}
	thing := parts[0]
	principal := parts[1]

	out, err := findThingPrincipalAttachmentByTwoPartKey(ctx, conn, thing, principal)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IoT Thing Principal Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Thing Principal Attachment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrPrincipal, out.Principal)
	d.Set("thing", thing)
	d.Set("thing_principal_type", out.ThingPrincipalType)

	return diags
}

func resourceThingPrincipalAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

func findThingPrincipalAttachmentByTwoPartKey(ctx context.Context, conn *iot.Client, thing, principal string) (*awstypes.ThingPrincipalObject, error) {
	input := &iot.ListThingPrincipalsV2Input{
		ThingName: aws.String(thing),
	}

	return findThingPrincipal(ctx, conn, input, func(v string) bool {
		return principal == v
	})
}

func findThingPrincipal(ctx context.Context, conn *iot.Client, input *iot.ListThingPrincipalsV2Input, filter tfslices.Predicate[string]) (*awstypes.ThingPrincipalObject, error) {
	output, err := findThingPrincipals(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findThingPrincipals(ctx context.Context, conn *iot.Client, input *iot.ListThingPrincipalsV2Input, filter tfslices.Predicate[string]) ([]awstypes.ThingPrincipalObject, error) {
	var output []awstypes.ThingPrincipalObject

	pages := iot.NewListThingPrincipalsV2Paginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ThingPrincipalObjects {
			if filter(aws.ToString(v.Principal)) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
