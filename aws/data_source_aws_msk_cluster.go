package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsMskCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zookeeper_connect": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	state, err := readMskClusterState(conn, d.Id())
	if err != nil {
		return err
	}

	d.SetId(*state.ClusterArn)
	d.Set("arn", *state.ClusterArn)
	d.Set("status", *state.State)
	d.Set("creation_timestamp", *state.CreationTime)
	d.Set("broker_count", *state.NumberOfBrokerNodes)
	d.Set("encryption_key", *state.EncryptionInfo.EncryptionAtRest.DataVolumeKMSKeyId)
	d.Set("zookeeper_connect", *state.ZookeeperConnectString)

	if *state.State == kafka.ClusterStateActive {
		bb, err := conn.GetBootstrapBrokers(&kafka.GetBootstrapBrokersInput{ClusterArn: state.ClusterArn})
		if err != nil {
			return err
		}

		d.Set("bootstrap_brokers", aws.StringValue(bb.BootstrapBrokerString))
	}

	return nil
}
