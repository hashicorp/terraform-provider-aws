// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_kinesis_stream")
func DataSourceStream() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStreamRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"closed_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creation_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"open_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_mode_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	name := d.Get("name").(string)

	stream, err := FindStreamByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream (%s): %s", name, err)
	}

	input := &kinesis.ListShardsInput{
		StreamName: aws.String(name),
	}
	var shards []*kinesis.Shard

	for {
		output, err := conn.ListShardsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Kinesis Stream (%s) shards: %s", name, err)
		}

		if output == nil {
			break
		}

		shards = append(shards, output.Shards...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input = &kinesis.ListShardsInput{
			NextToken: output.NextToken,
		}
	}

	d.SetId(aws.StringValue(stream.StreamARN))
	d.Set("arn", stream.StreamARN)

	var closedShards []*string
	for _, v := range filterShards(shards, false) {
		closedShards = append(closedShards, v.ShardId)
	}

	d.Set("closed_shards", aws.StringValueSlice(closedShards))
	d.Set("creation_timestamp", aws.TimeValue(stream.StreamCreationTimestamp).Unix())
	d.Set("name", stream.StreamName)

	var openShards []*string
	for _, v := range filterShards(shards, true) {
		openShards = append(openShards, v.ShardId)
	}
	d.Set("open_shards", aws.StringValueSlice(openShards))

	d.Set("retention_period", stream.RetentionPeriodHours)

	var shardLevelMetrics []*string
	for _, v := range stream.EnhancedMonitoring {
		shardLevelMetrics = append(shardLevelMetrics, v.ShardLevelMetrics...)
	}
	d.Set("shard_level_metrics", aws.StringValueSlice(shardLevelMetrics))

	d.Set("status", stream.StreamStatus)

	if details := stream.StreamModeDetails; details != nil {
		if err := d.Set("stream_mode_details", []interface{}{flattenStreamModeDetails(details)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting stream_mode_details: %s", err)
		}
	} else {
		d.Set("stream_mode_details", nil)
	}

	tags, err := listTags(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Kinesis Stream (%s): %s", name, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

// See http://docs.aws.amazon.com/kinesis/latest/dev/kinesis-using-sdk-java-resharding-merge.html
func filterShards(shards []*kinesis.Shard, open bool) []*kinesis.Shard {
	var output []*kinesis.Shard

	for _, shard := range shards {
		if open && shard.SequenceNumberRange.EndingSequenceNumber == nil {
			output = append(output, shard)
		} else if !open && shard.SequenceNumberRange.EndingSequenceNumber != nil {
			output = append(output, shard)
		}
	}

	return output
}
