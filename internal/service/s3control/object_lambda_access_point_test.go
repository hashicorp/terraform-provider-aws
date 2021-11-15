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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
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

func TestAccS3ControlObjectLambdaAccessPoint_disappears_Bucket(t *testing.T) {
	var v s3control.ObjectLambdaConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckDestroyBucket(bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
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
			return fmt.Errorf("No S3 Object Lambda Access Point is set")
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
