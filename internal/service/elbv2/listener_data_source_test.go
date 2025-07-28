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
	dataSource1Name := "data.aws_lb_listener.test"
	dataSourceName2 := "data.aws_alb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "alpn_policy", resourceName, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrCertificateARN, resourceName, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(dataSource1Name, "default_action.#", resourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "default_action.0.target_group_arn", resourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "load_balancer_arn", resourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "mutual_authentication.#", resourceName, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrProtocol, resourceName, names.AttrProtocol),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ssl_policy", resourceName, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSource1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName2, "alpn_policy", dataSource1Name, "alpn_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrARN, dataSource1Name, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrCertificateARN, dataSource1Name, names.AttrCertificateARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, "default_action.#", dataSource1Name, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "default_action.0.target_group_arn", dataSource1Name, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "load_balancer_arn", dataSource1Name, "load_balancer_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "mutual_authentication.#", dataSource1Name, "mutual_authentication.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrPort, dataSource1Name, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrProtocol, dataSource1Name, names.AttrProtocol),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ssl_policy", dataSource1Name, "ssl_policy"),
					resource.TestCheckResourceAttrPair(dataSourceName2, acctest.CtTagsPercent, dataSource1Name, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccELBV2ListenerDataSource_mutualAuthentication(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	resourceName := "aws_lb_listener.test"
	dataSourceName := "data.aws_lb_listener.test"
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccListenerDataSourceConfig_mutualAuthentication(rName, key, certificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "mutual_authentication.#", resourceName, "mutual_authentication.#"),
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

func testAccListenerDataSourceConfig_mutualAuthentication(rName, key, certificate string) string {
	return acctest.ConfigCompose(testAccListenerConfig_mutualAuthentication(rName, key, certificate), `
data "aws_lb_listener" "test" {
  arn = aws_lb_listener.test.arn
}
`)
}
