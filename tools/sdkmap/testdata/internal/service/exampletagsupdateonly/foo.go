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

// @SDKResource("aws_exampletagsupdateonly_foo", name="Foo")
// @Tags(identifierAttribute="arn")
func resourceFoo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		UpdateWithoutTimeout: resourceFooUpdate,
		DeleteWithoutTimeout: resourceFooDelete,

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

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	name := d.Get("name").(string)
	input := exampletagsupdateonly.CreateFooInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if _, err := conn.CreateFoo(ctx, &input); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	input := exampletagsupdateonly.GetFooInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.GetFoo(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Foo.Arn)
	d.Set("name", output.Foo.Name)

	setTagsOut(ctx, output.Foo.Tags)

	return diags
}

func resourceFooUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceFooRead(ctx, d, meta)
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleTagsUpdateOnlyClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := exampletagsupdateonly.DeleteFooInput{
		Name: aws.String(d.Id()),
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
