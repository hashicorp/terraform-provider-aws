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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCInternetGateway_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_internet_gateway.test[0]"
	resourceName2 := "aws_internet_gateway.test[1]"

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckInternetGatewayDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_basic/"),
				ConfigVariables: config.Variables{
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
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_basic/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity1.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_internet_gateway.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity2.Checks()),
					tfquerycheck.ExpectNoResourceObject("aws_internet_gateway.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccVPCInternetGateway_List_filter(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_internet_gateway.expected[0]"
	resourceName2 := "aws_internet_gateway.expected[1]"
	resourceName3 := "aws_internet_gateway.other"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	identity3 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckInternetGatewayDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_filter/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
					identity3.GetIdentity(resourceName3),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_filter/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccVPCInternetGateway_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_internet_gateway.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		CheckDestroy:             testAccCheckInternetGatewayDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_internet_gateway.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile("^"+rName+"-0 \\(igw-[a-z0-9]+\\)$"))),
					querycheck.ExpectResourceKnownValues("aws_internet_gateway.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), knownvalue.StringRegexp(regexache.MustCompile(`^arn:[^:]+:ec2:[^:]+:[0-9]{12}:internet-gateway/igw-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^igw-[a-z0-9]+$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrOwnerID), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9]{12}$`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVPCID), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName + "-0"),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							"Name":         knownvalue.StringExact(rName + "-0"),
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccVPCInternetGateway_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_internet_gateway.test[0]"
	resourceName2 := "aws_internet_gateway.test[1]"

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
		CheckDestroy:             testAccCheckInternetGatewayDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_region_override/"),
				ConfigVariables: config.Variables{
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
				ConfigDirectory: config.StaticDirectory("testdata/InternetGateway/list_region_override/"),
				ConfigVariables: config.Variables{
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_internet_gateway.test", identity2.Checks()),
				},
			},
		},
	})
}
