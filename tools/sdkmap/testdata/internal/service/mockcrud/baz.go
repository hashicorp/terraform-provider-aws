// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockcrud

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockcrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockcrud/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_mockcrud_baz", name="Baz")
func resourceBaz() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBazCreate,
		ReadWithoutTimeout:   resourceBazRead,
		DeleteWithoutTimeout: resourceBazDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceBazCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	name := d.Get("name").(string)
	input := mockcrud.CreateBazInput{
		Name: aws.String(name),
	}

	output, err := conn.CreateBaz(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Baz (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.BazId))

	return append(diags, resourceBazRead(ctx, d, meta)...)
}

func resourceBazRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	input := mockcrud.GetBazInput{
		BazId: aws.String(d.Id()),
	}

	output, err := conn.GetBaz(ctx, &input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Baz (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Baz (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)

	return diags
}

func resourceBazDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockCRUDClient(ctx)

	log.Printf("[DEBUG] Deleting Baz: %s", d.Id())
	input := mockcrud.DeleteBazInput{
		BazId: aws.String(d.Id()),
	}
	_, err := conn.DeleteBaz(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Baz (%s): %s", d.Id(), err)
	}

	return diags
}
