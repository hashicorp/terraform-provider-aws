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

func TestAccELBV2LoadBalancerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	dataSourceName2 := "data.aws_lb.alb_test_with_name"
	dataSourceName3 := "data.aws_lb.alb_test_with_tags"
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enforce_security_group_inbound_rules_on_private_link_traffic", resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enforce_security_group_inbound_rules_on_private_link_traffic", resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName3, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName3, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "desync_mitigation_mode", resourceName, "desync_mitigation_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enforce_security_group_inbound_rules_on_private_link_traffic", resourceName, "enforce_security_group_inbound_rules_on_private_link_traffic"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_tls_version_and_cipher_suite_headers", resourceName, "enable_tls_version_and_cipher_suite_headers"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_xff_client_port", resourceName, "enable_xff_client_port"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "xff_header_processing_mode", resourceName, "xff_header_processing_mode"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancerDataSource_outpost(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	resourceName := "aws_lb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerDataSourceConfig_outpost(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.0.outpost_id", resourceName, "subnet_mapping.0.outpost_id"),
				),
			},
		},
	})
}

func TestAccELBV2LoadBalancerDataSource_backwardsCompatibility(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName1 := "data.aws_alb.alb_test_with_arn"
	dataSourceName2 := "data.aws_alb.alb_test_with_name"
	dataSourceName3 := "data.aws_alb.alb_test_with_tags"
	resourceName := "aws_alb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLoadBalancerDataSourceConfig_backwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName1, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName1, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName1, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "preserve_host_header", resourceName, "preserve_host_header"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "connection_logs.#", resourceName, "connection_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "preserve_host_header", resourceName, "preserve_host_header"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "connection_logs.#", resourceName, "connection_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName3, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Name", resourceName, "tags.Name"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "tags.Config", resourceName, "tags.Config"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(dataSourceName3, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName3, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSourceName3, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "preserve_host_header", resourceName, "preserve_host_header"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "enable_waf_fail_open", resourceName, "enable_waf_fail_open"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "connection_logs.#", resourceName, "connection_logs.#"),
				),
			},
		},
	})
}

func testAccLoadBalancerDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  desync_mitigation_mode = "defensive"

  tags = {
    Name   = %[1]q
    Config = "Basic"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

data "aws_lb" "alb_test_with_arn" {
  arn = aws_lb.test.arn
}

data "aws_lb" "alb_test_with_name" {
  name = aws_lb.test.name
}

data "aws_lb" "alb_test_with_tags" {
  tags = aws_lb.test.tags
}
`, rName))
}

func testAccLoadBalancerDataSourceConfig_outpost(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = [aws_subnet.test.id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name   = %[1]q
    Config = "Outposts"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

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

data "aws_lb" "alb_test_with_arn" {
  arn = aws_lb.test.arn
}
`, rName)
}

func testAccLoadBalancerDataSourceConfig_backwardsCompatibility(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_alb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    Name   = %[1]q
    Config = "BackwardsCompatibility"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

data "aws_alb" "alb_test_with_arn" {
  arn = aws_alb.test.arn
}

data "aws_alb" "alb_test_with_name" {
  name = aws_alb.test.name
}

data "aws_alb" "alb_test_with_tags" {
  tags = aws_alb.test.tags
}
`, rName))
}
