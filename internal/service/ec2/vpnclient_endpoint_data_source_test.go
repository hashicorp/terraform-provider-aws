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
		PreCheck:          func() { testAccPreCheckClientVPNSyncronize(t); acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClientVPNEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasource1Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource1Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasource1Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasource1Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource1Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource1Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource1Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource1Name, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasource1Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource1Name, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(datasource1Name, "vpn_port", resourceName, "vpn_port"),

					resource.TestCheckResourceAttrPair(datasource2Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource2Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasource2Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasource2Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource2Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource2Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource2Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource2Name, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasource2Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource2Name, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(datasource2Name, "vpn_port", resourceName, "vpn_port"),

					resource.TestCheckResourceAttrPair(datasource3Name, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasource3Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_vpn_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasource3Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasource3Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource3Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource3Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource3Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource3Name, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasource3Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource3Name, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(datasource3Name, "vpn_port", resourceName, "vpn_port"),
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
