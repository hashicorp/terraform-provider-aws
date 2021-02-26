package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDxConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxConnectionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDxConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	name := d.Get("name").(string)

	connections := make([]*directconnect.Connection, 0)
	// DescribeConnectionsInput does not have a name parameter for filtering
	input := &directconnect.DescribeConnectionsInput{}
	for {
		output, err := conn.DescribeConnections(input)
		if err != nil {
			return fmt.Errorf("error reading Direct Connect Connections: %w", err)
		}
		for _, connection := range output.Connections {
			if aws.StringValue(connection.ConnectionName) == name {
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
	d.Set("owner_account_id", aws.StringValue(connection.OwnerAccount))

	return nil
}
