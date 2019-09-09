package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsRoute53Zone(t *testing.T) {
	rInt := acctest.RandInt()
	publicResourceName := "aws_route53_zone.test"
	privateResourceName := "aws_route53_zone.test_private"
	serviceDiscoveryResourceName := "aws_service_discovery_private_dns_namespace.test_service_discovery"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(publicResourceName, "id", "data.aws_route53_zone.by_zone_id", "id"),
					resource.TestCheckResourceAttrPair(publicResourceName, "name", "data.aws_route53_zone.by_zone_id", "name"),
					resource.TestCheckResourceAttrPair(publicResourceName, "name_servers", "data.aws_route53_zone.by_zone_id", "name_servers"),
					resource.TestCheckResourceAttrPair(publicResourceName, "id", "data.aws_route53_zone.by_name", "id"),
					resource.TestCheckResourceAttrPair(publicResourceName, "name", "data.aws_route53_zone.by_name", "name"),
					resource.TestCheckResourceAttrPair(publicResourceName, "name_servers", "data.aws_route53_zone.by_name", "name_servers"),
					resource.TestCheckResourceAttrPair(privateResourceName, "id", "data.aws_route53_zone.by_vpc", "id"),
					resource.TestCheckResourceAttrPair(privateResourceName, "name", "data.aws_route53_zone.by_vpc", "name"),
					resource.TestCheckResourceAttrPair(privateResourceName, "id", "data.aws_route53_zone.by_tag", "id"),
					resource.TestCheckResourceAttrPair(privateResourceName, "name", "data.aws_route53_zone.by_tag", "name"),
					resource.TestCheckResourceAttrPair(serviceDiscoveryResourceName, "hosted_zone", "data.aws_route53_zone.service_discovery_by_vpc", "id"),
					resource.TestCheckResourceAttrPair(serviceDiscoveryResourceName, "name", "data.aws_route53_zone.service_discovery_by_vpc", "name"),
					resource.TestCheckResourceAttr("data.aws_route53_zone.service_discovery_by_vpc", "linked_service_principal", "servicediscovery.amazonaws.com"),
					resource.TestMatchResourceAttr("data.aws_route53_zone.service_discovery_by_vpc", "linked_service_description", regexp.MustCompile(`^arn:[^:]+:servicediscovery:[^:]+:[^:]+:namespace/ns-\w+$`)),
				),
			},
		},
	})
}

func testAccDataSourceAwsRoute53ZoneConfig(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source"
  }
}

resource "aws_route53_zone" "test_private" {
  name = "test.acc-%[1]d."

  vpc {
    vpc_id = "${aws_vpc.test.id}"
  }

  tags = {
    Environment = "dev-%[1]d"
  }
}

data "aws_route53_zone" "by_vpc" {
  name   = "${aws_route53_zone.test_private.name}"
  vpc_id = "${aws_vpc.test.id}"
}

data "aws_route53_zone" "by_tag" {
  name         = "${aws_route53_zone.test_private.name}"
  private_zone = true

  tags = {
    Environment = "dev-%[1]d"
  }
}

resource "aws_route53_zone" "test" {
  name = "terraformtestacchz-%[1]d.com."
}

data "aws_route53_zone" "by_zone_id" {
  zone_id = "${aws_route53_zone.test.zone_id}"
}

data "aws_route53_zone" "by_name" {
  name = "${data.aws_route53_zone.by_zone_id.name}"
}

resource "aws_service_discovery_private_dns_namespace" "test_service_discovery" {
  name        = "test.acc-sd-%[1]d."
  vpc         = "${aws_vpc.test.id}"
}

data "aws_route53_zone" "service_discovery_by_vpc" {
  name   = "${aws_service_discovery_private_dns_namespace.test_service_discovery.name}"
  vpc_id = "${aws_vpc.test.id}"
}
`, rInt)
}
