package ds_test

import (
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccVpcBase() string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}
`,
	)
}
