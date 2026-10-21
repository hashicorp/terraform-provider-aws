// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_appstream_usage_report_subscription", name="Usage Report Subscription")
func resourceUsageReportSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUsageReportSubscriptionCreate,
		ReadWithoutTimeout:   resourceUsageReportSubscriptionRead,
		DeleteWithoutTimeout: resourceUsageReportSubscriptionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"s3_bucket_name": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"schedule": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func resourceUsageReportSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	output, err := conn.CreateUsageReportSubscription(ctx, &appstream.CreateUsageReportSubscriptionInput{})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream Usage Report Subscription: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID(ctx))
	d.Set("s3_bucket_name", output.S3BucketName)
	d.Set("schedule", output.Schedule)

	return append(diags, resourceUsageReportSubscriptionRead(ctx, d, meta)...)
}

func resourceUsageReportSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	subscription, err := findUsageReportSubscription(ctx, conn)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] AppStream Usage Report Subscription (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream Usage Report Subscription (%s): %s", d.Id(), err)
	}

	d.Set("s3_bucket_name", subscription.S3BucketName)
	d.Set("schedule", subscription.Schedule)

	return diags
}

func resourceUsageReportSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	log.Printf("[DEBUG] Deleting AppStream Usage Report Subscription: %s", d.Id())
	_, err := conn.DeleteUsageReportSubscription(ctx, &appstream.DeleteUsageReportSubscriptionInput{})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream Usage Report Subscription (%s): %s", d.Id(), err)
	}

	return diags
}

func findUsageReportSubscription(ctx context.Context, conn *appstream.Client) (*awstypes.UsageReportSubscription, error) {
	output, err := conn.DescribeUsageReportSubscriptions(ctx, &appstream.DescribeUsageReportSubscriptionsInput{})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.UsageReportSubscriptions)
}
