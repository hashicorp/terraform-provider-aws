// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockcrud

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockcrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockcrud/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @SDKResource("aws_mockcrud_foo", name="Foo")
func resourceFoo() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFooCreate,
		ReadWithoutTimeout:   resourceFooRead,
		UpdateWithoutTimeout: resourceFooUpdate,
		DeleteWithoutTimeout: resourceFooDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFooCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	name := d.Get("name").(string)
	input := mockcrud.CreateFooInput{
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(name),
	}

	output, err := conn.CreateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Foo (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FooId))

	if _, err := waitFooCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Foo (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	foo, err := findFooByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Foo (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", d.Id(), err)
	}

	d.Set("description", foo.Description)
	d.Set("name", foo.Name)
	d.Set("region", conn.Options().Region)

	return diags
}

func resourceFooUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	input := mockcrud.UpdateFooInput{
		Description: aws.String(d.Get("description").(string)),
		FooId:       aws.String(d.Id()),
	}

	_, err := conn.UpdateFoo(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Foo (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFooRead(ctx, d, meta)...)
}

func resourceFooDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	if err := deleteFooParts(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s) parts: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Foo: %s", d.Id())
	input := mockcrud.DeleteFooInput{
		FooId: aws.String(d.Id()),
	}
	_, err := conn.DeleteFoo(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Foo (%s): %s", d.Id(), err)
	}

	if _, err := waitFooDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Foo (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func deleteFooParts(ctx context.Context, conn *mockcrud.Client, fooID string) error {
	input := mockcrud.ListFooPartsInput{
		FooId: aws.String(fooID),
	}

	pages := mockcrud.NewListFooPartsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return err
		}

		for _, part := range page.Parts {
			input := mockcrud.DeleteFooPartInput{
				PartId: part.PartId,
				FooId:  aws.String(fooID),
			}

			if _, err := conn.DeleteFooPart(ctx, &input); err != nil {
				return err
			}
		}
	}

	return nil
}
