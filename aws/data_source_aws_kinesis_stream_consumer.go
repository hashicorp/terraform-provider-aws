package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsKinesisStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKinesisStreamConsumerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"stream_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"creation_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsKinesisStreamConsumerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn
	cn := d.Get("name").(string)
	sa := d.Get("stream_arn").(string)

	state, err := readKinesisStreamConsumerState(conn, cn, sa)
	if err != nil {
		return err
	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("name", cn)
	d.Set("stream_arn", sa)
	d.Set("status", state.status)
	d.Set("creation_timestamp", state.creationTimestamp)

	return nil
}
