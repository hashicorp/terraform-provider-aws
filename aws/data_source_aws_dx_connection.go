package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDxConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jumbo_frame_capable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsDxConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	name := d.Get("name").(string)

	tags := tagsFromMapDX(d.Get("tags").(map[string]interface{}))

	connections := make([]*directconnect.Connection, 0)
	// DescribeDirectConnectionsInput does not have a name parameter for filtering
	input := &directconnect.DescribeConnectionsInput{}
	output, err := conn.DescribeConnections(input)

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connections %s", err)
	}

	for _, connection := range output.Connections {
		var tagsMatched int
		if aws.StringValue(connection.ConnectionName) == name {
			if len(tags) > 0 {
				tagsMatched = 0
				for _, tag := range tags {
					for _, tagRequested := range connection.Tags {
						if *tag.Key == *tagRequested.Key && *tag.Value == *tagRequested.Value {
							tagsMatched = tagsMatched + 1
						}
					}
				}

				if tagsMatched == len(tags) {
					connections = append(connections, connection)
				}

			} else {
				connections = append(connections, connection)
			}
		}
	}

	if len(connections) == 0 {
		return fmt.Errorf("Direct Connect Connection not found for name: %s", name)
	}

	if len(connections) > 1 {
		return fmt.Errorf("Multiple Direct Connect Connections found for name: %s", name)
	}

	connection := connections[0]

	d.SetId(aws.StringValue(connection.ConnectionId))

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("state", connection.ConnectionState)
	d.Set("location", connection.Location)
	d.Set("bandwidth", connection.Bandwidth)
	d.Set("jumbo_frame_capable", connection.JumboFrameCapable)

	return nil
}
