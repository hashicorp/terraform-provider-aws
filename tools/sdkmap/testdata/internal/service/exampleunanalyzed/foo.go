// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampleunanalyzed

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampleunanalyzed"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampleunanalyzed/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_exampleunanalyzed_foo", name="Foo")
func resourceFoo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		DeleteWithoutTimeout: resourceFooDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_foo_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	name := d.Get("name").(string)
	input := exampleunanalyzed.CreateFooInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("source_foo_id"); ok {
		presignInput := exampleunanalyzed.CopyFooInput{
			SourceFooId: aws.String(v.(string)),
		}

		output, err := exampleunanalyzed.NewPresignClient(conn).PresignCopyFoo(ctx, &presignInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "presigning Foo (%s) copy: %s", v, err)
		}

		input.PreSignedUrl = aws.String(output.URL)
	}

	output, err := conn.CreateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FooId))

	waiter := exampleunanalyzed.NewFooReadyWaiter(conn)
	describeInput := exampleunanalyzed.DescribeFooInput{
		FooId: aws.String(d.Id()),
	}

	if err := waiter.Wait(ctx, &describeInput, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Foo (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	foo, err := findFooByName(ctx, conn, d.Get("name").(string))

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("name", foo.Name)

	return diags
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := exampleunanalyzed.DeleteFooInput{
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

func findFooByName(ctx context.Context, conn *exampleunanalyzed.Client, name string) (*awstypes.Foo, error) {
	var input exampleunanalyzed.ListFoosInput

	pages := exampleunanalyzed.NewListFoosPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, foo := range page.Foos {
			if aws.ToString(foo.Name) == name {
				return &foo, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError()
}
