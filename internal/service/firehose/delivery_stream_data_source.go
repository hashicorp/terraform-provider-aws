package firehose

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceDeliveryStream() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDeliveryStreamRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceDeliveryStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).FirehoseConn

	sn := d.Get("name").(string)
	output, err := FindDeliveryStreamByName(conn, sn)

	if err != nil {
		return fmt.Errorf("error reading Kinesis Firehose Delivery Stream (%s): %w", sn, err)
	}

	d.SetId(aws.StringValue(output.DeliveryStreamARN))
	d.Set("arn", output.DeliveryStreamARN)
	d.Set("name", output.DeliveryStreamName)

	return nil
}
