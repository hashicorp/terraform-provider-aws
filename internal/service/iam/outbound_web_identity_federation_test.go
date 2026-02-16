// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOutboundWebIdentityFederation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccOutboundWebIdentityFederation_basic,
		acctest.CtDisappears: testAccOutboundWebIdentityFederation_disappears,
		"Identity":           testAccIAMOutboundWebIdentityFederation_identitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOutboundWebIdentityFederation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_outbound_web_identity_federation.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOutboundWebIdentityFederationExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("issuer_identifier"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIDAccountID(resourceName),
				ImportStateVerifyIdentifierAttribute: "issuer_identifier",
			},
		},
	})
}

func testAccOutboundWebIdentityFederation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_outbound_web_identity_federation.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOutboundWebIdentityFederationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfiam.ResourceOutboundWebIdentityFederation, resourceName),
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

func testAccCheckOutboundWebIdentityFederationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_outbound_web_identity_federation" {
				continue
			}

			_, err := tfiam.FindOutboundWebIdentityFederation(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("IAM Outbound Web Identity Federation still exists")
		}

		return nil
	}
}

func testAccCheckOutboundWebIdentityFederationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := tfiam.FindOutboundWebIdentityFederation(ctx, conn)

		return err
	}
}

func importStateIDAccountID(_ string) resource.ImportStateIdFunc {
	return acctest.ImportStateIDAccountID(context.Background())
}

const testAccOutboundWebIdentityFederationConfig_basic = `
resource "aws_iam_outbound_web_identity_federation" "test" {}
`
