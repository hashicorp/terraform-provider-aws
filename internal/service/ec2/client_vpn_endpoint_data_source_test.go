package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccClientVPNEndpointDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	datasource1Name := "data.aws_ec2_client_vpn_endpoint.by_id"
	datasource2Name := "data.aws_ec2_client_vpn_endpoint.by_filter"
	datasource3Name := "data.aws_ec2_client_vpn_endpoint.by_tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasource1Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(datasource2Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(datasource3Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccEc2ClientVpnEndpointDataSourceConfig(rName string) string {
	return acctest.ConfigCompose(testAccEc2ClientVpnEndpointConfig(rName), `
data "aws_ec2_client_vpn_endpoint" "by_id" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
}

data "aws_ec2_client_vpn_endpoint" "by_tags" {
  tags = {
    Name = aws_ec2_client_vpn_endpoint.test.tags["Name"]
  }
}

data "aws_ec2_client_vpn_endpoint" "by_filter" {
  filter {
    name   = "endpoint-id"
    values = [aws_ec2_client_vpn_endpoint.test.id]
  }
}
`)
}
