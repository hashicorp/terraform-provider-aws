// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaLayerVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:1", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "layer_name", rName),
					resource.TestCheckResourceAttr(resourceName, "license_info", ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "layer_arn", "lambda", fmt.Sprintf("layer:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "signing_profile_version_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "signing_job_arn", ""),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceLayerVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaLayerVersion_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_createBeforeDestroy(rName, "test-fixtures/lambdatest.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:1", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "source_code_hash", names.AttrSkipDestroy},
			},
			{
				Config: testAccLayerVersionConfig_createBeforeDestroy(rName, "test-fixtures/lambdatest_modified.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:2", rName)),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersion_sourceCodeHash(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_sourceCodeHash(rName, "test-fixtures/lambdatest.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:1", rName)),
				),
			},
			{
				Config: testAccLayerVersionConfig_sourceCodeHash(rName, "test-fixtures/lambdatest.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:1", rName)),
				),
			},
			{
				Config: testAccLayerVersionConfig_sourceCodeHash(rName, "test-fixtures/lambdatest_modified.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:2", rName)),
				),
			},
		},
	})
}

func TestAccLambdaLayerVersion_s3(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_s3(rName),
				Check:  testAccCheckLayerVersionExists(ctx, resourceName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrS3Bucket, "s3_key", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleRuntimes(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_compatibleRuntimes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", acctest.Ct2),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleArchitectures(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_compatibleArchitecturesNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", acctest.Ct0),
				),
			},
			{
				Config: testAccLayerVersionConfig_compatibleArchitecturesX86(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compatible_architectures.*", string(awstypes.ArchitectureX8664)),
				),
			},
			{
				Config: testAccLayerVersionConfig_compatibleArchitecturesArm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", acctest.Ct1),
				),
			},
			{
				Config: testAccLayerVersionConfig_compatibleArchitecturesX86Arm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", acctest.Ct2),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_description(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testDescription := "test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_description(rName, testDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, testDescription),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_licenseInfo(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testLicenseInfo := "MIT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLayerVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_licenseInfo(rName, testLicenseInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "license_info", testLicenseInfo),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", names.AttrSkipDestroy},
			},
		},
	})
}

func TestAccLambdaLayerVersion_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_layer_version.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop, // this purposely leaves dangling resources, since skip_destroy = true
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionConfig_skipDestroy(rName, "nodejs18.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:1", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
			{
				Config: testAccLayerVersionConfig_skipDestroy(rName, "nodejs20.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "lambda", fmt.Sprintf("layer:%s:2", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckLayerVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_layer_version" {
				continue
			}

			layerName, versionNumber, err := tflambda.LayerVersionParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tflambda.FindLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Layer Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLayerVersionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		layerName, versionNumber, err := tflambda.LayerVersionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		_, err = tflambda.FindLayerVersionByTwoPartKey(ctx, conn, layerName, versionNumber)

		return err
	}
}

func testAccLayerVersionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionConfig_s3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = %[1]q
}

resource "aws_s3_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.id
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_lambda_layer_version" "test" {
  s3_bucket  = aws_s3_bucket.lambda_bucket.id
  s3_key     = aws_s3_object.lambda_code.id
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionConfig_createBeforeDestroy(rName string, filename string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename         = %[1]q
  layer_name       = %[2]q
  source_code_hash = filebase64sha256(%[1]q)

  lifecycle {
    create_before_destroy = true
  }
}
`, filename, rName)
}

func testAccLayerVersionConfig_sourceCodeHash(rName string, filename string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename         = %[1]q
  layer_name       = %[2]q
  source_code_hash = filebase64sha256(%[1]q)
}
`, filename, rName)
}

func testAccLayerVersionConfig_compatibleRuntimes(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  compatible_runtimes = ["nodejs18.x", "nodejs20.x"]
}
`, rName)
}

func testAccLayerVersionConfig_compatibleArchitecturesNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionConfig_compatibleArchitecturesX86Arm(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["x86_64", "arm64"]
}
`, rName)
}

func testAccLayerVersionConfig_compatibleArchitecturesX86(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["x86_64"]
}
`, rName)
}

func testAccLayerVersionConfig_compatibleArchitecturesArm(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["arm64"]
}
`, rName)
}

func testAccLayerVersionConfig_description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  description = %[2]q
}
`, rName, description)
}

func testAccLayerVersionConfig_licenseInfo(rName string, licenseInfo string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  license_info = %[2]q
}
`, rName, licenseInfo)
}

func testAccLayerVersionConfig_skipDestroy(rName, compatRuntime string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = [%[2]q]
  skip_destroy        = true
}
`, rName, compatRuntime)
}
