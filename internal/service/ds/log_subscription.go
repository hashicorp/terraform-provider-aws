// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_directory_service_log_subscription")
func ResourceLogSubscription() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLogSubscriptionCreate,
		ReadWithoutTimeout:   resourceLogSubscriptionRead,
		DeleteWithoutTimeout: resourceLogSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"log_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLogSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId := d.Get("directory_id")
	logGroupName := d.Get("log_group_name")

	input := directoryservice.CreateLogSubscriptionInput{
		DirectoryId:  aws.String(directoryId.(string)),
		LogGroupName: aws.String(logGroupName.(string)),
	}

	_, err := conn.CreateLogSubscriptionWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Directory Service Log Subscription: %s", err)
	}

	d.SetId(directoryId.(string))

	return append(diags, resourceLogSubscriptionRead(ctx, d, meta)...)
}

func resourceLogSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId := d.Id()

	input := directoryservice.ListLogSubscriptionsInput{
		DirectoryId: aws.String(directoryId),
	}

	out, err := conn.ListLogSubscriptionsWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Directory Service Log Subscription: %s", err)
	}

	if len(out.LogSubscriptions) == 0 {
		log.Printf("[WARN] No log subscriptions for directory %s found", directoryId)
		d.SetId("")
		return diags
	}

	logSubscription := out.LogSubscriptions[0]
	d.Set("directory_id", logSubscription.DirectoryId)
	d.Set("log_group_name", logSubscription.LogGroupName)

	return diags
}

func resourceLogSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSConn(ctx)

	directoryId := d.Id()

	input := directoryservice.DeleteLogSubscriptionInput{
		DirectoryId: aws.String(directoryId),
	}

	_, err := conn.DeleteLogSubscriptionWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Directory Service Log Subscription: %s", err)
	}

	return diags
}
