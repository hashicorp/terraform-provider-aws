// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplefuncrefs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplefuncrefs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplefuncrefs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_examplefuncrefs_method_value", name="Method Value")
func resourceMethodValue() *schema.Resource {
	h := &attributeHandler{
		attributeName: awstypes.ThingAttributeNamePolicy,
		schemaKey:     "policy",
	}

	return &schema.Resource{
		CreateWithoutTimeout: h.Upsert,
		ReadWithoutTimeout:   h.Read,
		UpdateWithoutTimeout: h.Upsert,
		DeleteWithoutTimeout: h.Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"thing_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type attributeHandler struct {
	attributeName awstypes.ThingAttributeName
	schemaKey     string
}

func (h *attributeHandler) Upsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	thingID := d.Get("thing_id").(string)
	input := examplefuncrefs.SetThingAttributeInput{
		AttributeName:  h.attributeName,
		AttributeValue: aws.String(d.Get(h.schemaKey).(string)),
		ThingId:        aws.String(thingID),
	}

	_, err := conn.SetThingAttribute(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Thing (%s) attribute (%s): %s", thingID, h.attributeName, err)
	}

	d.SetId(thingID)

	return append(diags, h.Read(ctx, d, meta)...)
}

func (h *attributeHandler) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	value, err := findThingAttributeByTwoPartKey(ctx, conn, d.Id(), h.attributeName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) attribute (%s) not found, removing from state", d.Id(), h.attributeName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s) attribute (%s): %s", d.Id(), h.attributeName, err)
	}

	d.Set(h.schemaKey, value)
	d.Set("thing_id", d.Id())

	return diags
}

func (h *attributeHandler) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleFuncRefsClient(ctx)

	log.Printf("[DEBUG] Deleting Thing (%s) attribute (%s)", d.Id(), h.attributeName)
	input := examplefuncrefs.DeleteThingAttributeInput{
		AttributeName: h.attributeName,
		ThingId:       aws.String(d.Id()),
	}
	_, err := conn.DeleteThingAttribute(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Thing (%s) attribute (%s): %s", d.Id(), h.attributeName, err)
	}

	return diags
}

func findThingAttributeByTwoPartKey(ctx context.Context, conn *examplefuncrefs.Client, thingID string, attributeName awstypes.ThingAttributeName) (string, error) {
	input := examplefuncrefs.GetThingAttributeInput{
		AttributeName: attributeName,
		ThingId:       aws.String(thingID),
	}

	output, err := conn.GetThingAttribute(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return "", &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.AttributeValue == nil {
		return "", tfresource.NewEmptyResultError()
	}

	return aws.ToString(output.AttributeValue), nil
}
