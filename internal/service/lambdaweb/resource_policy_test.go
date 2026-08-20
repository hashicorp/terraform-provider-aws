// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaWebResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdaweb.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_lambdaweb_function.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{names.AttrPolicy}, // Whitespace differences cause import comparison to fail.
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
			},
		},
	})
}

func TestAccLambdaWebResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdaweb.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceResourcePolicy, resourceName),
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

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *lambdaweb.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return smarterr.NewError(fmt.Errorf("Not found: %s", n))
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		out, err := tflambdaweb.FindResourcePolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
		if err != nil {
			return smarterr.NewError(err)
		}

		*v = *out

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdaweb_resource_policy" {
				continue
			}

			_, err := tflambdaweb.FindResourcePolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
			if errs.IsA[*retry.NotFoundError](err) {
				continue
			}
			if err != nil {
				return err
			}

			return smarterr.NewError(fmt.Errorf("Lambda Web Resource Policy %s still exists", rs.Primary.Attributes[names.AttrResourceARN]))
		}

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_basic(rName), `
resource "aws_lambdaweb_resource_policy" "test" {
  resource_arn = aws_lambdaweb_function.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowAccount"
        Effect    = "Allow"
        Principal = { AWS = data.aws_caller_identity.current.account_id }
        Action    = "lambda:InvokeWebFunction"
        Resource  = aws_lambdaweb_function.test.arn
      }
    ]
  })
}

data "aws_caller_identity" "current" {}
`)
}
