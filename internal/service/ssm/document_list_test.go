// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

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
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMDocument_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssm_document.test[0]"
	resourceName2 := "aws_ssm_document.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy:             testAccCheckDocumentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccSSMDocument_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssm_document.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy:             testAccCheckDocumentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-0")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_ssm_document.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("document_format"), knownvalue.StringExact("JSON")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("document_type"), knownvalue.StringExact("Command")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("Active")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
					}),
				},
			},
		},
	})
}

func TestAccSSMDocument_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_ssm_document.test[0]"
	resourceName2 := "aws_ssm_document.test[1]"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy:             acctest.CheckDestroyNoop,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("ssm", "document/"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNAlternateRegionExact("ssm", "document/"+rName+"-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccSSMDocument_List_filtered(t *testing.T) {
	ctx := acctest.Context(t)

	resourceNameExpected1 := "aws_ssm_document.expected[0]"
	resourceNameExpected2 := "aws_ssm_document.expected[1]"
	resourceNameNotExpected1 := "aws_ssm_document.not_expected[0]"
	resourceNameNotExpected2 := "aws_ssm_document.not_expected[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identityExpected1 := tfstatecheck.Identity()
	identityExpected2 := tfstatecheck.Identity()
	identityNotExpected1 := tfstatecheck.Identity()
	identityNotExpected2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy:             testAccCheckDocumentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identityExpected1.GetIdentity(resourceNameExpected1),
					statecheck.ExpectKnownValue(resourceNameExpected1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-expected-0")),

					identityExpected2.GetIdentity(resourceNameExpected2),
					statecheck.ExpectKnownValue(resourceNameExpected2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-expected-1")),

					identityNotExpected1.GetIdentity(resourceNameNotExpected1),
					statecheck.ExpectKnownValue(resourceNameNotExpected1, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-other-0")),

					identityNotExpected2.GetIdentity(resourceNameNotExpected2),
					statecheck.ExpectKnownValue(resourceNameNotExpected2, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("ssm", "document/"+rName+"-other-1")),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Document/list_filtered/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identityExpected1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_ssm_document.test", identityExpected2.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_ssm_document.test", identityNotExpected1.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_ssm_document.test", identityNotExpected2.Checks()),
				},
			},
		},
	})
}
