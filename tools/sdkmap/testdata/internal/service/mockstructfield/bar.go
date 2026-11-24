// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockstructfield

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockstructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockstructfield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_mockstructfield_bar", name="Bar")
func resourceBar() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBarCreate,
		ReadWithoutTimeout:   resourceBarRead,
		UpdateWithoutTimeout: resourceBarUpdate,
		DeleteWithoutTimeout: resourceBarDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"predicate": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceBarCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	name := d.Get("name").(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := mockstructfield.CreateBarInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateBar(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Bar (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*mockstructfield.CreateBarOutput).Bar.BarId))

	return append(diags, resourceBarRead(ctx, d, meta)...)
}

func resourceBarRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	bar, err := findBarByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Bar (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Bar (%s): %s", d.Id(), err)
	}

	d.Set("name", bar.Name)
	d.Set("predicate", bar.Predicate)

	return diags
}

func resourceBarUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := mockstructfield.UpdateBarInput{
			ChangeToken: token,
			Predicate:   aws.String(d.Get("predicate").(string)),
			BarId:       aws.String(d.Id()),
		}

		return conn.UpdateBar(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Bar (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBarRead(ctx, d, meta)...)
}

func resourceBarDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockStructFieldClient(ctx)

	log.Printf("[INFO] Deleting Bar: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := mockstructfield.DeleteBarInput{
			ChangeToken: token,
			BarId:       aws.String(d.Id()),
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

func findBarByID(ctx context.Context, conn *mockstructfield.Client, id string) (*awstypes.Bar, error) {
	input := mockstructfield.GetBarInput{
		BarId: aws.String(id),
	}

	output, err := conn.GetBar(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Bar == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Bar, nil
}
