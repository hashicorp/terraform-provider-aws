// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverEndpointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_endpoint.test"
	datasourceName := "data.aws_route53_resolver_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "direction", resourceName, "direction"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_addresses.#", resourceName, "ip_address.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "protocols.#", resourceName, "protocols.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_type", resourceName, "resolver_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, "host_vpc_id"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverEndpointDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_endpoint.test"
	datasourceName := "data.aws_route53_resolver_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{

			{
				Config: testAccEndpointDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "direction", resourceName, "direction"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "ip_addresses.#", resourceName, "ip_address.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "protocols.#", resourceName, "protocols.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_type", resourceName, "resolver_endpoint_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, "host_vpc_id"),
				),
			},
		},
	})
}

func testAccEndpointDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_basic(rName), `
data "aws_route53_resolver_endpoint" "test" {
  resolver_endpoint_id = aws_route53_resolver_endpoint.test.id
}
`)
}

func testAccEndpointDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_outbound(rName, rName), `
data "aws_route53_resolver_endpoint" "test" {
  filter {
    name   = "Name"
    values = [aws_route53_resolver_endpoint.test.name]
  }

  filter {
    name   = "SecurityGroupIds"
    values = aws_security_group.test[*].id
  }
}
`)
}
