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

	var connections []*directconnect.Connection
	input := &directconnect.DescribeConnectionsInput{}
	name := d.Get("name").(string)

	// DescribeConnections is not paginated.
	output, err := conn.DescribeConnections(input)

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connections: %w", err)
	}

	for _, connection := range output.Connections {
		if aws.StringValue(connection.ConnectionName) == name {
			connections = append(connections, connection)
		}
	}

	switch count := len(connections); count {
	case 0:
		return fmt.Errorf("no matching Direct Connect Connection found")
	case 1:
	default:
		return fmt.Errorf("%d Direct Connect Connections matched; use additional constraints to reduce matches to a single Direct Connect Connection", count)
	}

	connection := connections[0]

	d.SetId(aws.StringValue(connection.ConnectionId))
	d.Set("name", connection.ConnectionName)
	d.Set("owner_account_id", connection.OwnerAccount)

	return nil
}
