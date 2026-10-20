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

// @SDKResource("aws_exampleunanalyzed_thing", name="Thing")
func resourceThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceThingCreate,
		ReadWithoutTimeout:   resourceThingRead,
		DeleteWithoutTimeout: resourceThingDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_thing_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	name := d.Get("name").(string)
	input := exampleunanalyzed.CreateThingInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("source_thing_id"); ok {
		presignInput := exampleunanalyzed.CopyThingInput{
			SourceThingId: aws.String(v.(string)),
		}

		output, err := exampleunanalyzed.NewPresignClient(conn).PresignCopyThing(ctx, &presignInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "presigning Thing (%s) copy: %s", v, err)
		}

		input.PreSignedUrl = aws.String(output.URL)
	}

	output, err := conn.CreateThing(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Thing (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.ThingId))

	waiter := exampleunanalyzed.NewThingReadyWaiter(conn)
	describeInput := exampleunanalyzed.DescribeThingInput{
		ThingId: aws.String(d.Id()),
	}

	if err := waiter.Wait(ctx, &describeInput, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Thing (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceThingRead(ctx, d, meta)...)
}

func resourceThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	thing, err := findThingByName(ctx, conn, d.Get("name").(string))

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
	conn := meta.(*conns.AWSClient).ExampleUnanalyzedClient(ctx)

	log.Printf("[DEBUG] Deleting Thing: %s", d.Id())
	input := exampleunanalyzed.DeleteThingInput{
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

func findThingByName(ctx context.Context, conn *exampleunanalyzed.Client, name string) (*awstypes.Thing, error) {
	var input exampleunanalyzed.ListThingsInput

	pages := exampleunanalyzed.NewListThingsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, thing := range page.Things {
			if aws.ToString(thing.Name) == name {
				return &thing, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError()
}
