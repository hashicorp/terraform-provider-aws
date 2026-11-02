// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCNetworkACLRule_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_acl_rule.test[0]"
	resourceName2 := "aws_network_acl_rule.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("network_acl_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("egress")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("rule_number")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrProtocol)),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("network_acl_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("egress")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("rule_number")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrProtocol)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_acl_rule.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_network_acl_rule.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
					tfquerycheck.ExpectNoResourceObject("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCNetworkACLRule_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_acl_rule.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()
	naclID := tfstatecheck.StateValue()
	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					naclID.GetStateValue("aws_network_acl.test", tfjsonpath.New(names.AttrID)),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("network_acl_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("egress")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("rule_number")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrProtocol)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					// IncludeResource is true, so all resource attributes should be populated.
					tfquerycheck.ExpectIdentityFunc("aws_network_acl_rule.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),
					querycheck.ExpectResourceKnownValues("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("network_acl_id"), naclID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("egress"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("rule_number"), knownvalue.Int64Exact(200)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrProtocol), knownvalue.StringExact("6")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("rule_action"), knownvalue.StringExact("allow")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCIDRBlock), knownvalue.StringExact("0.0.0.0/0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ipv6_cidr_block"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("from_port"), knownvalue.Int64Exact(22)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("to_port"), knownvalue.Int64Exact(22)),
					}),
				},
			},
		},
	})
}

func TestAccVPCNetworkACLRule_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_network_acl_rule.test[0]"
	resourceName2 := "aws_network_acl_rule.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	id1 := tfstatecheck.StateValue()
	id2 := tfstatecheck.StateValue()

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckNetworkACLRuleDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("network_acl_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("egress")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New("rule_number")),
					statecheck.ExpectIdentityValueMatchesState(resourceName1, tfjsonpath.New(names.AttrProtocol)),

					identity2.GetIdentity(resourceName2),
					id2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("network_acl_id")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("egress")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New("rule_number")),
					statecheck.ExpectIdentityValueMatchesState(resourceName2, tfjsonpath.New(names.AttrProtocol)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/NetworkACLRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_network_acl_rule.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), id1.ValueCheck()),

					tfquerycheck.ExpectIdentityFunc("aws_network_acl_rule.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_network_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), id2.ValueCheck()),
				},
			},
		},
	})
}
