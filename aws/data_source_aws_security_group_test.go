package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
					resource.TestCheckResourceAttrPair("data.aws_security_group.default_by_name", "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair("data.aws_security_group.default_by_name", "id", "aws_vpc.test", "default_security_group_id"),
				),
			},
		},
	})
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
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name    = %[1]q
    SomeTag = "SomeValue"
  }

  description = "sg description"
}

data "aws_security_group" "by_id" {
  id = aws_security_group.test.id
}

data "aws_security_group" "by_name" {
  name = aws_security_group.test.name
}

data "aws_security_group" "default_by_name" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_security_group" "by_tag" {
  tags = {
    Name = "${aws_security_group.test.tags["Name"]}"
  }
}

data "aws_security_group" "by_filter" {
  filter {
    name   = "group-name"
    values = [aws_security_group.test.name]
  }
}
`, rName)
}
