package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSpotDataFeedSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSpotDataFeedSubscriptionCreate,
		ReadWithoutTimeout:   resourceSpotDataFeedSubscriptionRead,
		DeleteWithoutTimeout: resourceSpotDataFeedSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSpotDataFeedSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.CreateSpotDatafeedSubscriptionInput{
		Bucket: aws.String(d.Get("bucket").(string)),
	}

	if v, ok := d.GetOk("prefix"); ok {
		input.Prefix = aws.String(v.(string))
	}

	_, err := conn.CreateSpotDatafeedSubscriptionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Spot Datafeed Subscription: %s", err)
	}

	d.SetId("spot-datafeed-subscription")

	return append(diags, resourceSpotDataFeedSubscriptionRead(ctx, d, meta)...)
}

func resourceSpotDataFeedSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	subscription, err := FindSpotDatafeedSubscription(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Spot Datafeed Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Spot Datafeed Subscription (%s): %s", d.Id(), err)
	}

	d.Set("bucket", subscription.Bucket)
	d.Set("prefix", subscription.Prefix)

	return diags
}

func resourceSpotDataFeedSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 Spot Datafeed Subscription: %s", d.Id())
	_, err := conn.DeleteSpotDatafeedSubscriptionWithContext(ctx, &ec2.DeleteSpotDatafeedSubscriptionInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Spot Datafeed Subscription (%s): %s", d.Id(), err)
	}

	return diags
}
