// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccApprovedOrigin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	origin := "https://example.com"
	resourceName := "aws_connect_approved_origin.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovedOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovedOriginConfig_basic(rName, origin),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovedOriginExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, "origin", origin),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccApprovedOrigin_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	origin := "https://example.com"
	resourceName := "aws_connect_approved_origin.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovedOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovedOriginConfig_basic(rName, origin),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovedOriginExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfconnect.ResourceApprovedOrigin, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccApprovedOrigin_import(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	origin := "https://example.com"
	resourceName := "aws_connect_approved_origin.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovedOriginDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovedOriginConfig_basic(rName, origin),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApprovedOriginExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "origin",
				ImportStateIdFunc:                    testAccApprovedOriginImportStateIDFunc(resourceName),
			},
		},
	})
}

func testAccCheckApprovedOriginDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_approved_origin" {
				continue
			}

			_, err := tfconnect.FindApprovedOriginByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["origin"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Approved Origin (%s/%s) still exists", rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["origin"])
		}

		return nil
	}
}

func testAccCheckApprovedOriginExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		_, err := tfconnect.FindApprovedOriginByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["origin"])

		return err
	}
}

func testAccApprovedOriginImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["origin"]), nil
	}
}

func testAccApprovedOriginConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccApprovedOriginConfig_basic(rName, origin string) string {
	return acctest.ConfigCompose(
		testAccApprovedOriginConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_approved_origin" "test" {
  instance_id = aws_connect_instance.test.id
  origin      = %[1]q
}
`, origin))
}
