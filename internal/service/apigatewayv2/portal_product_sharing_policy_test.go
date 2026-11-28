// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2PortalProductSharingPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_document"), knownvalue.StringRegexp(regexache.MustCompile(`"apigateway:GetPortalProduct"`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_product_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_product_id"),
				ImportStateVerifyIdentifierAttribute: "portal_product_id",
				ImportStateVerifyIgnore:              []string{"policy_document"},
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProductSharingPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfapigatewayv2.ResourcePortalProductSharingPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProductSharingPolicy_Disappears_portalProduct(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfapigatewayv2.ResourcePortalProduct, "aws_apigatewayv2_portal_product.test"),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProductSharingPolicy_policyDocument(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	actions := func(v ...string) config.Variable {
		vars := make([]config.Variable, len(v))
		for i, s := range v {
			vars[i] = config.StringVariable(s)
		}
		return config.ListVariable(vars...)
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/actions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"actions":       actions("apigateway:GetPortalProduct"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/actions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"actions":       actions("apigateway:GetPortalProduct", "apigateway:ListProductPages", "apigateway:GetProductPage"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_document"), knownvalue.StringRegexp(regexache.MustCompile(`"apigateway:ListProductPages"`))),
				},
			},
			// Reordering the actions applies cleanly and leaves no diff behind.
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/actions/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"actions":       actions("apigateway:GetProductPage", "apigateway:GetPortalProduct", "apigateway:ListProductPages"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProductSharingPolicy_organization(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/organization/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_document"), knownvalue.StringRegexp(regexache.MustCompile(`"aws:PrincipalOrgID"`))),
				},
			},
			// Narrow the share to an organization path.
			{
				ConfigDirectory: config.StaticDirectory("testdata/PortalProductSharingPolicy/organizationPath/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_document"), knownvalue.StringRegexp(regexache.MustCompile(`"aws:PrincipalOrgPaths"`))),
				},
			},
		},
	})
}

func TestAccAPIGatewayV2PortalProductSharingPolicy_alternateAccount(t *testing.T) {
	ctx := acctest.Context(t)

	var v apigatewayv2.GetPortalProductSharingPolicyOutput
	resourceName := "aws_apigatewayv2_portal_product_sharing_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckPortalProductSharingPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortalProductSharingPolicyConfig_alternateAccount(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPortalProductSharingPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("portal_product_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "portal_product_id"),
				ImportStateVerifyIdentifierAttribute: "portal_product_id",
				ImportStateVerifyIgnore:              []string{"policy_document"},
			},
		},
	})
}

func testAccCheckPortalProductSharingPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_portal_product_sharing_policy" {
				continue
			}

			_, err := tfapigatewayv2.FindPortalProductSharingPolicyByID(ctx, conn, rs.Primary.Attributes["portal_product_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Portal Product Sharing Policy %s still exists", rs.Primary.Attributes["portal_product_id"])
		}

		return nil
	}
}

func testAccCheckPortalProductSharingPolicyExists(ctx context.Context, t *testing.T, n string, v *apigatewayv2.GetPortalProductSharingPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindPortalProductSharingPolicyByID(ctx, conn, rs.Primary.Attributes["portal_product_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPortalProductSharingPolicyConfig_alternateAccount(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
resource "aws_apigatewayv2_portal_product_sharing_policy" "test" {
  portal_product_id = aws_apigatewayv2_portal_product.test.portal_product_id

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.alternate.account_id}:root"
      }
      Action = [
        "apigateway:GetPortalProduct",
        "apigateway:ListProductPages",
        "apigateway:ListProductRestEndpointPages",
        "apigateway:GetProductPage",
        "apigateway:GetProductRestEndpointPage",
      ]
      Resource = [
        aws_apigatewayv2_portal_product.test.portal_product_arn,
        "${aws_apigatewayv2_portal_product.test.portal_product_arn}/*",
      ]
    }]
  })
}

resource "aws_apigatewayv2_portal_product" "test" {
  display_name = %[1]q
}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_partition" "current" {}
`, rName))
}
