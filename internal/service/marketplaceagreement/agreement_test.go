// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package marketplaceagreement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/marketplaceagreement"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmarketplaceagreement "github.com/hashicorp/terraform-provider-aws/internal/service/marketplaceagreement"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMarketplaceAgreementAgreement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v marketplaceagreement.DescribeAgreementOutput
	resourceName := "aws_marketplaceagreement_agreement.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MarketplaceAgreementServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/basic/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgreementExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("acceptance_time"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agreement_id"), knownvalue.StringRegexp(regexache.MustCompile(`^agmt-[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agreement_proposal_id"), knownvalue.StringExact("at-32xgdgxyhupql1jw55dnk7ff1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agreement_type"), knownvalue.StringExact("PurchaseAgreement")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("end_time"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("offer_id"), knownvalue.StringExact("bm7ut40zurhgdd0y3siqup0u6")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("proposer_account_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("requested_term"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrConfiguration: knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrID:            knownvalue.StringExact("term-7e5d56619c3a08143428c3771d7f6de4bf687508be034e6c4b4dd26c1c1d966e"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrConfiguration: knownvalue.ListExact([]knownvalue.Check{}),
							names.AttrID:            knownvalue.StringExact("term-f5bdd147164b3f3b593e9b56a886c9e0fa3cd4fd9e7cf832ab13ef88161c14c8"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStartTime), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Agreement/basic/"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "agreement_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "agreement_id",
				ImportStateVerifyIgnore: []string{
					"agreement_proposal_id",
				},
			},
		},
	})
}

func testAccMarketplaceAgreementAgreement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v marketplaceagreement.DescribeAgreementOutput
	resourceName := "aws_marketplaceagreement_agreement.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MarketplaceAgreementServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgreementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Agreement/basic/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAgreementExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmarketplaceagreement.ResourceAgreement, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAgreementDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MarketplaceAgreementClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_marketplaceagreement_agreement" {
				continue
			}

			_, err := tfmarketplaceagreement.FindAgreementByID(ctx, conn, rs.Primary.Attributes["agreement_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Marketplace Agreement Agreement %s still exists", rs.Primary.Attributes["agreement_id"])
		}

		return nil
	}
}

func testAccCheckAgreementExists(ctx context.Context, t *testing.T, n string, v *marketplaceagreement.DescribeAgreementOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).MarketplaceAgreementClient(ctx)

		output, err := tfmarketplaceagreement.FindAgreementByID(ctx, conn, rs.Primary.Attributes["agreement_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
