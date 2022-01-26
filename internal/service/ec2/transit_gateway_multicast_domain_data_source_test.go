package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTransitGatewayMulticastDomainDataSource_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"MulticastDomain": {
			"Filter": testAccTransitGatewayMulticastDomainDataSource_Filter,
			"ID":     testAccTransitGatewayMulticastDomainDataSource_ID,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccTransitGatewayMulticastDomainDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_id", dataSourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainFilterDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway_multicast_domain" "test" {}

data "aws_ec2_transit_gateway_multicast_domain" "test" {
  filter {
    name   = "transit-gateway-multicast-domain-id"
    values = [aws_ec2_transit_gateway_multicast_domain.test.id]
  }
}
`
}

func testAccTransitGatewayMulticastDomainIDDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway_multicast_domain" "test" {}

data "aws_ec2_transit_gateway_multicast_domain" "test" {
  id = aws_ec2_transit_gateway_multicast_domain.test.id
}
`
}
