// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDefaultSecurityGroup_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_default_security_group.test[0]"
	resourceName2 := "aws_default_security_group.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_default_security_group.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_default_security_group.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^default \(sg-.+\)$`))),
					querycheck.ExpectResourceDisplayName("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^default \(sg-.+\)$`))),
				},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_default_security_group.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	vpcID := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					vpcID.GetStateValue("aws_vpc.test", tfjsonpath.New(names.AttrID)),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_include_resource"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_default_security_group.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+`-0 \(sg-.+\)$`))),
					querycheck.ExpectResourceKnownValues("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`security-group/sg-.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("default VPC security group")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("egress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("ingress"), knownvalue.SetExact([]knownvalue.Check{})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact("default")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrNamePrefix), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), vpcID.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name": knownvalue.StringExact(rName + "-0"),
						})),
					}),
				},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_default_security_group.test[0]"
	resourceName2 := "aws_default_security_group.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

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
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_region_override"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/DefaultSecurityGroup/list_region_override"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_default_security_group.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_default_security_group.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^default \(sg-.+\)$`))),
					querycheck.ExpectResourceDisplayName("aws_default_security_group.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^default \(sg-.+\)$`))),
				},
			},
		},
	})
}
