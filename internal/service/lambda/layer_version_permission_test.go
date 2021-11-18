package lambda_test

import (
	"fmt"
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
)

func TestLambdaLayerVersionPermission_all(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_all(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.test", "layer_arn"),
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

func TestLambdaLayerVersionPermission_org(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_org(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttr(resourceName, "organization_id", "o-0123456789"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.test", "layer_arn"),
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

func TestLambdaLayerVersionPermission_account(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "456789820214"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
					resource.TestCheckResourceAttrPair(resourceName, "layer_arn", "aws_lambda_layer_version.test", "layer_arn"),
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

func TestLambdaLayerVersionPermission_disappears(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLambdaLayerVersionPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLayerVersionPermission_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionPermissionExists(resourceName, rName),
					acctest.CheckResourceDisappears(acctest.Provider, tflambda.ResourceLayerVersionPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Creating Lambda layer and Lambda layer permissions

func testLayerVersionPermission_all(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_arn     = aws_lambda_layer_version.test.layer_arn
  layer_version = aws_lambda_layer_version.test.version
  action        = "lambda:GetLayerVersion"
  statement_id  = "xaccount"
  principal     = "*"
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
  layer_arn       = aws_lambda_layer_version.test.layer_arn
  layer_version   = aws_lambda_layer_version.test.version
  action          = "lambda:GetLayerVersion"
  statement_id    = "xaccount"
  principal       = "*"
  organization_id = "o-0123456789"
}
`, layerName)
}

func testLayerVersionPermission_account(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename   = "test-fixtures/lambdatest.zip"
  layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "test" {
  layer_arn     = aws_lambda_layer_version.test.layer_arn
  layer_version = aws_lambda_layer_version.test.version
  action        = "lambda:GetLayerVersion"
  statement_id  = "xaccount"
  principal     = "456789820214"
}
`, layerName)
}

func testAccCheckAwsLambdaLayerVersionPermissionExists(res, layerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda Layer version permission not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda Layer version policy ID not set")
		}

		if rs.Primary.Attributes["revision_id"] == "" {
			return fmt.Errorf("Lambda Layer Version Permission not set")
		}

		_, _, version, err := tflambda.ResourceLayerVersionPermissionParseId(rs.Primary.Attributes["id"])
		if err != nil {
			return fmt.Errorf("Error parsing lambda layer ID: %s", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			return err
		}

		return err
	}
}

func testAccCheckLambdaLayerVersionPermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version_permission" {
			continue
		}

		layerName, _, version, err := tflambda.ResourceLayerVersionPermissionParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersionPolicy(&lambda.GetLayerVersionPolicyInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})

		if tfawserr.ErrCodeEquals(err, lambda.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		// as I've created Lambda layer, not only layer permission, need to check if layer was destroyed.
		err = testAccCheckLambdaLayerVersionDestroy(s)
		if err != nil {
			return err
		}

		return fmt.Errorf("Lambda Layer Version Permission (%s) still exists", rs.Primary.ID)
	}
	return nil
}
