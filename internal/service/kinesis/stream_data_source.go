// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kinesis_stream")
func DataSourceStream() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStreamRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"open_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrRetentionPeriod: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrStatus: {
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
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	stream, err := findStreamByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kinesis Stream (%s): %s", name, err)
	}

	input := &kinesis.ListShardsInput{
		StreamName: aws.String(name),
	}
	var shards []types.Shard

	err = listShardsPages(ctx, conn, input, func(page *kinesis.ListShardsOutput, lastPage bool) bool {
		shards = append(shards, page.Shards...)
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Kinesis Stream (%s) shards: %s", name, err)
	}

	// See http://docs.aws.amazon.com/kinesis/latest/dev/kinesis-using-sdk-java-resharding-merge.html.
	var openShards, closedShards []*string
	for _, shard := range shards {
		if shard.SequenceNumberRange.EndingSequenceNumber == nil {
			openShards = append(openShards, shard.ShardId)
		} else {
			closedShards = append(closedShards, shard.ShardId)
		}
	}

	d.SetId(aws.ToString(stream.StreamARN))
	d.Set(names.AttrARN, stream.StreamARN)
	d.Set("closed_shards", aws.ToStringSlice(closedShards))
	d.Set("creation_timestamp", aws.ToTime(stream.StreamCreationTimestamp).Unix())
	d.Set(names.AttrName, stream.StreamName)
	d.Set("open_shards", aws.ToStringSlice(openShards))
	d.Set(names.AttrRetentionPeriod, stream.RetentionPeriodHours)
	var shardLevelMetrics []types.MetricsName
	for _, v := range stream.EnhancedMonitoring {
		shardLevelMetrics = append(shardLevelMetrics, v.ShardLevelMetrics...)
	}
	d.Set("shard_level_metrics", shardLevelMetrics)
	d.Set(names.AttrStatus, stream.StreamStatus)
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

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
