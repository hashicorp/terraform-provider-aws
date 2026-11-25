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
	tfmockcrossservice "github.com/hashicorp/terraform-provider-aws/internal/service/mockcrossservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	barDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_mockcrud_bar", name="Bar")
func resourceBar() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBarCreate,
		ReadWithoutTimeout:   resourceBarRead,
		UpdateWithoutTimeout: resourceBarUpdate,
		DeleteWithoutTimeout: resourceBarDelete,

		Schema: map[string]*schema.Schema{
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"cross_service_foo_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"foo_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBarCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.MockCRUDClient(ctx)

	fooID := d.Get("foo_id").(string)
	if _, err := findFooByID(ctx, conn, fooID); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", fooID, err)
	}

	if v, ok := d.GetOk("cross_service_foo_id"); ok {
		if _, err := tfmockcrossservice.FindFooByID(ctx, c.MockCrossServiceClient(ctx), v.(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Foo (%s): %s", v, err)
		}
	}

	input := mockcrud.CreateBarInput{
		Size:  aws.Int32(int32(d.Get("size").(int))),
		FooId: aws.String(fooID),
	}

	output, err := conn.CreateBar(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Bar: %s", err)
	}

	d.SetId(aws.ToString(output.BarId))

	const (
		inARow = 2
	)
	if _, err := retry.Op(func(ctx context.Context) (*awstypes.Bar, error) {
		return findBarByID(ctx, conn, d.Id())
	}).UntilFoundN(inARow)(ctx, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Bar (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBarRead(ctx, d, meta)...)
}

func resourceBarRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	bar, err := findBarByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Bar (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Bar (%s): %s", d.Id(), err)
	}

	d.Set("size", bar.Size)
	d.Set("foo_id", bar.FooId)

	return diags
}

func resourceBarUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	input := mockcrud.UpdateBarInput{
		BarId: aws.String(d.Id()),
		Size:  aws.Int32(int32(d.Get("size").(int))),
	}

	_, err := conn.UpdateBar(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Bar (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBarRead(ctx, d, meta)...)
}

func resourceBarDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	log.Printf("[DEBUG] Deleting Bar: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceInUseException](ctx, barDeleteTimeout, func(ctx context.Context) (any, error) {
		input := mockcrud.DeleteBarInput{
			BarId: aws.String(d.Id()),
		}

		return conn.DeleteBar(ctx, &input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Bar (%s): %s", d.Id(), err)
	}

	return diags
}
