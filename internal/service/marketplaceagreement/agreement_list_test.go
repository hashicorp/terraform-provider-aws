// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement_test

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

func testAccMarketplaceAgreementAgreement_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_marketplaceagreement_agreement.test"
	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MarketplaceAgreementServiceID),
		CheckDestroy:             testAccCheckAgreementDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/list_basic/"),
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/list_basic/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity1.Checks()),
				},
			},
		},
	})
}

func testAccMarketplaceAgreementAgreement_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_marketplaceagreement_agreement.test"
	identity1 := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MarketplaceAgreementServiceID),
		CheckDestroy:             testAccCheckAgreementDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/list_include_resource/"),
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/list_include_resource/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity1.Checks()),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("agreement_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("agreement_proposal_id"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("offer_id"), knownvalue.StringExact("bm7ut40zurhgdd0y3siqup0u6")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("requested_term"), knownvalue.SetSizeExact(2)),
					}),
				},
			},
		},
	})
}
