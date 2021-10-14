package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGatewayRead,

		Schema: map[string]*schema.Schema{
			"amazon_side_asn": {
				Type:     schema.TypeString,
				Computed: true,
			},
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

func dataSourceGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn
	name := d.Get("name").(string)

	gateways := make([]*directconnect.Gateway, 0)
	// DescribeDirectConnectGatewaysInput does not have a name parameter for filtering
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	for {
		output, err := conn.DescribeDirectConnectGateways(input)
		if err != nil {
			return fmt.Errorf("error reading Direct Connect Gateway: %w", err)
		}
		for _, gateway := range output.DirectConnectGateways {
			if aws.StringValue(gateway.DirectConnectGatewayName) == name {
				gateways = append(gateways, gateway)
			}
		}
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	if len(gateways) == 0 {
		return fmt.Errorf("Direct Connect Gateway not found for name: %s", name)
	}

	if len(gateways) > 1 {
		return fmt.Errorf("Multiple Direct Connect Gateways found for name: %s", name)
	}

	gateway := gateways[0]

	d.SetId(aws.StringValue(gateway.DirectConnectGatewayId))
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(gateway.AmazonSideAsn), 10))
	d.Set("owner_account_id", gateway.OwnerAccount)

	return nil
}
