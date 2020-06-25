package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsSecurityGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_security_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecurityGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_id", "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_id", "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_id", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_id", "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_id", "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_name", "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_name", "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_name", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_name", "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_name", "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_filter", "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_filter", "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_filter", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_filter", "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_filter", "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_tag", "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_tag", "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_tag", "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_tag", "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.by_tag", "vpc_id", resourceName, "vpc_id"),
					testAccDataSourceAwsSecurityGroupCheckDefault("data.aws_security_group.default_by_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSecurityGroupCheckDefault(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vpcRs, ok := s.RootModule().Resources["aws_vpc.test"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc.test in state")
		}
		attr := rs.Primary.Attributes

		if attr["id"] != vpcRs.Primary.Attributes["default_security_group_id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				vpcRs.Primary.Attributes["default_security_group_id"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsSecurityGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
<<<<<<< HEAD
  vpc_id = aws_vpc.test.id
  name   = "test-%d"
=======
  vpc_id = aws_vpc.test.id
  name   = %[1]q
>>>>>>> 5d0742b09 (use aws sdk wrappers, refactor tests)

