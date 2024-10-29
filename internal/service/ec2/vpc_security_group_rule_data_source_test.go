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

func TestAccVPCSecurityGroupRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_security_group_rule.test"
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_ipv4", resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_ipv6", resourceName, "cidr_ipv6"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "from_port", resourceName, "from_port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_protocol", resourceName, "ip_protocol"),
					resource.TestCheckResourceAttr(dataSourceName, "is_egress", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(dataSourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "referenced_security_group_id", resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_rule_id", resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "to_port", resourceName, "to_port"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRuleDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_security_group_rule.test"
	resourceName := "aws_vpc_security_group_egress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_ipv4", resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_ipv6", resourceName, "cidr_ipv6"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "from_port", resourceName, "from_port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_protocol", resourceName, "ip_protocol"),
					resource.TestCheckResourceAttr(dataSourceName, "is_egress", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(dataSourceName, "prefix_list_id", resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "referenced_security_group_id", resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_rule_id", resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "to_port", resourceName, "to_port"),
				),
			},
		},
	})
}

func testAccVPCSecurityGroupRuleDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}

data "aws_vpc_security_group_rule" "test" {
  security_group_rule_id = aws_vpc_security_group_ingress_rule.test.id
}
`)
}

func testAccVPCSecurityGroupRuleDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_egress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "2001:db8:85a3::/64"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_security_group_rule" "test" {
  filter {
    name   = "security-group-rule-id"
    values = [aws_vpc_security_group_egress_rule.test.id]
  }
}
`, rName))
}
