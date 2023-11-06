// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_kinesis_stream_consumer")
func DataSourceStreamConsumer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStreamConsumerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},

			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceStreamConsumerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)

	streamArn := d.Get("stream_arn").(string)

	input := &kinesis.ListStreamConsumersInput{
		StreamARN: aws.String(streamArn),
	}

	var results []*kinesis.Consumer

	err := conn.ListStreamConsumersPagesWithContext(ctx, input, func(page *kinesis.ListStreamConsumersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, consumer := range page.Consumers {
			if consumer == nil {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(consumer.ConsumerName) {
				continue
			}

			if v, ok := d.GetOk("arn"); ok && v.(string) != aws.StringValue(consumer.ConsumerARN) {
				continue
			}

			results = append(results, consumer)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Kinesis Stream Consumers: %s", err)
	}

	if len(results) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Kinesis Stream Consumer found matching criteria; try different search")
	}

	if len(results) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Kinesis Stream Consumers found matching criteria; try different search")
	}

	consumer := results[0]

	d.SetId(aws.StringValue(consumer.ConsumerARN))
	d.Set("arn", consumer.ConsumerARN)
	d.Set("name", consumer.ConsumerName)
	d.Set("status", consumer.ConsumerStatus)
	d.Set("stream_arn", streamArn)
	d.Set("creation_timestamp", aws.TimeValue(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))

	return diags
}
