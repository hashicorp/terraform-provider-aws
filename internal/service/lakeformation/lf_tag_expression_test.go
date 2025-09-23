// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.domain.#", "1"),
				),
			},
			{
				// Remove LF Tag Expression but keep Data Lake Settings to verify destruction with proper permissions
				Config: testAccLFTagExpressionConfig_onlyDataLakeSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					// Verify LF Tag Expression is destroyed while admin permissions still exist
					testAccCheckLFTagExpressionDestroy(ctx),
				),
			},
		},
	})
}

func testAccLFTagExpression_update(t *testing.T) {
	ctx := acctest.Context(t)

	var lftagexpression lakeformation.GetLFTagExpressionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_update1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Initial description"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.domain.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.domain.*", "finance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.domain.*", "hr"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.environment.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.environment.*", "dev"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.environment.*", "staging"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.environment.*", "prod"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.team.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.team.*", "data-eng"),
				),
			},
			{
				Config: testAccLFTagExpressionConfig_update2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, resourceName, &lftagexpression),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					// Verify tag_expression changes: removed 'team', added 'project', modified 'domain' and 'environment'
					resource.TestCheckResourceAttr(resourceName, "tag_expression.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.domain.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.domain.*", "finance"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.domain.*", "marketing"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.domain.*", "operations"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.environment.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.environment.*", "prod"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.environment.*", "test"),
					resource.TestCheckResourceAttr(resourceName, "tag_expression.project.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.project.*", "alpha"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tag_expression.project.*", "beta"),
				),
			},
			{
				// Remove LF Tag Expression but keep Data Lake Settings to verify destruction with proper permissions
				Config: testAccLFTagExpressionConfig_updateOnlyDataLakeSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					// Verify LF Tag Expression is destroyed while admin permissions still exist
					testAccCheckLFTagExpressionDestroy(ctx),
				),
			},
		},
	})
}

func testAccLFTagExpression_import(t *testing.T) {
	ctx := acctest.Context(t)

	var lftagexpression lakeformation.GetLFTagExpressionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lakeformation_lf_tag_expression.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormation)
			testAccLFTagExpressionPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagExpressionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagExpressionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExpressionExists(ctx, resourceName, &lftagexpression),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Remove LF Tag Expression but keep Data Lake Settings to verify destruction with proper permissions
				Config: testAccLFTagExpressionConfig_onlyDataLakeSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					// Verify LF Tag Expression is destroyed while admin permissions still exist
					testAccCheckLFTagExpressionDestroy(ctx),
				),
			},
		},
	})
}

func testAccCheckLFTagExpressionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_lf_tag_expression" {
				continue
			}

			_, err := tflakeformation.FindLFTagExpression(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrCatalogID])

			if retry.NotFound(err) {
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

func testAccCheckLFTagExpressionExists(ctx context.Context, name string, lftagexpression *lakeformation.GetLFTagExpressionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)
		resp, err := tflakeformation.FindLFTagExpression(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrCatalogID])

		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, ResNameLFTagExpression, rs.Primary.ID, err)
		}

		*lftagexpression = *resp

		return nil
	}
}

func testAccLFTagExpressionPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationClient(ctx)

	input := lakeformation.ListLFTagExpressionsInput{}
	_, err := conn.ListLFTagExpressions(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccLFTagExpressionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "domain" {
  key    = "domain"
  values = ["prisons"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag_expression" "test" {
  name = %[1]q
  
  tag_expression = {
    domain = ["prisons"]
  }

  depends_on = [
    aws_lakeformation_lf_tag.domain,
    aws_lakeformation_data_lake_settings.test
  ]
}
`, rName)
}

func testAccLFTagExpressionConfig_onlyDataLakeSettings(rName string) string {
	return `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "domain" {
  key    = "domain"
  values = ["prisons"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`
}

func testAccLFTagExpressionConfig_update1(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "domain" {
  key    = "domain"
  values = ["finance", "hr", "marketing", "operations"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "environment" {
  key    = "environment"
  values = ["dev", "staging", "prod", "test"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "team" {
  key    = "team"
  values = ["data-eng"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "project" {
  key    = "project"
  values = ["alpha", "beta"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag_expression" "test" {
  name        = %[1]q
  description = "Initial description"
  
  tag_expression = {
    domain      = ["finance", "hr"]
    environment = ["dev", "staging", "prod"]
    team        = ["data-eng"]
  }

  depends_on = [
    aws_lakeformation_lf_tag.domain,
    aws_lakeformation_lf_tag.environment,
    aws_lakeformation_lf_tag.team,
    aws_lakeformation_lf_tag.project,
    aws_lakeformation_data_lake_settings.test,
  ]
}
`, rName)
}

func testAccLFTagExpressionConfig_update2(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "domain" {
  key    = "domain"
  values = ["finance", "hr", "marketing", "operations"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "environment" {
  key    = "environment"
  values = ["dev", "staging", "prod", "test"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "team" {
  key    = "team"
  values = ["data-eng"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "project" {
  key    = "project"
  values = ["alpha", "beta"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag_expression" "test" {
  name        = %[1]q
  description = "Updated description"
  
  tag_expression = {
    domain      = ["finance", "marketing", "operations"]
    environment = ["prod", "test"]
    project     = ["alpha", "beta"]
  }

  depends_on = [
    aws_lakeformation_lf_tag.domain,
    aws_lakeformation_lf_tag.environment,
    aws_lakeformation_lf_tag.team,
    aws_lakeformation_lf_tag.project,
    aws_lakeformation_data_lake_settings.test,
  ]
}
`, rName)
}

func testAccLFTagExpressionConfig_updateOnlyDataLakeSettings(rName string) string {
	return `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "domain" {
  key    = "domain"
  values = ["finance", "hr", "marketing", "operations"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "environment" {
  key    = "environment"
  values = ["dev", "staging", "prod", "test"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "team" {
  key    = "team"
  values = ["data-eng"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_lf_tag" "project" {
  key    = "project"
  values = ["alpha", "beta"]
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`
}
