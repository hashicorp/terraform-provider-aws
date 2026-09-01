// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2ManagedRuleGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_wafv2_managed_rule_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccManagedRuleGroupDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
			{
				Config: testAccManagedRuleGroupDataSourceConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New("capacity"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(datasourceName, tfjsonpath.New(names.AttrSNSTopicARN), knownvalue.NotNull()),
				},
			},
		},
	})
}

const (
	testAccManagedRuleGroupDataSourceConfig_nonExistent = `
data "aws_wafv2_managed_rule_group" "test" {
  name        = "tf-acc-test-does-not-exist"
  scope       = "REGIONAL"
  vendor_name = "AWS"
}
`

	testAccManagedRuleGroupDataSourceConfig_basic = `
data "aws_wafv2_managed_rule_group" "test" {
  name        = "AWSManagedRulesCommonRuleSet"
  scope       = "REGIONAL"
  vendor_name = "AWS"
}
`
)
