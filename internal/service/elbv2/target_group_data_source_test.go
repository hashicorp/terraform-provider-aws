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

func TestAccELBV2TargetGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceNameByARN := "data.aws_lb_target_group.alb_tg_test_with_arn"
	datasourceNameByName := "data.aws_lb_target_group.alb_tg_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, names.AttrARN),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancer_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByARN, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancing_algorithm_type", "round_robin"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "load_balancing_anomaly_mitigation", "off"),
					resource.TestCheckResourceAttrSet(datasourceNameByARN, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "slow_start", acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceNameByARN, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByARN, "tags.Name", rName),
					resource.TestCheckResourceAttr(datasourceNameByARN, "stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.timeout", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.unhealthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceNameByARN, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(datasourceNameByName, names.AttrARN),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "arn_suffix"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancer_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(datasourceNameByName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancing_algorithm_type", "round_robin"),
					resource.TestCheckResourceAttr(datasourceNameByName, "load_balancing_anomaly_mitigation", "off"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttrSet(datasourceNameByName, names.AttrVPCID),
					resource.TestCheckResourceAttr(datasourceNameByName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(datasourceNameByName, "slow_start", acctest.Ct0),
					resource.TestCheckResourceAttr(datasourceNameByName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByName, "tags.Name", rName),
					resource.TestCheckResourceAttr(datasourceNameByName, "stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.timeout", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(datasourceNameByName, "health_check.0.unhealthy_threshold", acctest.Ct3),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_appCookie(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceNameArn, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "load_balancer_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol_version", "HTTP1"),
					resource.TestCheckResourceAttrSet(resourceNameArn, names.AttrVPCID),
					resource.TestCheckResourceAttrSet(resourceNameArn, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "load_balancing_cross_zone_enabled"),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceNameArn, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.0.cookie_duration", "600"),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.0.cookie_name", "cookieName"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", acctest.Ct3),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTargetGroupDataSourceConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceNameArn, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "load_balancer_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrSet(resourceNameArn, names.AttrVPCID),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceNameArn, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "8080"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "HTTP"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func TestAccELBV2TargetGroupDataSource_matchTags(t *testing.T) {
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
					resource.TestCheckResourceAttr(dataSourceMatchFirstTag, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceMatchFirstTag, "tags.Name", resourceTg1, "tags.Name"),

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
					resource.TestCheckResourceAttr(dataSourceMatchSecondTag, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceMatchSecondTag, "tags.Name", resourceTg2, "tags.Name"),

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
					resource.TestCheckResourceAttr(dataSourceMatchFirstTagAndName, acctest.CtTagsPercent, acctest.Ct1),
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
