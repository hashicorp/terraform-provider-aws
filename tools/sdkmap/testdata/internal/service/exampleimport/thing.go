// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampleimport

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampleimport"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampleimport/types"
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
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"parent_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	name := d.Get("name").(string)
	input := exampleimport.CreateThingInput{
		Name:     aws.String(name),
		ParentId: aws.String(d.Get("parent_id").(string)),
	}

	output, err := conn.CreateThing(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Thing (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ThingId))

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	thing, err := findThingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("name", thing.Name)
	d.Set("parent_id", thing.ParentId)

	return diags
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := exampleimport.DeleteThingInput{
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

func findThingByID(ctx context.Context, conn *exampleimport.Client, id string) (*awstypes.Thing, error) {
	input := exampleimport.DescribeThingInput{
		ThingId: aws.String(id),
	}

	return findThing(ctx, conn, &input)
}

func findThingByName(ctx context.Context, conn *exampleimport.Client, name string) (*awstypes.Thing, error) {
	input := exampleimport.DescribeThingInput{
		Name: aws.String(name),
	}

	return findThing(ctx, conn, &input)
}

func findThing(ctx context.Context, conn *exampleimport.Client, input *exampleimport.DescribeThingInput) (*awstypes.Thing, error) {
	output, err := conn.DescribeThing(ctx, input)

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
