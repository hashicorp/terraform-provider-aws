// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_product_subscription", name="Product Subscription")
func resourceProductSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProductSubscriptionCreate,
		ReadWithoutTimeout:   resourceProductSubscriptionRead,
		DeleteWithoutTimeout: resourceProductSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

const (
	productSubscriptionResourceIDPartCount = 2
)

func resourceProductSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	productARN := d.Get("product_arn").(string)
	input := &securityhub.EnableImportFindingsForProductInput{
		ProductArn: aws.String(productARN),
	}

	output, err := conn.EnableImportFindingsForProduct(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Product Subscription (%s): %s", productARN, err)
	}

	d.SetId(errs.Must(flex.FlattenResourceId([]string{productARN, aws.ToString(output.ProductSubscriptionArn)}, productSubscriptionResourceIDPartCount, false)))

	return append(diags, resourceProductSubscriptionRead(ctx, d, meta)...)
}

func resourceProductSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), productSubscriptionResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	productARN, productSubscriptionARN := parts[0], parts[1]

	_, err = findProductSubscriptionByARN(ctx, conn, productSubscriptionARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Product Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, productSubscriptionARN)
	d.Set("product_arn", productARN)

	return diags
}

func resourceProductSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	parts, err := flex.ExpandResourceId(d.Id(), productSubscriptionResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	productSubscriptionARN := parts[1]

	log.Printf("[DEBUG] Deleting Security Hub Product Subscription: %s", d.Id())
	_, err = conn.DisableImportFindingsForProduct(ctx, &securityhub.DisableImportFindingsForProductInput{
		ProductSubscriptionArn: aws.String(productSubscriptionARN),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func findProductSubscriptionByARN(ctx context.Context, conn *securityhub.Client, productSubscriptionARN string) (*string, error) {
	input := &securityhub.ListEnabledProductsForImportInput{}

	return findProductSubscription(ctx, conn, input, func(v string) bool {
		return v == productSubscriptionARN
	})
}

func findProductSubscription(ctx context.Context, conn *securityhub.Client, input *securityhub.ListEnabledProductsForImportInput, filter tfslices.Predicate[string]) (*string, error) {
	output, err := findProductSubscriptions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findProductSubscriptions(ctx context.Context, conn *securityhub.Client, input *securityhub.ListEnabledProductsForImportInput, filter tfslices.Predicate[string]) ([]string, error) {
	var output []string

	pages := securityhub.NewListEnabledProductsForImportPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ProductSubscriptions {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
