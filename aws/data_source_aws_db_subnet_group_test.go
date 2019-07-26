package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDbSubnetGroupDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBSubnetGroupDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "description"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "name"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "status"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "subnets.0"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "subnets.1"),
					resource.TestCheckResourceAttrSet("data.aws_db_subnet_group.bar", "vpc_id"),
				),
			},
		},
	})
}

func testAccAWSDBSubnetGroupDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "foo" {
  count             = 2
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}" 
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.foo.id}"
}

resource "aws_db_subnet_group" "bar" {
  name       = "datasource-test-terraform-%d"
  subnet_ids = "${aws_subnet.foo.*.id}"
}

data "aws_db_subnet_group" "bar" {
  name = "${aws_db_subnet_group.bar.name}"
}
`, rInt)
}
