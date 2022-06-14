package lambda_test

import (
	"fmt"
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

func TestAccLambdaLayerVersionPermission_basic_byARN(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_basic_arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_basic_byName(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_basic_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_org(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_org(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttr(resourceName, "organization_id", "o-0123456789"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_account(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_name", "aws_lambda_layer_version.test", "layer_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaLayerVersionPermission_disappears(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerVersionPermissionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflambda.ResourceLayerVersionPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Creating Lambda layer and Lambda layer permissions

func testLayerVersionPermission_basic_arn(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = %[1]q
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_name     = aws_lambda_layer_version.test.layer_arn
  version_number = aws_lambda_layer_version.test.version
  action         = "lambda:GetLayerVersion"
  statement_id   = "xaccount"
  principal      = "*"
}
`, layerName)
}

func testLayerVersionPermission_basic_name(layerName string) string {
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

func testLayerVersionPermission_org(layerName string) string {
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

func testLayerVersionPermission_account(layerName string) string {
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

func testAccCheckLayerVersionPermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda Layer version policy ID not set")
		}

		layerName, versionNumber, err := tflambda.ResourceLayerVersionPermissionParseId(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing lambda layer ID: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(versionNumber),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckLayerVersionPermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version_permission" {
			continue
		}

		layerName, versionNumber, err := tflambda.ResourceLayerVersionPermissionParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(versionNumber),
		})

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}
	return nil
}
