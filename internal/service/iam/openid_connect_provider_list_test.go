// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
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

func TestAccIAMOpenIDConnectProvider_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_iam_openid_connect_provider.test[0]"
	resourceName2 := "aws_iam_openid_connect_provider.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OpenIDConnectProvider/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", fmt.Sprintf("oidc-provider/accounts.testle.com/%s-0", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", fmt.Sprintf("oidc-provider/accounts.testle.com/%s-1", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/OpenIDConnectProvider/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_openid_connect_provider.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_openid_connect_provider.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(fmt.Sprintf("accounts.testle.com/%s-0", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names
					tfquerycheck.ExpectNoResourceObject("aws_iam_openid_connect_provider.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_iam_openid_connect_provider.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_iam_openid_connect_provider.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(fmt.Sprintf("accounts.testle.com/%s-1", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names
					tfquerycheck.ExpectNoResourceObject("aws_iam_openid_connect_provider.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccIAMOpenIDConnectProvider_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_iam_openid_connect_provider.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		CheckDestroy:             testAccCheckOpenIDConnectProviderDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OpenIDConnectProvider/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/OpenIDConnectProvider/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_iam_openid_connect_provider.test", identity.Checks()),
					querycheck.ExpectResourceKnownValues("aws_iam_openid_connect_provider.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNExact("iam", fmt.Sprintf("oidc-provider/accounts.testle.com/%s-0", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrID), tfknownvalue.GlobalARNExact("iam", fmt.Sprintf("oidc-provider/accounts.testle.com/%s-0", rName))),  // nosemgrep:ci.semgrep.domain-names.domain-names
						tfquerycheck.KnownValueCheck(tfjsonpath.New("client_id_list"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"), // nosemgrep:ci.semgrep.domain-names.domain-names
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("thumbprint_list"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.StringExact("cf23df2207d99a74fbe169e3eba035e633b65d94"),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrURL), knownvalue.StringExact(fmt.Sprintf("accounts.testle.com/%s-0", rName))), // nosemgrep:ci.semgrep.domain-names.domain-names
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
