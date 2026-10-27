// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockcustomizediff

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockcustomizediff"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockcustomizediff/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func fooSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"checksum": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"content": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"parent_id": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"size": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}
}

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCustomizeDiffClient(ctx)

	input := mockcustomizediff.CreateFooInput{
		Content:  aws.String(d.Get("content").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Size:     aws.Int32(int32(d.Get("size").(int))),
	}

	output, err := conn.CreateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo: %s", err)
	}

	d.SetId(aws.ToString(output.FooId))

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCustomizeDiffClient(ctx)

	foo, err := findFooByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("checksum", foo.Checksum)
	d.Set("content", foo.Content)
	d.Set("parent_id", foo.ParentId)
	d.Set("size", foo.Size)

	return diags
}

func resourceFooUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCustomizeDiffClient(ctx)

	input := mockcustomizediff.UpdateFooInput{
		Content: aws.String(d.Get("content").(string)),
		Size:    aws.Int32(int32(d.Get("size").(int))),
		FooId:   aws.String(d.Id()),
	}

	_, err := conn.UpdateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Foo (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCustomizeDiffClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := mockcustomizediff.DeleteFooInput{
		FooId: aws.String(d.Id()),
	}
	_, err := conn.DeleteFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s): %s", d.Id(), err)
	}

	return diags
}

func findFooByID(ctx context.Context, conn *mockcustomizediff.Client, id string) (*awstypes.Foo, error) {
	input := mockcustomizediff.DescribeFooInput{
		FooId: aws.String(id),
	}

	output, err := conn.DescribeFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Foo == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Foo, nil
}

func findParentByID(ctx context.Context, conn *mockcustomizediff.Client, id string) (*awstypes.Parent, error) {
	input := mockcustomizediff.DescribeParentInput{
		ParentId: aws.String(id),
	}

	output, err := conn.DescribeParent(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Parent == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Parent, nil
}

func findQuotaByName(ctx context.Context, conn *mockcustomizediff.Client, name string) (*awstypes.Quota, error) {
	input := mockcustomizediff.GetQuotaInput{
		QuotaName: aws.String(name),
	}

	output, err := conn.GetQuota(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Quota, nil
}
