// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent_test

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

func TestAccDevOpsAgentAssetFile_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_basic/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity("aws_devopsagent_asset_file.test[0]"),
					identity2.GetIdentity("aws_devopsagent_asset_file.test[1]"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_basic/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				// The parent asset also contains SKILL.md, so assert test identities rather than total counts.
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset_file.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset_file.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact("references/runbook-0.md")),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact("references/runbook-1.md")),
					tfquerycheck.ExpectNoResourceObject("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectNoResourceObject("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_include_resource/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity("aws_devopsagent_asset_file.test"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_include_resource/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset_file.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), knownvalue.StringExact("references/runbook.md")),
					querycheck.ExpectResourceKnownValues("aws_devopsagent_asset_file.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("agent_space_id"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("asset_id"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrContent), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"bytes": knownvalue.Null(),
								"text":  knownvalue.StringExact("Check the runbook first."),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreatedAt), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("metadata"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("metadata_actual"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPath), knownvalue.StringExact("references/runbook.md")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("updated_at"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVersion), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccDevOpsAgentAssetFile_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		CheckDestroy:             testAccCheckAssetFileDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity("aws_devopsagent_asset_file.test[0]"),
					identity2.GetIdentity("aws_devopsagent_asset_file.test[1]"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AssetFile/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset_file.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset_file.test", identity2.Checks()),
				},
			},
		},
	})
}
