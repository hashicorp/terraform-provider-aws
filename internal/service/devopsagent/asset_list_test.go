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

func TestAccDevOpsAgentAsset_List_basic(t *testing.T) {
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
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_basic/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity("aws_devopsagent_asset.test[0]"),
					identity2.GetIdentity("aws_devopsagent_asset.test[1]"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_basic/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				// Agent Spaces include a sample skill, so assert test identities rather than total counts.
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), tfknownvalue.StringNotEmpty()),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), tfknownvalue.StringNotEmpty()),
					tfquerycheck.ExpectNoResourceObject("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),
					tfquerycheck.ExpectNoResourceObject("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_include_resource/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity("aws_devopsagent_asset.test"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_include_resource/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.test", identity.Checks()),
					querycheck.ExpectResourceDisplayName("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.StringNotEmpty()),
					querycheck.ExpectResourceKnownValues("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("agent_space_id"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("asset_id"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("asset_type"), knownvalue.StringExact("agents_md")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrContent), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrCreatedAt), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("metadata"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("metadata_all"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("updated_at"), tfknownvalue.StringNotEmpty()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrVersion), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_List_regionOverride(t *testing.T) {
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
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity("aws_devopsagent_asset.test[0]"),
					identity2.GetIdentity("aws_devopsagent_asset.test[1]"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.test", identity2.Checks()),
					querycheck.ExpectResourceKnownValues("aws_devopsagent_asset.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
					}),
				},
			},
		},
	})
}

func TestAccDevOpsAgentAsset_List_filters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_14_0)},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DevOpsAgentServiceID),
		CheckDestroy:             testAccCheckAssetDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_filters/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity("aws_devopsagent_asset.test"),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Asset/list_filters/"),
				ConfigVariables: config.Variables{acctest.CtRName: config.StringVariable(rName)},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_devopsagent_asset.included", identity.Checks()),
					// The skill filter can return the sample skill, but must exclude our agents_md asset.
					tfquerycheck.ExpectNoIdentityFunc("aws_devopsagent_asset.excluded_type", identity.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_devopsagent_asset.excluded_after", identity.Checks()),
					tfquerycheck.ExpectNoIdentityFunc("aws_devopsagent_asset.excluded_before", identity.Checks()),
				},
			},
		},
	})
}
