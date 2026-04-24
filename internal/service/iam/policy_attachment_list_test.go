// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
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

func TestAccIAMPolicyAttachment_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_policy_attachment.test[0]"
	resourceName2 := "aws_iam_policy_attachment.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/PolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("policy_arn"), "aws_iam_policy.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-1")),
					statecheck.CompareValuePairs(resourceName2, tfjsonpath.New("policy_arn"), "aws_iam_policy.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/PolicyAttachment/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_policy_attachment.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_iam_policy_attachment.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccIAMPolicyAttachment_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_policy_attachment.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	policyARN1 := tfstatecheck.StateValue()
	id1 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/PolicyAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					policyARN1.GetStateValue(resourceName1, tfjsonpath.New("policy_arn")),
					id1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrID)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
					statecheck.CompareValuePairs(resourceName1, tfjsonpath.New("policy_arn"), "aws_iam_policy.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/PolicyAttachment/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_policy_attachment.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_iam_policy_attachment.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("policy_arn"), policyARN1.ValueCheck()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("groups"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(rName + "-0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("roles"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(rName + "-0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("users"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(rName + "-0"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), id1.ValueCheck()),
					}),
				},
			},
		},
	})
}
