package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStandardsSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStandardsSubscriptionCreate,
		ReadWithoutTimeout:   resourceStandardsSubscriptionRead,
		DeleteWithoutTimeout: resourceStandardsSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"standards_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceStandardsSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	standardsARN := d.Get("standards_arn").(string)
	input := &securityhub.BatchEnableStandardsInput{
		StandardsSubscriptionRequests: []*securityhub.StandardsSubscriptionRequest{
			{
				StandardsArn: aws.String(standardsARN),
			},
		},
	}

	log.Printf("[DEBUG] Creating Security Hub Standards Subscription: %s", input)
	output, err := conn.BatchEnableStandardsWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "enabling Security Hub Standard (%s): %s", standardsARN, err)
	}

	d.SetId(aws.StringValue(output.StandardsSubscriptions[0].StandardsSubscriptionArn))

	_, err = waitStandardsSubscriptionCreated(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) to create: %s", d.Id(), err)
	}

	return append(diags, resourceStandardsSubscriptionRead(ctx, d, meta)...)
}

func resourceStandardsSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	output, err := FindStandardsSubscriptionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("standards_arn", output.StandardsArn)

	return diags
}

func resourceStandardsSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	log.Printf("[DEBUG] Deleting Security Hub Standards Subscription: %s", d.Id())
	_, err := conn.BatchDisableStandardsWithContext(ctx, &securityhub.BatchDisableStandardsInput{
		StandardsSubscriptionArns: aws.StringSlice([]string{d.Id()}),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling Security Hub Standard (%s): %s", d.Id(), err)
	}

	_, err = waitStandardsSubscriptionDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Security Hub Standards Subscription (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
