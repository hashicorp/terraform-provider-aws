package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsMksCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMksClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"broker_node_group_info": {
				Type:     schema.TypeString,
				Required: true,
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
				Required: true,
			},
			"broker_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsMksClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn
	arn := d.Get("arn").(string)

	state, err := readMskClusterState(conn, arn)
	if err != nil {
		return err
	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("status", state.status)
	d.Set("creation_timestamp", state.creationTimestamp)
	d.Set("encrypt_rest_arn", state.encryptRestArn)

	return nil
}
