package securityhub

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProductSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProductSubscriptionCreate,
		ReadWithoutTimeout:   resourceProductSubscriptionRead,
		DeleteWithoutTimeout: resourceProductSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"product_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceProductSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()
	productArn := d.Get("product_arn").(string)

	log.Printf("[DEBUG] Enabling Security Hub Product Subscription for product %s", productArn)

	resp, err := conn.EnableImportFindingsForProductWithContext(ctx, &securityhub.EnableImportFindingsForProductInput{
		ProductArn: aws.String(productArn),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Product Subscription for product %s: %s", productArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", productArn, *resp.ProductSubscriptionArn))

	return append(diags, resourceProductSubscriptionRead(ctx, d, meta)...)
}

func resourceProductSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	productArn, productSubscriptionArn, err := ProductSubscriptionParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Reading Security Hub Product Subscriptions to find %s", d.Id())

	exists, err := ProductSubscriptionCheckExists(ctx, conn, productSubscriptionArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	if !exists {
		log.Printf("[WARN] Security Hub Product Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("product_arn", productArn)
	d.Set("arn", productSubscriptionArn)

	return diags
}

func ProductSubscriptionCheckExists(ctx context.Context, conn *securityhub.SecurityHub, productSubscriptionArn string) (bool, error) {
	input := &securityhub.ListEnabledProductsForImportInput{}
	exists := false

	err := conn.ListEnabledProductsForImportPagesWithContext(ctx, input, func(page *securityhub.ListEnabledProductsForImportOutput, lastPage bool) bool {
		for _, readProductSubscriptionArn := range page.ProductSubscriptions {
			if aws.StringValue(readProductSubscriptionArn) == productSubscriptionArn {
				exists = true
				return false
			}
		}
		return !lastPage
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

func ProductSubscriptionParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected Security Hub Product Subscription ID in format <product_arn>,<arn> - received: %s", id)
	}

	return parts[0], parts[1], nil
}

func resourceProductSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()
	log.Printf("[DEBUG] Disabling Security Hub Product Subscription %s", d.Id())

	_, productSubscriptionArn, err := ProductSubscriptionParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	_, err = conn.DisableImportFindingsForProductWithContext(ctx, &securityhub.DisableImportFindingsForProductInput{
		ProductSubscriptionArn: aws.String(productSubscriptionArn),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Product Subscription (%s): %s", d.Id(), err)
	}

	return diags
}
