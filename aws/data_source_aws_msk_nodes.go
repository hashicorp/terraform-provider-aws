package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sort"
)

func dataSourceAwsMskNodes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskNodesRead,

		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"nodes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"broker_id": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"attached_eni_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_subnet": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsMskNodesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	listNodesInput := &kafka.ListNodesInput{
		ClusterArn: aws.String(d.Get("cluster_arn").(string)),
	}

	var nodes []*kafka.NodeInfo
	for {
		listNodesOutput, err := conn.ListNodes(listNodesInput)

		if err != nil {
			return fmt.Errorf("error listing MSK Nodes: %w", err)
		}

		if listNodesOutput == nil {
			break
		}

		nodes = append(nodes, listNodesOutput.NodeInfoList...)

		if aws.StringValue(listNodesOutput.NextToken) == "" {
			break
		}

		listNodesInput.NextToken = listNodesOutput.NextToken
	}

	if len(nodes) == 0 {
		return fmt.Errorf("error reading MSK Nodes: no results found")
	}

	// node list is returned unsorted sort on broker id
	sort.Slice(nodes, func(i, j int) bool {
		iBrokerId := aws.Float64Value(nodes[i].BrokerNodeInfo.BrokerId)
		jBrokerId := aws.Float64Value(nodes[j].BrokerNodeInfo.BrokerId)
		return iBrokerId < jBrokerId
	})

	brokerList := make([]interface{}, len(nodes))
	for i, node := range nodes {
		broker := map[string]interface{}{
			"broker_id":       aws.Float64Value(node.BrokerNodeInfo.BrokerId),
			"attached_eni_id": aws.StringValue(node.BrokerNodeInfo.AttachedENIId),
			"client_subnet":   aws.StringValue(node.BrokerNodeInfo.ClientSubnet),
			"endpoints":       aws.StringValueSlice(node.BrokerNodeInfo.Endpoints),
		}
		brokerList[i] = broker
	}

	d.SetId(d.Get("cluster_arn").(string))
	d.Set("nodes", brokerList)
	return nil
}
