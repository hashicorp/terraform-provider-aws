// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecustomizediff

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecustomizediff"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecustomizediff/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func thingSchema() map[string]*schema.Schema {
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

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx)

	input := examplecustomizediff.CreateThingInput{
		Content:  aws.String(d.Get("content").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Size:     aws.Int32(int32(d.Get("size").(int))),
	}

	output, err := conn.CreateThing(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Thing: %s", err)
	}

	d.SetId(aws.ToString(output.ThingId))

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx)

	thing, err := findThingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("checksum", thing.Checksum)
	d.Set("content", thing.Content)
	d.Set("parent_id", thing.ParentId)
	d.Set("size", thing.Size)

	return diags
}

func resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx)

	input := examplecustomizediff.UpdateThingInput{
		Content: aws.String(d.Get("content").(string)),
		Size:    aws.Int32(int32(d.Get("size").(int))),
		ThingId: aws.String(d.Id()),
	}

	_, err := conn.UpdateThing(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Thing (%s): %s", d.Id(), err)
	}

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCustomizeDiffClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := examplecustomizediff.DeleteThingInput{
		ThingId: aws.String(d.Id()),
	}
	_, err := conn.DeleteThing(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Thing (%s): %s", d.Id(), err)
	}

	return diags
}

func findThingByID(ctx context.Context, conn *examplecustomizediff.Client, id string) (*awstypes.Thing, error) {
	input := examplecustomizediff.DescribeThingInput{
		ThingId: aws.String(id),
	}

	output, err := conn.DescribeThing(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Thing == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Thing, nil
}

func findParentByID(ctx context.Context, conn *examplecustomizediff.Client, id string) (*awstypes.Parent, error) {
	input := examplecustomizediff.DescribeParentInput{
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

func findQuotaByName(ctx context.Context, conn *examplecustomizediff.Client, name string) (*awstypes.Quota, error) {
	input := examplecustomizediff.GetQuotaInput{
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
