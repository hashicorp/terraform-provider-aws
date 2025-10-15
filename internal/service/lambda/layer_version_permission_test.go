// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaLayerVersionPermission_basic_byARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	// Used to ensure the correct resource is read when other statements exist before and after.
	resourceNameFoo := "aws_lambda_layer_version_permission.foo"
	resourceNameBar := "aws_lambda_layer_version_permission.bar"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_basicARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					testAccCheckLayerVersionPermissionExists(ctx, resourceNameFoo),
					testAccCheckLayerVersionPermissionExists(ctx, resourceNameBar),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				/* Each permission resource keeps track of the overall policy and policy
				 * revision separately. This means that when another permission is added,
				 * these attributes become out of date in the state of prior existing
				 * permission resources. Therefore, for an accurate ImportStateVerify
				 * test of multiple permission resources, each one must be refreshed prior
				 * to the test. */
				RefreshState: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_basic_byName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_basicName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_org(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_org(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttr(resourceName, "organization_id", "o-0123456789"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_account(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceLayerVersionPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil, // this purposely leaves dangling resources, since skip_destroy = true
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionPermissionConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccLayerVersionPermissionConfig_skipDestroy(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

// Creating Lambda layer and Lambda layer permissions

func testAccLayerVersionPermissionConfig_basicARN(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}

resource "aws_lambda_layer_version_permission" "foo" {
  layer_name     = aws_lambda_layer_version.test.layer_arn
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "fooaccount"
  principal      = "*"
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name     = aws_lambda_layer_version.test.layer_arn
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "xaccount"
  principal      = "*"
  depends_on = [aws_lambda_layer_version_permission.foo]
}

resource "aws_lambda_layer_version_permission" "bar" {
  layer_name     = aws_lambda_layer_version.test.layer_arn
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "baraccount"
  principal      = "*"
  depends_on = [aws_lambda_layer_version_permission.test]
}
`, layerName)
}

func testAccLayerVersionPermissionConfig_basicName(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name     = aws_lambda_layer_version.test.layer_name
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "xaccount"
  principal      = "*"
}
`, layerName)
}

func testAccLayerVersionPermissionConfig_org(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name      = aws_lambda_layer_version.test.layer_arn
  version_number  = aws_lambda_layer_version.test.version
  action          = "lambda:GetLayerVersion"
  statement_id    = "xaccount"
  principal       = "*"
  organization_id = "o-0123456789"
}
`, layerName)
}

func testAccLayerVersionPermissionConfig_account(layerName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name     = aws_lambda_layer_version.test.layer_arn
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "xaccount"
  principal      = data.aws_caller_identity.current.account_id
}
`, layerName)
}

func testAccLayerVersionPermissionConfig_skipDestroy(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name     = aws_lambda_layer_version.test.layer_name
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "xaccount"
  principal      = "*"
  skip_destroy   = true
}
`, layerName)
}

func testAccCheckLayerVersionPermissionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		layerName, versionNumber, statementId, err := tflambda.LayerVersionPermissionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		_, _, err = tflambda.FindLayerVersionPermissionByThreePartKey(ctx, conn, layerName, versionNumber, statementId)

		return err
	}
}

func testAccCheckLayerVersionPermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_layer_version_permission" {
				continue
			}

			layerName, versionNumber, statementId, err := tflambda.LayerVersionPermissionParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, _, err = tflambda.FindLayerVersionPermissionByThreePartKey(ctx, conn, layerName, versionNumber, statementId)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Layer Version Permission %s still exists", rs.Primary.ID)
		}
		return nil
	}
}
