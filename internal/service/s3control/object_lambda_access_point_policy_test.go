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

func TestAccS3ControlObjectLambdaAccessPointPolicy_basic(t *testing.T) {
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccS3ControlObjectLambdaAccessPointPolicy_disappears(t *testing.T) {
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceObjectLambdaAccessPointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPointPolicy_Disappears_accessPoint(t *testing.T) {
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	accessPointResourceName := "aws_s3control_object_lambda_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceObjectLambdaAccessPoint(), accessPointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPointPolicy_update(t *testing.T) {
	resourceName := "aws_s3control_object_lambda_access_point_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckObjectLambdaAccessPointPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccObjectLambdaAccessPointPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointPolicyExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccCheckObjectLambdaAccessPointPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_object_lambda_access_point_policy" {
			continue
		}

		accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, _, err = tfs3control.FindObjectLambdaAccessPointPolicyAndStatusByAccountIDAndName(conn, accountID, name)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Object Lambda Access Point Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckObjectLambdaAccessPointPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Object Lambda Access Point Policy ID is set")
		}

		accountID, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		_, _, err = tfs3control.FindObjectLambdaAccessPointPolicyAndStatusByAccountIDAndName(conn, accountID, name)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccObjectLambdaAccessPointPolicyConfig(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

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

resource "aws_s3control_object_lambda_access_point_policy" "test" {
  name = aws_s3control_object_lambda_access_point.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3-object-lambda:GetObject"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = aws_s3control_object_lambda_access_point.test.arn
    }]
  })
}
`, rName))
}

func testAccObjectLambdaAccessPointPolicyUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(testAccObjectLambdaAccessPointBaseConfig(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

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

resource "aws_s3control_object_lambda_access_point_policy" "test" {
  name = aws_s3control_object_lambda_access_point.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Effect = "Allow"
      Action = "s3-object-lambda:*"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Resource = aws_s3control_object_lambda_access_point.test.arn
    }]
  })
}
`, rName))
}
