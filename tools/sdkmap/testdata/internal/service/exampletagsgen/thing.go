// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampletagsgen

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampletagsgen"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampletagsgen/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_exampletagsgen_thing", name="Thing")
// @Tags(identifierAttribute="arn")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsGenClient(ctx)

	name := d.Get("name").(string)
	input := exampletagsgen.CreateThingInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
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
	conn := meta.(*conns.AWSClient).ExampleTagsGenClient(ctx)

	thing, err := findThingByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("arn", thing.ThingArn)
	d.Set("name", thing.Name)

	return diags
}

func resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceThingRead(ctx, d, meta)
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsGenClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := exampletagsgen.DeleteThingInput{
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

func findThingByID(ctx context.Context, conn *exampletagsgen.Client, id string) (*awstypes.Thing, error) {
	input := exampletagsgen.DescribeThingInput{
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
