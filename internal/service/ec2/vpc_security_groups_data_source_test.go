package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCSecurityGroupsDataSource_tag(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupsDataSourceConfig_tag(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", "3"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupsDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupsDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_security_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupsDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_ids.#", "0"),
				),
			},
		},
	})
}

func testAccSecurityGroupsDataSourceConfig_tag(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count  = 3
  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

data "aws_security_groups" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_security_group.test[0], aws_security_group.test[1], aws_security_group.test[2]]
}
`, rName)
}

func testAccSecurityGroupsDataSourceConfig_filter(rName string) string {
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
    Name = %[1]q
  }
}

data "aws_security_groups" "test" {
  filter {
    name   = "vpc-id"
    values = [aws_vpc.test.id]
  }

  filter {
    name   = "group-name"
    values = [aws_security_group.test.name]
  }
}
`, rName)
}

func testAccSecurityGroupsDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
data "aws_security_groups" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
