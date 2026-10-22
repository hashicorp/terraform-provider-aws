// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkFunctionRevisionARN(functionName string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("lambda", regexache.MustCompile(`web-function/`+functionName+`/revision/rev-[0-9a-z]+`))
}

func TestAccLambdaWebFunctionRevision_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdaweb.GetWebFunctionRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionRevision/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionRevisionExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkFunctionRevisionARN(rName)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionRevision/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    testAccFunctionRevisionImportStateIDFunc(resourceName),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "revision_id",
			},
		},
	})
}

func TestAccLambdaWebFunctionRevision_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lambdaweb.GetWebFunctionRevisionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function_revision.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionRevisionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FunctionRevision/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionRevisionExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceFunctionRevision, resourceName),
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

func testAccCheckFunctionRevisionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdaweb_function_revision" {
				continue
			}

			_, err := tflambdaweb.FindWebFunctionRevisionByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["revision_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Web Function Revision %s still exists", rs.Primary.Attributes["revision_id"])
		}

		return nil
	}
}

func testAccCheckFunctionRevisionExists(ctx context.Context, t *testing.T, n string, v *lambdaweb.GetWebFunctionRevisionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		out, err := tflambdaweb.FindWebFunctionRevisionByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.Attributes["revision_id"])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFunctionRevisionImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(n, ",", "function_name", "revision_id")
}
