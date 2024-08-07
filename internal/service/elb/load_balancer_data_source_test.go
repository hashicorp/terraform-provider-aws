// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBLoadBalancerDataSource_basic(t *testing.T) {
	// Must be less than 32 characters for ELB name
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_elb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerDataSourceConfig_basic(rName, t.Name()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "cross_zone_load_balancing", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "internal", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "subnets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "security_groups.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "desync_mitigation_mode", "defensive"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.TestName", t.Name()),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrSet(dataSourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, "aws_elb.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccLoadBalancerDataSourceConfig_basic(rName, testName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout = 30

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name     = %[1]q
    TestName = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = "%[1]s"
  description = "%[2]s"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name     = %[1]q
    TestName = %[2]q
  }
}

data "aws_elb" "test" {
  name = aws_elb.test.name
}
`, rName, testName))
}
