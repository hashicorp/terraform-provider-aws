// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCSecurityGroupRulesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRulesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_security_group_rules.test", "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRulesDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRulesDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_security_group_rules.test", "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccVPCSecurityGroupRulesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}

data "aws_vpc_security_group_rules" "test" {
  filter {
    name   = "security-group-rule-id"
    values = [aws_vpc_security_group_ingress_rule.test.id]
  }
}
`)
}

func testAccVPCSecurityGroupRulesDataSourceConfig_tags(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_security_group_rules" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_security_group_ingress_rule.test, aws_vpc_security_group_egress_rule.test]
}
`, rName))
}
