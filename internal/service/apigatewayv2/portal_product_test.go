// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2PortalProduct_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductOutput
	resourceName := "aws_apigatewayv2_portal_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalProductConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_product_arn"), tfknownvalue.RegionalARNRegexp("apigateway", regexache.MustCompile(`/portalproducts/[0-9A-Za-z]{10,30}`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_product_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("last_modified"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_product_id"),
				ImportStateVerifyIdentifierAttribute: "portal_product_id",
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProduct_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductOutput
	resourceName := "aws_apigatewayv2_portal_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalProductConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfapigatewayv2.ResourcePortalProduct, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProduct_description(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductOutput
	resourceName := "aws_apigatewayv2_portal_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalProductConfig_description(rName, "original"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("original")),
				},
			},
			{
				Config: testAccPortalProductConfig_description(rName, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
				},
			},
			// Clearing is done by setting "" rather than removing the argument, because
			// UpdatePortalProduct is a PATCH and omits nil fields.
			{
				Config: testAccPortalProductConfig_description(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
					testAccCheckPortalProductDescription(&v, ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("")),
				},
			},
			// Removing the argument is a documented no-op: description is Optional+Computed,
			// so the prior value is retained and the plan stays empty.
			{
				Config:   testAccPortalProductConfig_basic(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProduct_displayName(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductOutput
	resourceName := "aws_apigatewayv2_portal_product.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalProductConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rName)),
				},
			},
			{
				Config: testAccPortalProductConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDisplayName), knownvalue.StringExact(rNameUpdated)),
					// Updating the display name must not replace the product.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_product_id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccCheckPortalProductDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_portal_product" {
				continue
			}

			_, err := tfapigatewayv2.FindPortalProductByID(ctx, conn, rs.Primary.Attributes["portal_product_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Portal Product %s still exists", rs.Primary.Attributes["portal_product_id"])
		}

		return nil
	}
}

func testAccCheckPortalProductExists(ctx context.Context, t *testing.T, n string, v *apigatewayv2.GetPortalProductOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		out, err := tfapigatewayv2.FindPortalProductByID(ctx, conn, rs.Primary.Attributes["portal_product_id"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccCheckPortalProductDescription(v *apigatewayv2.GetPortalProductOutput, want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if got := aws.ToString(v.Description); got != want {
			return fmt.Errorf("Description = %q, want %q", got, want)
		}

		return nil
	}
}

func testAccPortalProductConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_portal_product" "test" {
  display_name = %[1]q
}
`, rName)
}

func testAccPortalProductConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_portal_product" "test" {
  display_name = %[1]q
  description  = %[2]q
}
`, rName, description)
}
