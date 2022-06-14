package lambda_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
)

func TestAccLambdaLayerVersion_basic(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("layer:%s:1", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "layer_name", rName),
					resource.TestCheckResourceAttr(resourceName, "license_info", ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "layer_arn", "lambda", fmt.Sprintf("layer:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "signing_profile_version_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "signing_job_arn", ""),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_update(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCreateBeforeDestroy(rName, "test-fixtures/lambdatest.zip"),
				Check:  testAccCheckLayerVersionExists(resourceName, rName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "source_code_hash", "skip_destroy"},
			},

			{
				Config: testAccLayerVersionCreateBeforeDestroy(rName, "test-fixtures/lambdatest_modified.zip"),
				Check:  testAccCheckLayerVersionExists(resourceName, rName),
			},
		},
	})
}

func TestAccLambdaLayerVersion_s3(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionS3(rName),
				Check:  testAccCheckLayerVersionExists(resourceName, rName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleRuntimes(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCompatibleRuntimes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "2"),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleArchitectures(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCompatibleArchitecturesNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "0"),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesX86(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compatible_architectures.*", lambda.ArchitectureX8664),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesArm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "1"),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesX86Arm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "2"),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_description(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testDescription := "test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDescription(rName, testDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "description", testDescription),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_licenseInfo(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testLicenseInfo := "MIT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionLicenseInfo(rName, testLicenseInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "license_info", testLicenseInfo),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "skip_destroy"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_skipDestroy(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, // this purposely leaves dangling resources, since skip_destroy = true
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionSkipDestroyConfig(rName, "nodejs12.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("layer:%s:1", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
			{
				Config: testAccLayerVersionSkipDestroyConfig(rName, "nodejs14.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("layer:%s:2", rName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
		},
	})
}

func testAccCheckLayerVersionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version" {
			continue
		}

		layerName, version, err := tflambda.LayerVersionParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersion(&lambda.GetLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})
		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Lambda Layer Version (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckLayerVersionExists(res, layerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda Layer version not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda Layer ID not set")
		}

		if rs.Primary.Attributes["version"] == "" {
			return fmt.Errorf("Lambda Layer Version not set")
		}

		version, err := strconv.Atoi(rs.Primary.Attributes["version"])
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn
		_, err = conn.GetLayerVersion(&lambda.GetLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(int64(version)),
		})
		return err
	}
}

func testAccLayerVersionBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = %[1]q
}

resource "aws_s3_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.id
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_lambda_layer_version" "lambda_layer_test" {
  s3_bucket  = aws_s3_bucket.lambda_bucket.id
  s3_key     = aws_s3_object.lambda_code.id
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionCreateBeforeDestroy(rName string, filename string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename         = %[1]q
  layer_name       = %[2]q
  source_code_hash = filebase64sha256(%[1]q)

  lifecycle {
    create_before_destroy = true
  }
}
`, filename, rName)
}

func testAccLayerVersionCompatibleRuntimes(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  compatible_runtimes = ["nodejs12.x", "nodejs14.x"]
}
`, rName)
}

func testAccLayerVersionCompatibleArchitecturesNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}
`, rName)
}

func testAccLayerVersionCompatibleArchitecturesX86Arm(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["x86_64", "arm64"]
}
`, rName)
}

func testAccLayerVersionCompatibleArchitecturesX86(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["x86_64"]
}
`, rName)
}

func testAccLayerVersionCompatibleArchitecturesArm(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_architectures = ["arm64"]
}
`, rName)
}

func testAccLayerVersionDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  description = %[2]q
}
`, rName, description)
}

func testAccLayerVersionLicenseInfo(rName string, licenseInfo string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q

  license_info = %[2]q
}
`, rName, licenseInfo)
}

func testAccLayerVersionSkipDestroyConfig(rName, compatRuntime string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = %[1]q
  compatible_runtimes = [%[2]q]
  skip_destroy        = true
}
`, rName, compatRuntime)
}
