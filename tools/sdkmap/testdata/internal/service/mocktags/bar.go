// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mocktags

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mocktags"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mocktags/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_mocktags_bar", name="Bar")
// @Tags(identifierAttribute="arn", resourceType="Bar")
func resourceBar() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBarCreate,
		ReadWithoutTimeout:   resourceBarRead,
		UpdateWithoutTimeout: resourceBarUpdate,
		DeleteWithoutTimeout: resourceBarDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceBarCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockTagsClient(ctx)

	input := mocktags.CreateBarInput{
		Tags: getTagsIn(ctx),
	}

	output, err := conn.CreateBar(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Bar: %s", err)
	}

	d.SetId(aws.ToString(output.BarId))

	return append(diags, resourceBarRead(ctx, d, meta)...)
}

func resourceBarRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockTagsClient(ctx)

	bar, err := findBarByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Bar (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Bar (%s): %s", d.Id(), err)
	}

	d.Set("arn", bar.BarArn)

	return diags
}

func resourceBarUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceBarRead(ctx, d, meta)
}

func resourceBarDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MockTagsClient(ctx)

	log.Printf("[DEBUG] Deleting Bar: %s", d.Id())
	input := mocktags.DeleteBarInput{
		BarId: aws.String(d.Id()),
	}
	_, err := conn.DeleteBar(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Bar (%s): %s", d.Id(), err)
	}

	return diags
}

func findBarByID(ctx context.Context, conn *mocktags.Client, id string) (*awstypes.Bar, error) {
	input := mocktags.DescribeBarInput{
		BarId: aws.String(id),
	}

	output, err := conn.DescribeBar(ctx, &input)

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
