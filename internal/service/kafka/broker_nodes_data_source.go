// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_broker_nodes", name="Broker Nodes")
func dataSourceBrokerNodes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerNodesRead,

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"node_info_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attached_eni_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"broker_id": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"client_subnet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_vpc_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEndpoints: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"node_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBrokerNodesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	clusterARN := d.Get("cluster_arn").(string)
	input := &kafka.ListNodesInput{
		ClusterArn: aws.String(clusterARN),
	}
	var nodeInfos []types.NodeInfo

	pages := kafka.NewListNodesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing MSK Cluster (%s) Broker Nodes: %s", clusterARN, err)
		}

		nodeInfos = append(nodeInfos, page.NodeInfoList...)
	}

	// node list is returned unsorted sort on broker id
	sort.Slice(nodeInfos, func(i, j int) bool {
		iBrokerId := aws.ToFloat64(nodeInfos[i].BrokerNodeInfo.BrokerId)
		jBrokerId := aws.ToFloat64(nodeInfos[j].BrokerNodeInfo.BrokerId)
		return iBrokerId < jBrokerId
	})

	tfList := []interface{}{}
	for _, apiObject := range nodeInfos {
		if brokerNodeInfo := apiObject.BrokerNodeInfo; brokerNodeInfo != nil {
			tfMap := map[string]interface{}{
				"attached_eni_id":       aws.ToString(brokerNodeInfo.AttachedENIId),
				"broker_id":             aws.ToFloat64(brokerNodeInfo.BrokerId),
				"client_subnet":         aws.ToString(brokerNodeInfo.ClientSubnet),
				"client_vpc_ip_address": aws.ToString(brokerNodeInfo.ClientVpcIpAddress),
				names.AttrEndpoints:     brokerNodeInfo.Endpoints,
				"node_arn":              aws.ToString(apiObject.NodeARN),
			}
			tfList = append(tfList, tfMap)
		}
	}

	d.SetId(clusterARN)
	if err := d.Set("node_info_list", tfList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_info_list: %s", err)
	}

	return diags
}
