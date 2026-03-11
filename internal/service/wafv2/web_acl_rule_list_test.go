// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2WebACLRule_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_wafv2_web_acl_rule.test[0]"
	resourceName2 := "aws_wafv2_web_acl_rule.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	webACLARN := tfstatecheck.StateValue()
	name1 := tfstatecheck.StateValue()
	name2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					webACLARN.GetStateValue("aws_wafv2_web_acl.test", tfjsonpath.New(names.AttrARN)),
					name1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrName)),
					name2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrName)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_wafv2_web_acl_rule.test", map[string]knownvalue.Check{
						"web_acl_arn": webACLARN.Value(),
						names.AttrName: name1.Value(),
					}),
					querycheck.ExpectIdentity("aws_wafv2_web_acl_rule.test", map[string]knownvalue.Check{
						"web_acl_arn": webACLARN.Value(),
						names.AttrName: name2.Value(),
					}),
				},
			},
		},
	})
}
