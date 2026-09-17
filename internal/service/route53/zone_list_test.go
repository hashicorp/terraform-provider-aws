// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

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

func TestAccRoute53Zone_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_zone.test[0]"
	resourceName2 := "aws_route53_zone.test[1]"
	zoneName := acctest.RandomDomainName(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_basic/"),
				ConfigVariables: config.Variables{
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact("subdomain0."+zoneName)),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact("subdomain1."+zoneName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_basic/"),
				ConfigVariables: config.Variables{
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_zone.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("subdomain0."+zoneName)),
					tfquerycheck.ExpectNoResourceObject("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_route53_zone.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact("subdomain1."+zoneName)),
					tfquerycheck.ExpectNoResourceObject("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccRoute53Zone_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_route53_zone.test[0]"
	zoneName := acctest.RandomDomainName(t)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_include_resource/"),
				ConfigVariables: config.Variables{
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact("subdomain0."+zoneName)),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_include_resource/"),
				ConfigVariables: config.Variables{
					"zoneName":       config.StringVariable(zoneName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_zone.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("subdomain0."+zoneName)),
					querycheck.ExpectResourceKnownValues("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrComment), knownvalue.StringExact("Managed by Terraform")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("delegation_set_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("enable_accelerated_recovery"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact("subdomain0."+zoneName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("name_servers"), knownvalue.ListSizeExact(4)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("primary_name_server"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("vpc"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccRoute53Zone_List_privateZone(t *testing.T) {
	ctx := acctest.Context(t)

	publicResourceName := "aws_route53_zone.public"
	privateResourceName := "aws_route53_zone.private"
	zoneName := acctest.RandomDomainName(t)

	publicIdentity := tfstatecheck.Identity()
	privateIdentity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		CheckDestroy:             testAccCheckZoneDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_private_zone/"),
				ConfigVariables: config.Variables{
					"zoneName": config.StringVariable(zoneName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					publicIdentity.GetIdentity(publicResourceName),
					statecheck.ExpectKnownValue(publicResourceName, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(publicResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("public."+zoneName)),

					privateIdentity.GetIdentity(privateResourceName),
					statecheck.ExpectKnownValue(privateResourceName, tfjsonpath.New("zone_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(privateResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("private."+zoneName)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Zone/list_private_zone/"),
				ConfigVariables: config.Variables{
					"zoneName": config.StringVariable(zoneName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_route53_zone.test", privateIdentity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(privateIdentity.Checks()), knownvalue.StringExact("private."+zoneName)),
					tfquerycheck.ExpectNoResourceObject("aws_route53_zone.test", tfqueryfilter.ByResourceIdentityFunc(privateIdentity.Checks())),

					tfquerycheck.ExpectNoIdentityFunc("aws_route53_zone.test", publicIdentity.Checks()),
				},
			},
		},
	})
}
