// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_msk_serverless_cluster", name="Serverless Cluster")
func dataSourceServerlessCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServerlessClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers_sasl_iam": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"cluster_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceServerlessClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName := d.Get("cluster_name").(string)
	input := &kafka.ListClustersV2Input{
		ClusterTypeFilter: aws.String("SERVERLESS"),
		ClusterNameFilter: aws.String(clusterName),
	}
	cluster, err := findServerlessCluster(ctx, conn, input, func(v *types.Cluster) bool {
		return aws.ToString(v.ClusterName) == clusterName
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s): %s", clusterName, err)
	}

	clusterARN := aws.ToString(cluster.ClusterArn)
	bootstrapBrokersOutput, err := findBootstrapBrokersByARN(ctx, conn, clusterARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Cluster (%s) bootstrap brokers: %s", clusterARN, err)
	}

	d.SetId(clusterARN)
	d.Set("arn", clusterARN)
	d.Set("bootstrap_brokers_sasl_iam", SortEndpointsString(aws.ToString(bootstrapBrokersOutput.BootstrapBrokerStringSaslIam)))
	d.Set("cluster_name", cluster.ClusterName)
	clusterUUID, _ := clusterUUIDFromARN(clusterARN)
	d.Set("cluster_uuid", clusterUUID)

	if err := d.Set("tags", KeyValueTags(ctx, cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func findServerlessCluster(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersV2Input, filter tfslices.Predicate[*types.Cluster]) (*types.Cluster, error) {
	output, err := findServerlessClusters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func findServerlessClusters(ctx context.Context, conn *kafka.Client, input *kafka.ListClustersV2Input, filter tfslices.Predicate[*types.Cluster]) ([]types.Cluster, error) {
	var output []types.Cluster

	pages := kafka.NewListClustersV2Paginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ClusterInfoList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
