package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccDataSourceAwsSecurityGroups_tag(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dataSourceName := "data.aws_security_groups.by_tag"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecurityGroupsConfig_tag(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "3"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsSecurityGroups_filter(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dataSourceName := "data.aws_security_groups.by_filter"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsSecurityGroupsConfig_filter(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "3"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSecurityGroupsConfig_tag(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_tag" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-data-source"
  }
}

resource "aws_security_group" "test" {
  count  = 3
  vpc_id = aws_vpc.test_tag.id
  name   = "tf-%[1]d-${count.index}"

  tags = {
    Seed = "%[1]d"
  }
}

data "aws_security_groups" "by_tag" {
  tags = {
    Seed = aws_security_group.test[0].tags["Seed"]
  }
}
`, rInt)
}

func testAccDataSourceAwsSecurityGroupsConfig_filter(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test_filter" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-data-source"
  }
}

resource "aws_security_group" "test" {
  count  = 3
  vpc_id = aws_vpc.test_filter.id
  name   = "tf-%[1]d-${count.index}"

  tags = {
    Seed = "%[1]d"
  }
}

data "aws_security_groups" "by_filter" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test_filter.id]
  }

  filter {
    name   = "group-name"
    values = ["tf-${aws_security_group.test[0].tags["Seed"]}-*"]
  }
}
`, rInt)
}
