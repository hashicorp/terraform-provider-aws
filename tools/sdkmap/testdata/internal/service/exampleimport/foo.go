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

func fooSchema() map[string]*schema.Schema {
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

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	name := d.Get("name").(string)
	input := exampleimport.CreateFooInput{
		Name:     aws.String(name),
		ParentId: aws.String(d.Get("parent_id").(string)),
	}

	output, err := conn.CreateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FooId))

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	foo, err := findFooByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("name", foo.Name)
	d.Set("parent_id", foo.ParentId)

	return diags
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleImportClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := exampleimport.DeleteFooInput{
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

func findFooByID(ctx context.Context, conn *exampleimport.Client, id string) (*awstypes.Foo, error) {
	input := exampleimport.DescribeFooInput{
		FooId: aws.String(id),
	}

	return findFoo(ctx, conn, &input)
}

func findFooByName(ctx context.Context, conn *exampleimport.Client, name string) (*awstypes.Foo, error) {
	input := exampleimport.DescribeFooInput{
		Name: aws.String(name),
	}

	return findFoo(ctx, conn, &input)
}

func findFoo(ctx context.Context, conn *exampleimport.Client, input *exampleimport.DescribeFooInput) (*awstypes.Foo, error) {
	output, err := conn.DescribeFoo(ctx, input)

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
