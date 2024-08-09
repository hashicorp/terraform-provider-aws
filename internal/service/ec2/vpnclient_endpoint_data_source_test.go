// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccClientVPNEndpointDataSource_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	datasource1Name := "data.aws_ec2_client_vpn_endpoint.by_id"
	datasource2Name := "data.aws_ec2_client_vpn_endpoint.by_filter"
	datasource3Name := "data.aws_ec2_client_vpn_endpoint.by_tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckClientVPNSyncronize(t, semaphore)
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClientVPNEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClientVPNEndpointDataSourceConfig_basic(t, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasource1Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "client_vpn_endpoint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource1Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(datasource1Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource1Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource1Name, "self_service_portal_url", resourceName, "self_service_portal_url"),
					resource.TestCheckResourceAttrPair(datasource1Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource1Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource1Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasource1Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource1Name, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(datasource1Name, "vpn_port", resourceName, "vpn_port"),

					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasource2Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "client_vpn_endpoint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource2Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(datasource2Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "self_service_portal_url", resourceName, "self_service_portal_url"),
					resource.TestCheckResourceAttrPair(datasource2Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource2Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource2Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource2Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource2Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasource2Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource2Name, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(datasource2Name, "vpn_port", resourceName, "vpn_port"),

					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasource3Name, "authentication_options.#", resourceName, "authentication_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_connect_options.#", resourceName, "client_connect_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_login_banner_options.#", resourceName, "client_login_banner_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "client_vpn_endpoint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasource3Name, "connection_log_options.#", resourceName, "connection_log_options.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(datasource3Name, "dns_servers.#", resourceName, "dns_servers.#"),
					resource.TestCheckResourceAttrPair(datasource3Name, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(datasource2Name, "self_service_portal_url", resourceName, "self_service_portal_url"),
					resource.TestCheckResourceAttrPair(datasource3Name, "self_service_portal", resourceName, "self_service_portal"),
					resource.TestCheckResourceAttrPair(datasource3Name, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasource3Name, "session_timeout_hours", resourceName, "session_timeout_hours"),
					resource.TestCheckResourceAttrPair(datasource3Name, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasource3Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasource3Name, "transport_protocol", resourceName, "transport_protocol"),
					resource.TestCheckResourceAttrPair(datasource3Name, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(datasource3Name, "vpn_port", resourceName, "vpn_port"),
				),
			},
		},
	})
}

func testAccClientVPNEndpointDataSourceConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccClientVPNEndpointConfig_basic(t, rName), `
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
