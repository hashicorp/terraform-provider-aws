package aws

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceBrokerNodes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBrokerNodesRead,

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
						"endpoints": {
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

func dataSourceBrokerNodesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	clusterARN := d.Get("cluster_arn").(string)
	input := &kafka.ListNodesInput{
		ClusterArn: aws.String(clusterARN),
	}
	var nodeInfos []*kafka.NodeInfo

	err := conn.ListNodesPages(input, func(page *kafka.ListNodesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		nodeInfos = append(nodeInfos, page.NodeInfoList...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Cluster (%s) Broker Nodes: %w", clusterARN, err)
	}

	// node list is returned unsorted sort on broker id
	sort.Slice(nodeInfos, func(i, j int) bool {
		iBrokerId := aws.Float64Value(nodeInfos[i].BrokerNodeInfo.BrokerId)
		jBrokerId := aws.Float64Value(nodeInfos[j].BrokerNodeInfo.BrokerId)
		return iBrokerId < jBrokerId
	})

	tfList := make([]interface{}, len(nodeInfos))

	for i, apiObject := range nodeInfos {
		brokerNodeInfo := apiObject.BrokerNodeInfo
		tfMap := map[string]interface{}{
			"attached_eni_id":       aws.StringValue(brokerNodeInfo.AttachedENIId),
			"broker_id":             aws.Float64Value(brokerNodeInfo.BrokerId),
			"client_subnet":         aws.StringValue(brokerNodeInfo.ClientSubnet),
			"client_vpc_ip_address": aws.StringValue(brokerNodeInfo.ClientVpcIpAddress),
			"endpoints":             aws.StringValueSlice(brokerNodeInfo.Endpoints),
			"node_arn":              aws.StringValue(apiObject.NodeARN),
		}

		tfList[i] = tfMap
	}

	d.SetId(clusterARN)

	if err := d.Set("node_info_list", tfList); err != nil {
		return fmt.Errorf("error setting node_info_list: %w", err)
	}

	return nil
}
