// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2TargetGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceNameByARN := "data.aws_lb_target_group.alb_tg_test_with_arn"
	datasourceNameByName := "data.aws_lb_target_group.alb_tg_test_with_name"
	resourceName := "aws_lb_target_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(datasourceNameByARN, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceNameByARN, "arn_suffix", resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancing_algorithm_type", "round_robin"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancing_anomaly_mitigation", "off"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByARN, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttrPair(datasourceNameByARN, "target_control_port", resourceName, "target_control_port"),

					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(datasourceNameByName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceNameByName, "arn_suffix", resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancing_algorithm_type", "round_robin"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancing_anomaly_mitigation", "off"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByName, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByName, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttrPair(datasourceNameByName, "target_control_port", resourceName, "target_control_port"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_appCookie(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb_target_group.alb_tg_test_with_arn"
	resourceName := "aws_lb_target_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_appCookie(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn_suffix", resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(dataSourceName, "load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(dataSourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(dataSourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "stickiness.0.cookie_duration", "600"),
					resource.TestCheckResourceAttr(dataSourceName, "stickiness.0.cookie_name", "cookieName"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(dataSourceName, "health_check.0.matcher", "200-299"),
					resource.TestCheckResourceAttrPair(dataSourceName, "target_control_port", resourceName, "target_control_port"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	datasourceNameByARN := "data.aws_alb_target_group.alb_tg_test_with_arn"
	datasourceNameByName := "data.aws_alb_target_group.alb_tg_test_with_name"
	resourceName := "aws_alb_target_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(datasourceNameByARN, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceNameByARN, "arn_suffix", resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByARN, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByARN, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttrPair(datasourceNameByARN, "target_control_port", resourceName, "target_control_port"),

					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(datasourceNameByName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceNameByName, "arn_suffix", resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancer_arns.#", "0"),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByName, "slow_start", "0"),
					resource.TestCheckResourceAttr(datasourceNameByName, acctest.CtTagsPercent, "1"),
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
					resource.TestCheckResourceAttrPair(datasourceNameByName, "target_control_port", resourceName, "target_control_port"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_matchTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceTg1 := "aws_lb_target_group.test1"
	resourceTg2 := "aws_lb_target_group.test2"

	dataSourceMatchFirstTag := "data.aws_lb_target_group.tag_match_first"
	dataSourceMatchSecondTag := "data.aws_lb_target_group.tag_match_second"
	dataSourceMatchFirstTagAndName := "data.aws_lb_target_group.tag_and_arn_match_first"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_matchTags(rName1, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, names.AttrName, resourceTg1, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, names.AttrARN, resourceTg1, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "arn_suffix", resourceTg1, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "load_balancer_arns.#", resourceTg1, "load_balancer_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, names.AttrID, resourceTg1, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "load_balancing_algorithm_type", resourceTg1, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "load_balancing_cross_zone_enabled", resourceTg1, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.#", resourceTg1, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.path", resourceTg1, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.port", resourceTg1, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.protocol", resourceTg1, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.healthy_threshold", resourceTg1, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "health_check.0.matcher", resourceTg1, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchFirstTag, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "tags.Name", resourceTg1, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "target_control_port", resourceTg1, "target_control_port"),

					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, names.AttrName, resourceTg2, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, names.AttrARN, resourceTg2, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "arn_suffix", resourceTg2, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "load_balancer_arns.#", resourceTg2, "load_balancer_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, names.AttrID, resourceTg2, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "load_balancing_algorithm_type", resourceTg2, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "load_balancing_cross_zone_enabled", resourceTg2, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.#", resourceTg2, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.path", resourceTg2, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.port", resourceTg2, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.protocol", resourceTg2, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.healthy_threshold", resourceTg2, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "health_check.0.matcher", resourceTg2, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchSecondTag, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "tags.Name", resourceTg2, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "target_control_port", resourceTg2, "target_control_port"),

					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, names.AttrName, resourceTg1, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, names.AttrARN, resourceTg1, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "arn_suffix", resourceTg1, "arn_suffix"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "load_balancer_arns.#", resourceTg1, "load_balancer_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, names.AttrID, resourceTg1, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "load_balancing_algorithm_type", resourceTg1, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "load_balancing_cross_zone_enabled", resourceTg1, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.#", resourceTg1, "health_check.#"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.path", resourceTg1, "health_check.0.path"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.port", resourceTg1, "health_check.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.protocol", resourceTg1, "health_check.0.protocol"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.healthy_threshold", resourceTg1, "health_check.0.healthy_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "health_check.0.matcher", resourceTg1, "health_check.0.matcher"),
					resource.TestCheckResourceAttr(dataSourceMatchFirstTagAndName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "tags.Name", resourceTg1, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTagAndName, "target_control_port", resourceTg1, "target_control_port"),
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

func testAccTargetGroupDataSourceConfig_matchTags(rName1, rName2 string) string {
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
