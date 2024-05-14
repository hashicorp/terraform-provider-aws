// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kinesis_stream_consumer", name="Stream Consumer)
func dataSourceStreamConsumer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStreamConsumerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStreamARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceStreamConsumerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	streamARN := d.Get(names.AttrStreamARN).(string)
	input := &kinesis.ListStreamConsumersInput{
		StreamARN: aws.String(streamARN),
	}

	consumer, err := findStreamConsumer(ctx, conn, input, func(c *types.Consumer) bool {
		if v, ok := d.GetOk(names.AttrName); ok && v.(string) != aws.ToString(c.ConsumerName) {
			return false
		}

		if v, ok := d.GetOk(names.AttrARN); ok && v.(string) != aws.ToString(c.ConsumerARN) {
			return false
		}

		return true
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Kinesis Stream Consumer", err))
	}

	d.SetId(aws.ToString(consumer.ConsumerARN))
	d.Set(names.AttrARN, consumer.ConsumerARN)
	d.Set("creation_timestamp", aws.ToTime(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))
	d.Set(names.AttrName, consumer.ConsumerName)
	d.Set(names.AttrStatus, consumer.ConsumerStatus)
	d.Set(names.AttrStreamARN, streamARN)

	return diags
}

func findStreamConsumer(ctx context.Context, conn *kinesis.Client, input *kinesis.ListStreamConsumersInput, filter tfslices.Predicate[*types.Consumer]) (*types.Consumer, error) {
	output, err := findStreamConsumers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStreamConsumers(ctx context.Context, conn *kinesis.Client, input *kinesis.ListStreamConsumersInput, filter tfslices.Predicate[*types.Consumer]) ([]types.Consumer, error) {
	var output []types.Consumer

	pages := kinesis.NewListStreamConsumersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Consumers {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
