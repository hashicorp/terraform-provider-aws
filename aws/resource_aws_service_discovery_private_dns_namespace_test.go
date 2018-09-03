package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_basic(t *testing.T) {
	rName := acctest.RandString(5) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists("aws_service_discovery_private_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "hosted_zone"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_longname(t *testing.T) {
	rName := acctest.RandString(64-len("example.com")) + ".example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists("aws_service_discovery_private_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_private_dns_namespace.test", "hosted_zone"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_private_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags {
    Name = "terraform-testacc-service-discovery-private-dns-ns"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%s"
  description = "test"
  vpc = "${aws_vpc.test.id}"
}
`, rName)
}
