// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockfuncrefs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockfuncrefs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockfuncrefs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_mockfuncrefs_method_value", name="Method Value")
func resourceMethodValue() *schema.Resource {
	h := &attributeHandler{
		attributeName: awstypes.FooAttributeNamePolicy,
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
			"foo_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

type attributeHandler struct {
	attributeName awstypes.FooAttributeName
	schemaKey     string
}

func (h *attributeHandler) Upsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockFuncRefsClient(ctx)

	fooID := d.Get("foo_id").(string)
	input := mockfuncrefs.SetFooAttributeInput{
		AttributeName:  h.attributeName,
		AttributeValue: aws.String(d.Get(h.schemaKey).(string)),
		FooId:          aws.String(fooID),
	}

	_, err := conn.SetFooAttribute(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Foo (%s) attribute (%s): %s", fooID, h.attributeName, err)
	}

	d.SetId(fooID)

	return append(diags, h.Read(ctx, d, meta)...)
}

func (h *attributeHandler) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockFuncRefsClient(ctx)

	value, err := findFooAttributeByTwoPartKey(ctx, conn, d.Id(), h.attributeName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) attribute (%s) not found, removing from state", d.Id(), h.attributeName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s) attribute (%s): %s", d.Id(), h.attributeName, err)
	}

	d.Set(h.schemaKey, value)
	d.Set("foo_id", d.Id())

	return diags
}

func (h *attributeHandler) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockFuncRefsClient(ctx)

	log.Printf("[DEBUG] Deleting Foo (%s) attribute (%s)", d.Id(), h.attributeName)
	input := mockfuncrefs.DeleteFooAttributeInput{
		AttributeName: h.attributeName,
		FooId:         aws.String(d.Id()),
	}
	_, err := conn.DeleteFooAttribute(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s) attribute (%s): %s", d.Id(), h.attributeName, err)
	}

	return diags
}

func findFooAttributeByTwoPartKey(ctx context.Context, conn *mockfuncrefs.Client, fooID string, attributeName awstypes.FooAttributeName) (string, error) {
	input := mockfuncrefs.GetFooAttributeInput{
		AttributeName: attributeName,
		FooId:         aws.String(fooID),
	}

	output, err := conn.GetFooAttribute(ctx, &input)

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
