// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrossservice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplecrossservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrossservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_examplecrossservice_thing", name="Thing")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCrossServiceClient(ctx)

	name := d.Get("name").(string)
	input := examplecrossservice.CreateThingInput{
		Name: aws.String(name),
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
	conn := meta.(*conns.AWSClient).ExampleCrossServiceClient(ctx)

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

	return diags
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleCrossServiceClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := examplecrossservice.DeleteThingInput{
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

func findThingByID(ctx context.Context, conn *examplecrossservice.Client, id string) (*awstypes.Thing, error) {
	input := examplecrossservice.DescribeThingInput{
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
