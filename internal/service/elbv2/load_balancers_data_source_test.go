// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2LoadBalancersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
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
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoadBalancerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancersDataSourceConfig_basic(rName, lbName1, lbName2, sharedTagVal),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceNameMatchFirstTag, "arns.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchFirstTag, "arns.*", resourceLb1, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceNameMatchBothTag, "arns.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchBothTag, "arns.*", resourceLb1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(dataSourceNameMatchBothTag, "arns.*", resourceLb2, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceNameMatchNoneTag, "arns.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccLoadBalancersDataSourceConfig_basic(rName, lbName1, lbName2, tagValue string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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
