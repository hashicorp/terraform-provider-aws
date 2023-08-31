// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_sqs_queues")
func DataSourceQueues() *schema.Resource {
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

const (
	DSNameQueues = "Queues Data Source"
)

func dataSourceQueuesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn(ctx)

	input := &sqs.ListQueuesInput{}

	if v, ok := d.GetOk("queue_name_prefix"); ok {
		input.QueueNamePrefix = aws.String(v.(string))
	}

	var queueUrls []string

	err := conn.ListQueuesPagesWithContext(ctx, input, func(page *sqs.ListQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queueUrl := range page.QueueUrls {
			if queueUrl == nil {
				continue
			}

			queueUrls = append(queueUrls, aws.StringValue(queueUrl))
		}

		return !lastPage
	})

	if err != nil {
		return create.DiagError(names.SQS, create.ErrActionReading, DSNameQueues, "", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("queue_urls", queueUrls)

	return nil
}
