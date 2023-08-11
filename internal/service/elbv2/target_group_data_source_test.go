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

func TestAccELBV2TargetGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceNameByARN := "data.aws_lb_target_group.alb_tg_test_with_arn"
	datasourceNameByName := "data.aws_lb_target_group.alb_tg_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByARN, "name", rName),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "arn"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "port", "8080"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "vpc_id"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "tags.Name", rName),
					resource.TestCheckResourceAttr(datasourceNameByARN, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.#", "1"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(datasourceNameByName, "name", rName),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "arn"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByName, "port", "8080"),
					resource.TestCheckResourceAttr(datasourceNameByName, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "vpc_id"),
					resource.TestCheckResourceAttr(datasourceNameByName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByName, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceNameByName, "tags.Name", rName),
					resource.TestCheckResourceAttr(datasourceNameByName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_appCookie(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceNameArn := "data.aws_lb_target_group.alb_tg_test_with_arn"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_appCookie(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, "name", rName),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "port", "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.0.cookie_duration", "600"),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.0.cookie_name", "cookieName"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceNameArn := "data.aws_alb_target_group.alb_tg_test_with_arn"
	resourceName := "data.aws_alb_target_group.alb_tg_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, "name", rName),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "port", "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "vpc_id"),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceTg1 := "aws_lb_target_group.test1"
	resourceTg2 := "aws_lb_target_group.test2"

	dataSourceMatchFirstTag := "data.aws_lb_target_group.tag_match_first"
	dataSourceMatchSecondTag := "data.aws_lb_target_group.tag_match_second"
	dataSourceMatchFirstTagAndName := "data.aws_lb_target_group.tag_and_arn_match_first"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elbv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_tags(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "name", resourceTg1, "name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "arn", resourceTg1, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "arn_suffix", resourceTg1, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "id", resourceTg1, "id"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "load_balancing_algorithm_type", resourceTg1, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "load_balancing_cross_zone_enabled", resourceTg1, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.#", resourceTg1, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.path", resourceTg1, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.port", resourceTg1, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.protocol", resourceTg1, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.healthy_threshold", resourceTg1, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.matcher", resourceTg1, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchFirstTag, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "tags.Name", resourceTg1, "tags.Name"),

					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "name", resourceTg2, "name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "arn", resourceTg2, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "arn_suffix", resourceTg2, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "id", resourceTg2, "id"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "load_balancing_algorithm_type", resourceTg2, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "load_balancing_cross_zone_enabled", resourceTg2, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.#", resourceTg2, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.path", resourceTg2, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.port", resourceTg2, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.protocol", resourceTg2, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.healthy_threshold", resourceTg2, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.matcher", resourceTg2, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchSecondTag, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "tags.Name", resourceTg2, "tags.Name"),

					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "name", resourceTg1, "name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "arn", resourceTg1, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "arn_suffix", resourceTg1, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "id", resourceTg1, "id"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "load_balancing_algorithm_type", resourceTg1, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "load_balancing_cross_zone_enabled", resourceTg1, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.#", resourceTg1, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.path", resourceTg1, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.port", resourceTg1, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.protocol", resourceTg1, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.healthy_threshold", resourceTg1, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.matcher", resourceTg1, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchFirstTagAndName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "tags.Name", resourceTg1, "tags.Name"),
				),
			},
		},
	})
}

func testAccTargetGroupDataSourceConfig_base(rName, resourceType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = %[2]s.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, resourceType))
}

func testAccTargetGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupDataSourceConfig_base(rName, "aws_lb_target_group"), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

data "aws_lb_target_group" "alb_tg_test_with_arn" {
  arn = aws_lb_target_group.test.arn
}

data "aws_lb_target_group" "alb_tg_test_with_name" {
  name = aws_lb_target_group.test.name
}
`, rName))
}

func testAccTargetGroupDataSourceConfig_appCookie(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupDataSourceConfig_base(rName, "aws_lb_target_group"), fmt.Sprintf(`
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

  stickiness {
    type            = "app_cookie"
    cookie_name     = "cookieName"
    cookie_duration = 600
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_lb_target_group" "alb_tg_test_with_arn" {
  arn = aws_lb_target_group.test.arn
}
`, rName))
}

func testAccTargetGroupDataSourceConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(testAccTargetGroupDataSourceConfig_base(rName, "aws_alb_target_group"), fmt.Sprintf(`
resource "aws_alb_target_group" "test" {
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

  tags = {
    Name = %[1]q
  }
}

data "aws_alb_target_group" "alb_tg_test_with_arn" {
  arn = aws_alb_target_group.test.arn
}

data "aws_alb_target_group" "alb_tg_test_with_name" {
  name = aws_alb_target_group.test.name
}
`, rName))
}

func testAccTargetGroupDataSourceConfig_tags(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_lb_target_group" "test1" {
  name        = %[1]q
  target_type = "lambda"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test2" {
  name        = %[2]q
  target_type = "lambda"

  tags = {
    Name = %[2]q
  }
}

data "aws_lb_target_group" "tag_match_first" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_lb_target_group.test1, aws_lb_target_group.test2]
}

data "aws_lb_target_group" "tag_match_second" {
  tags = {
    Name = %[2]q
  }

  depends_on = [aws_lb_target_group.test1, aws_lb_target_group.test2]
}

data "aws_lb_target_group" "tag_and_arn_match_first" {
  arn = aws_lb_target_group.test1.arn

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_lb_target_group.test1, aws_lb_target_group.test2]
}
`, rName1, rName2)
}
