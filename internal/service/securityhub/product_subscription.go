// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_product_subscription", name="Product Subscription")
// @IdentityAttribute("product_arn")
// @IdentityAttribute("arn")
// @ImportIDHandler("productSubscriptionImportID")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="v6.42.0")
// @Testing(generator=false)
// Custom setup steps.
// @Testing(identityTest=false)
func resourceProductSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProductSubscriptionCreate,
		ReadWithoutTimeout:   resourceProductSubscriptionRead,
		DeleteWithoutTimeout: resourceProductSubscriptionDelete,

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

func resourceProductSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	productARN := d.Get("product_arn").(string)
	input := securityhub.EnableImportFindingsForProductInput{
		ProductArn: aws.String(productARN),
	}

	output, err := conn.EnableImportFindingsForProduct(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Product Subscription (%s): %s", productARN, err)
	}

	d.SetId(productSubscriptionCreateResourceID(productARN, aws.ToString(output.ProductSubscriptionArn)))

	return append(diags, resourceProductSubscriptionRead(ctx, d, meta)...)
}

func resourceProductSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	productARN, productSubscriptionARN, err := productSubscriptionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findProductSubscriptionByARN(ctx, conn, productSubscriptionARN)

	if !d.IsNewResource() && retry.NotFound(err) {
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

func resourceProductSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	_, productSubscriptionARN, err := productSubscriptionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Security Hub Product Subscription: %s", d.Id())
	input := securityhub.DisableImportFindingsForProductInput{
		ProductSubscriptionArn: aws.String(productSubscriptionARN),
	}
	_, err = conn.DisableImportFindingsForProduct(ctx, &input)

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
				LastError: err,
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

const (
	productSubscriptionResourceIDPartCount = 2
)

func productSubscriptionCreateResourceID(productARN, productSubscriptionARN string) string {
	id, _ := flex.FlattenResourceId([]string{productARN, productSubscriptionARN}, productSubscriptionResourceIDPartCount, false)
	return id
}

func productSubscriptionParseResourceID(id string) (string, string, error) {
	parts, err := flex.ExpandResourceId(id, productSubscriptionResourceIDPartCount, false)
	if err != nil {
		return "", "", err
	}

	return parts[0], parts[1], nil
}

var (
	_ inttypes.SDKv2ImportID = productSubscriptionImportID{}
)

type productSubscriptionImportID struct{}

func (productSubscriptionImportID) Parse(id string) (string, map[string]any, error) {
	productARN, productSubscriptionARN, err := productSubscriptionParseResourceID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrARN: productSubscriptionARN,
		"product_arn": productARN,
	}

	return id, result, nil
}

func (productSubscriptionImportID) Create(d *schema.ResourceData) string {
	return productSubscriptionCreateResourceID(d.Get("product_arn").(string), d.Get(names.AttrARN).(string))
}
