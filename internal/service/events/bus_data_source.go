package events

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceBus() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBusRead,

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

func dataSourceBusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	name := d.Get("name").(string)

	input := &eventbridge.DescribeEventBusInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeEventBus(input)
	if err != nil {
		return fmt.Errorf("error reading EventBridge Bus (%s): %w", name, err)
	}

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)

	d.SetId(name)

	return nil
}
