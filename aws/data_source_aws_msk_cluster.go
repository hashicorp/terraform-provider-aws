package aws

import (
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
				Optional: true,
			},
			"encrypt_rest_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enhanced_monitoring": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kafka_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"bootstrap_brokers": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zookeeper_connect": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsMskClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	name := d.Get("name").(string)

	state, err := readMskClusterState(conn, name)
	if err != nil {
		return err
	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("status", state.status)
	d.Set("creation_timestamp", state.creationTimestamp)
	d.Set("broker_count", state)
	d.Set("encrypt_rest_key", state.encryptRestKey)
	d.Set("bootstrap_brokers", state.bootstrapBrokers)
	d.Set("zookeeper_connect", state.zookeeperConnect)

	return nil
}
