package s3control_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3ControlObjectLambdaAccessPoint_basic(t *testing.T) {
	var v s3control.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	accessPointResourceName := "aws_s3_access_point.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             "1",
						"content_transformation.#":              "1",
						"content_transformation.0.aws_lambda.#": "1",
						"content_transformation.0.aws_lambda.0.function_payload": "",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, "arn"),
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

func TestAccS3ControlObjectLambdaAccessPoint_disappears(t *testing.T) {
	var v s3control.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceObjectLambdaAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPoint_update(t *testing.T) {
	var v s3control.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	accessPointResourceName := "aws_s3_access_point.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointOptionalsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.allowed_features.*", "GetObject-PartNumber"),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.allowed_features.*", "GetObject-Range"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             "1",
						"content_transformation.#":              "1",
						"content_transformation.0.aws_lambda.#": "1",
						"content_transformation.0.aws_lambda.0.function_payload": "{\"res-x\": \"100\",\"res-y\": \"100\"}",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccObjectLambdaAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "s3-object-lambda", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.allowed_features.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.cloud_watch_metrics_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.supporting_access_point", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.transformation_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration.0.transformation_configuration.*", map[string]string{
						"actions.#":                             "1",
						"content_transformation.#":              "1",
						"content_transformation.0.aws_lambda.#": "1",
						"content_transformation.0.aws_lambda.0.function_payload": "",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "configuration.0.transformation_configuration.*.actions.*", "GetObject"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "configuration.0.transformation_configuration.*.content_transformation.0.aws_lambda.0.function_arn", lambdaFunctionResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckObjectLambdaAccessPointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_object_lambda_access_point" {
			continue
		}

		accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfs3control.FindObjectLambdaAccessPointByAccountIDAndName(conn, accountID, name)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Object Lambda Access Point %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckObjectLambdaAccessPointExists(n string, v *s3control.ObjectLambdaConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Object Lambda Access Point ID is set")
		}

		accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		output, err := tfs3control.FindObjectLambdaAccessPointByAccountIDAndName(conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccObjectLambdaAccessPointBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLambdaBase(rName, rName, rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "index.handler"
  runtime       = "nodejs14.x"
}
`, rName))
}

func testAccObjectLambdaAccessPointConfig(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q
}

resource "aws_s3control_object_lambda_access_point" "test" {
  name = %[1]q

  configuration {
    supporting_access_point = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn = aws_lambda_function.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccObjectLambdaAccessPointOptionalsConfig(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.id
  name   = %[1]q
}

resource "aws_s3control_object_lambda_access_point" "test" {
  name = %[1]q

  configuration {
    allowed_features            = ["GetObject-Range", "GetObject-PartNumber"]
    cloud_watch_metrics_enabled = true
    supporting_access_point     = aws_s3_access_point.test.arn

    transformation_configuration {
      actions = ["GetObject"]

      content_transformation {
        aws_lambda {
          function_arn     = aws_lambda_function.test.arn
          function_payload = "{\"res-x\": \"100\",\"res-y\": \"100\"}"
        }
      }
    }
  }
}
`, rName))
}
