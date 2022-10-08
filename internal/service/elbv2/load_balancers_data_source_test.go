package elbv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccELBV2LoadBalancersDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	lbName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sharedTagVal := sdkacctest.RandString(32)

	resourceLb1 := "aws_lb.test1"
	resourceLb2 := "aws_lb.test2"

	dataSourceNameMatchFirstTag := "data.aws_lbs.tag_match_first"
	dataSourceNameMatchBothTag := "data.aws_lbs.tag_match_shared"
	dataSourceNameMatchNoneTag := "data.aws_lbs.tag_match_none"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(elbv2.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancersDataSourceConfig_basic(rName, lbName1, lbName2, sharedTagVal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceNameMatchFirstTag, "arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchFirstTag, "arns.*", resourceLb1, "arn"),
					resource.TestCheckResourceAttr(dataSourceNameMatchBothTag, "arns.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchBothTag, "arns.*", resourceLb1, "arn"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchBothTag, "arns.*", resourceLb2, "arn"),
					resource.TestCheckResourceAttr(dataSourceNameMatchNoneTag, "arns.#", "0"),
				),
			},
		},
	})
}

func testAccLoadBalancersDataSourceConfig_basic(rName, lbName1, lbName2, tagValue string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = list(string)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  vpc_id            = aws_vpc.test.id
  cidr_block        = element(var.subnets, count.index)
  availability_zone = element(data.aws_availability_zones.available.names, count.index)
  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test1" {
  name               = %[2]q
  load_balancer_type = "application"
  internal           = true
  subnets            = aws_subnet.test[*].id
  tags = {
    "Name"               = %[2]q
    "TfTestingSharedTag" = %[4]q
  }
}

resource "aws_lb" "test2" {
  name               = %[3]q
  load_balancer_type = "application"
  internal           = true
  subnets            = aws_subnet.test[*].id
  tags = {
    "Name"               = %[3]q
    "TfTestingSharedTag" = %[4]q
  }
}

data "aws_lbs" "tag_match_first" {
  tags = {
    "Name" = %[2]q
  }
  depends_on = [aws_lb.test1, aws_lb.test2]
}

data "aws_lbs" "tag_match_shared" {
  tags = {
    "TfTestingSharedTag" = %[4]q
  }
  depends_on = [aws_lb.test1, aws_lb.test2]
}

data "aws_lbs" "tag_match_none" {
  tags = {
    "Unmatched" = "NotMatched"
  }
  depends_on = [aws_lb.test1, aws_lb.test2]
}
`, rName, lbName1, lbName2, tagValue))
}
