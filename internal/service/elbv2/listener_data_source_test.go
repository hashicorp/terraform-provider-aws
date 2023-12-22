// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccELBV2ListenerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener.test"
	dataSourceName := "data.aws_lb_listener.test"
	dataSourceName2 := "data.aws_alb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "alpn_policy", resourceName, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "certificate_arn", resourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.#", resourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "load_balancer_arn", resourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mutual_authentication.#", resourceName, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "protocol", resourceName, "protocol"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ssl_policy", resourceName, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "alpn_policy", dataSourceName, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "certificate_arn", dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "default_action.#", dataSourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "load_balancer_arn", dataSourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "mutual_authentication.#", dataSourceName, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "port", dataSourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "protocol", dataSourceName, "protocol"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ssl_policy", dataSourceName, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.%", dataSourceName, "tags.%"),
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
