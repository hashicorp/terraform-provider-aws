package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsDxConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxConnectionRead,

		Schema: map[string]*schema.Schema{
			"id": {
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

func dataSourceAwsDxConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	name := d.Get("name").(string)

	connections := make([]*directconnect.Connection, 0)
	// DescribeDirectConnectionsInput does not have a name parameter for filtering
	input := &directconnect.DescribeConnectionsInput{}
	output, err := conn.DescribeConnections(input)

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connections %s", err)
	}

	for _, connection := range output.Connections {
		if aws.StringValue(connection.ConnectionName) == name {
			connections = append(connections, connection)
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
	d.Set("state", aws.StringValue(connection.ConnectionState))
	d.Set("location", aws.StringValue(connection.Location))
	d.Set("bandwidth", aws.StringValue(connection.Bandwidth))
	d.Set("jumbo_frame_capable", strconv.FormatBool(aws.BoolValue(connection.JumboFrameCapable)))

	return nil
}
