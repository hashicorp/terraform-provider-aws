// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampletagsupdateonly

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampletagsupdateonly"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampletagsupdateonly/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKResource("aws_exampletagsupdateonly_thing", name="Thing")
// @Tags(identifierAttribute="arn")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		UpdateWithoutTimeout: resourceThingUpdate,
		DeleteWithoutTimeout: resourceThingDelete,

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
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	name := d.Get("name").(string)
	input := exampletagsupdateonly.CreateThingInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if _, err := conn.CreateThing(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Thing (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	input := exampletagsupdateonly.GetThingInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.GetThing(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Thing (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Thing (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Thing.Arn)
	d.Set("name", output.Thing.Name)

	setTagsOut(ctx, output.Thing.Tags)

	return diags
}

func resourceThingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceThingRead(ctx, d, meta)
}

func resourceThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := exampletagsupdateonly.DeleteThingInput{
		Name: aws.String(d.Id()),
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
