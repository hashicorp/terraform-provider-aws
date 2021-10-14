package lambda_test

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccLambdaLayerVersion_basic(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	layerName := fmt.Sprintf("tf_acc_lambda_layer_basic_%s", sdkacctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionBasic(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("layer:%s:1", layerName)),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "layer_name", layerName),
					resource.TestCheckResourceAttr(resourceName, "license_info", ""),
					acctest.CheckResourceAttrRegionalARN(resourceName, "layer_arn", "lambda", fmt.Sprintf("layer:%s", layerName)),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "signing_profile_version_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "signing_job_arn", ""),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_update(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	layerName := fmt.Sprintf("tf_acc_lambda_layer_basic_%s", sdkacctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCreateBeforeDestroy(layerName, "test-fixtures/lambdatest.zip"),
				Check:  testAccCheckLayerVersionExists(resourceName, layerName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "source_code_hash"},
			},

			{
				Config: testAccLayerVersionCreateBeforeDestroy(layerName, "test-fixtures/lambdatest_modified.zip"),
				Check:  testAccCheckLayerVersionExists(resourceName, layerName),
			},
		},
	})
}

func TestAccLambdaLayerVersion_s3(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := sdkacctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_s3_%s", rString)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-layer-s3-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionS3(bucketName, layerName),
				Check:  testAccCheckLayerVersionExists(resourceName, layerName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleRuntimes(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := sdkacctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_runtimes_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCompatibleRuntimes(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_runtimes.#", "2"),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_compatibleArchitectures(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := sdkacctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_architectures_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionCompatibleArchitecturesNone(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "0"),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesX86(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compatible_architectures.*", lambda.ArchitectureX8664),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesArm(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "1"),
				),
			},
			{
				Config: testAccLayerVersionCompatibleArchitecturesX86Arm(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "compatible_architectures.#", "2"),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_description(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := sdkacctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_description_%s", rString)
	testDescription := "test description"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionDescription(layerName, testDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "description", testDescription),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func TestAccLambdaLayerVersion_licenseInfo(t *testing.T) {
	resourceName := "aws_lambda_layer_version.lambda_layer_test"
	rString := sdkacctest.RandString(8)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_license_info_%s", rString)
	testLicenseInfo := "MIT"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLayerVersionLicenseInfo(layerName, testLicenseInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionExists(resourceName, layerName),
					resource.TestCheckResourceAttr(resourceName, "license_info", testLicenseInfo),
				),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename"},
			},
		},
	})
}

func testAccCheckLambdaLayerVersionDestroy(s *terraform.State) error {
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
		if tfawserr.ErrMessageContains(err, lambda.ErrCodeResourceNotFoundException, "") {
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

func testAccLayerVersionBasic(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}
`, layerName)
}

func testAccLayerVersionS3(bucketName, layerName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = "%s"
}

resource "aws_s3_bucket_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.id
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_lambda_layer_version" "lambda_layer_test" {
  s3_bucket  = aws_s3_bucket.lambda_bucket.id
  s3_key     = aws_s3_bucket_object.lambda_code.id
  layer_name = "%s"
}
`, bucketName, layerName)
}

func testAccLayerVersionCreateBeforeDestroy(layerName string, filename string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename         = "%s"
  layer_name       = "%s"
  source_code_hash = filebase64sha256("%s")

  lifecycle {
    create_before_destroy = true
  }
}
`, filename, layerName, filename)
}

func testAccLayerVersionCompatibleRuntimes(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"

  compatible_runtimes = ["nodejs12.x", "nodejs10.x"]
}
`, layerName)
}

func testAccLayerVersionCompatibleArchitecturesNone(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}
`, layerName)
}

func testAccLayerVersionCompatibleArchitecturesX86Arm(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = "%s"
  compatible_architectures = ["x86_64", "arm64"]
}
`, layerName)
}

func testAccLayerVersionCompatibleArchitecturesX86(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = "%s"
  compatible_architectures = ["x86_64"]
}
`, layerName)
}

func testAccLayerVersionCompatibleArchitecturesArm(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = "%s"
  compatible_architectures = ["arm64"]
}
`, layerName)
}

func testAccLayerVersionDescription(layerName string, description string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"

  description = "%s"
}
`, layerName, description)
}

func testAccLayerVersionLicenseInfo(layerName string, licenseInfo string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer_test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"

  license_info = "%s"
}
`, layerName, licenseInfo)
}
