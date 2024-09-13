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

func TestAccELBV2ListenerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"
	dataSourceName := "data.aws_lb_listener.test"
	dataSourceName2 := "data.aws_alb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "alpn_policy", resourceName, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCertificateARN, resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.#", resourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.0.target_group_arn", resourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "load_balancer_arn", resourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mutual_authentication.#", resourceName, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrProtocol, resourceName, names.AttrProtocol),
					resource.TestCheckResourceAttrPair(dataSourceName, "ssl_policy", resourceName, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName2, "alpn_policy", dataSourceName, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrCertificateARN, dataSourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, "default_action.#", dataSourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "default_action.0.target_group_arn", dataSourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "load_balancer_arn", dataSourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "mutual_authentication.#", dataSourceName, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrPort, dataSourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrProtocol, dataSourceName, names.AttrProtocol),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ssl_policy", dataSourceName, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccListenerDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccListenerConfig_base(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

data "aws_lb_listener" "test" {
  arn = aws_lb_listener.test.arn
}

data "aws_alb_listener" "from_lb_and_port" {
  load_balancer_arn = aws_lb.test.arn
  port              = aws_lb_listener.test.port
}
`, rName))
}
