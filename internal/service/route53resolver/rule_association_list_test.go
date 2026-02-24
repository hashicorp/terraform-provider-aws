// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

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
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverRuleAssociation_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_resolver_rule_association.test[0]"
	resourceName2 := "aws_route53_resolver_rule_association.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		CheckDestroy: testAccCheckRuleAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RuleAssociation/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"domain":         config.StringVariable(domainName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/RuleAssociation/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"domain":         config.StringVariable(domainName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_route53_resolver_rule_association.test", map[string]knownvalue.Check{
						names.AttrID:        id1.Value(),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
					querycheck.ExpectIdentity("aws_route53_resolver_rule_association.test", map[string]knownvalue.Check{
						names.AttrID:        id2.Value(),
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}
