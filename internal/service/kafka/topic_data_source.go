// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_topic", name="Topic")
func dataSourceTopic() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTopicRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"configs": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"partition_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"replication_factor": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"topic_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceTopicRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterArn := d.Get("cluster_arn").(string)
	topicName := d.Get("topic_name").(string)

	output, err := findTopicByTwoPartKey(ctx, conn, clusterArn, topicName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Topic (%s/%s): %s", clusterArn, topicName, err)
	}

	id, err := flex.FlattenResourceId([]string{clusterArn, topicName}, topicResourceIDPartCount, false)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.SetId(id)

	d.Set(names.AttrARN, output.TopicArn)
	d.Set("cluster_arn", clusterArn)
	d.Set("configs", output.Configs)
	d.Set("partition_count", output.PartitionCount)
	d.Set("replication_factor", output.ReplicationFactor)
	d.Set("topic_name", output.TopicName)

	return diags
}
