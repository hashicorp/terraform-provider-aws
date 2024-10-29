// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallResourcePolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`"Action":\["network-firewall:AssociateFirewallPolicy","network-firewall:ListFirewallPolicies"\]`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_networkfirewall_firewall_policy.test", names.AttrARN),
				),
			},
		},
	})
}

func testAccResourcePolicyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_networkfirewall_resource_policy" "test" {
  resource_arn = aws_networkfirewall_resource_policy.test.resource_arn
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_resource_policy" "test" {
  resource_arn = aws_networkfirewall_firewall_policy.test.arn

  policy = jsonencode({
    Statement = [{
      Action = [
        "network-firewall:AssociateFirewallPolicy",
        "network-firewall:ListFirewallPolicies",
      ]
      Effect   = "Allow"
      Resource = aws_networkfirewall_firewall_policy.test.arn
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}
