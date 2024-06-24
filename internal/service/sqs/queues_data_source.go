// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_sqs_queues")
func dataSourceQueues() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceQueuesRead,

		Schema: map[string]*schema.Schema{
			"queue_name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"queue_urls": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceQueuesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	input := &sqs.ListQueuesInput{}

	if v, ok := d.GetOk("queue_name_prefix"); ok {
		input.QueueNamePrefix = aws.String(v.(string))
	}

	var queueURLs []string
	pages := sqs.NewListQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing SQS Queues: %s", err)
		}

		queueURLs = append(queueURLs, page.QueueUrls...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("queue_urls", queueURLs)

	return diags
}
