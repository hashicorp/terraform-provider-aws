package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2VPCsDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue("data.aws_vpcs.test", "ids.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2VPCsDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpcs.test", "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccEC2VPCsDataSource_filters(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDataSourceConfig_filters(rName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue("data.aws_vpcs.test", "ids.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2VPCsDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCsDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpcs.test", "ids.#", "0"),
				),
			},
		},
	})
}

func testCheckResourceAttrGreaterThanValue(n, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if v, ok := rs.Primary.Attributes[key]; !ok || !(v > value) {
			if !ok {
				return fmt.Errorf("%s: Attribute %q not found", n, key)
			}

			return fmt.Errorf("%s: Attribute %q is not greater than %q, got %q", n, key, value, v)
		}

		return nil

	}
}

func testAccVPCsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpcs" "test" {}
`, rName)
}

func testAccVPCsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name    = %[1]q
    Service = "testacc-test"
  }
}

data "aws_vpcs" "test" {
  tags = {
    Name    = %[1]q
    Service = aws_vpc.test.tags["Service"]
  }
}
`, rName)
}

func testAccVPCsDataSourceConfig_filters(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/25"

  tags = {
    Name = %[1]q
  }
}

data "aws_vpcs" "test" {
  filter {
    name   = "cidr"
    values = [aws_vpc.test.cidr_block]
  }
}
`, rName)
}

func testAccVPCsDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_vpcs" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
