package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSServiceDiscoveryDnsNamespace_basic(t *testing.T) {
	resourceName := "data.aws_service_discovery_dns_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsServiceDiscoveryDnsNamespaceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_service_discovery_private_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", "aws_service_discovery_private_dns_namespace.test", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", "aws_service_discovery_private_dns_namespace.test", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "hosted_zone", "aws_service_discovery_private_dns_namespace.test", "hosted_zone"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryDnsNamespaceConfig() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "example.terraform.local"
  description = "example"
  vpc         = "${aws_vpc.test.id}"
}

data "aws_service_discovery_dns_namespace" "test" {
  name = "${aws_service_discovery_private_dns_namespace.test.name}"
  dns_type = "DNS_PRIVATE"
}
`)
}
