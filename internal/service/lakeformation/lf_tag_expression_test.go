// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameLFTagExpression = "LF Tag Expression"
)

func testAccLFTagExpression_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var lftagexpression lakeformation.GetLFTagExpressionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, t, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "expression.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrName, names.AttrCatalogID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func testAccLFTagExpression_update(t *testing.T) {
	ctx := acctest.Context(t)

	var lftagexpression lakeformation.GetLFTagExpressionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, t, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "expression.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", names.AttrName, names.AttrCatalogID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccLFTagExpressionConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, t, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description two"),
					resource.TestCheckResourceAttr(resourceName, "expression.#", "2"),
				),
			},
		},
	})
}

func testAccLFTagExpression_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var lftagexpression lakeformation.GetLFTagExpressionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, t, resourceName, &lftagexpression),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflakeformation.ResourceLFTagExpression, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
func testAccCheckLFTagExpressionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_lf_tag_expression" {
				continue
			}

			_, err := tflakeformation.FindLFTagExpression(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrCatalogID])

			if retry.NotFound(err) {
				continue
			}

			if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "Insufficient Lake Formation permission(s)") {
				continue
			}

			if err != nil {
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, ResNameLFTagExpression, rs.Primary.ID, err)
			}

			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, ResNameLFTagExpression, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLFTagExpressionExists(ctx context.Context, t *testing.T, name string, lftagexpression *lakeformation.GetLFTagExpressionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)
		resp, err := tflakeformation.FindLFTagExpression(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrCatalogID])

		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, rs.Primary.ID, err)
		}

		*lftagexpression = *resp

		return nil
	}
}

func testAccLFTagExpressionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

	input := lakeformation.ListLFTagExpressionsInput{}
	_, err := conn.ListLFTagExpressions(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccLFTagExpression_baseConfig = `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = "key"
  values = ["value"]

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`

func testAccLFTagExpressionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLFTagExpression_baseConfig,
		fmt.Sprintf(`
resource "aws_lakeformation_lf_tag_expression" "test" {
  name        = %[1]q
  description = "test description"

  expression {
    tag_key    = aws_lakeformation_lf_tag.test.key
    tag_values = aws_lakeformation_lf_tag.test.values
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName))
}

func testAccLFTagExpressionConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccLFTagExpression_baseConfig,
		fmt.Sprintf(`
resource "aws_lakeformation_lf_tag" "test2" {
  key    = "key2"
  values = ["value2"]

  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag_expression" "test" {
  name        = %[1]q
  description = "test description two"

  expression {
    tag_key    = aws_lakeformation_lf_tag.test.key
    tag_values = aws_lakeformation_lf_tag.test.values
  }

  expression {
    tag_key    = aws_lakeformation_lf_tag.test2.key
    tag_values = aws_lakeformation_lf_tag.test2.values
  }

  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName))
}
