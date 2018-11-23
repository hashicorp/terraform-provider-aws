package aws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsKinesisStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKinesisStreamRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"stream_arn": {
				Type:     schema.TypeString,
				Computed: true,
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

	state, err := readKinesisStreamConsumerState(conn, cn)
	if err != nil {
		return err
	}
	d.SetId(state.arn)
	d.Set("arn", state.arn)
	d.Set("name", cn)
	d.Set("stream_arn", state.streamArn)
	d.Set("status", state.status)
	d.Set("creation_timestamp", state.creationTimestamp)

	if err != nil {
		return err
	}

	return nil
}
