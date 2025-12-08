// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOutboundWebIdentityFederation_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccOutboundWebIdentityFederation_basic,
		acctest.CtDisappears: testAccOutboundWebIdentityFederation_disappears,
		"identity":           testAccIAMOutboundWebIdentityFederation_IdentitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOutboundWebIdentityFederation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_outbound_web_identity_federation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOutboundWebIdentityFederationExists(ctx, resourceName),
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
				ImportStateVerifyIdentifierAttribute: "issuer_identifier",
			},
		},
	})
}

func testAccOutboundWebIdentityFederation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_outbound_web_identity_federation.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOutboundWebIdentityFederationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfiam.ResourceOutboundWebIdentityFederation, resourceName),
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

func testAccCheckOutboundWebIdentityFederationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_outbound_web_identity_federation" {
				continue
			}

			_, err := tfiam.FindOutboundWebIdentityFederation(ctx, conn)

			if tfresource.NotFound(err) {
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

func testAccCheckOutboundWebIdentityFederationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, err := tfiam.FindOutboundWebIdentityFederation(ctx, conn)

		return err
	}
}

const testAccOutboundWebIdentityFederationConfig_basic = `
resource "aws_iam_outbound_web_identity_federation" "test" {}
`
