package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
)

func TestAccAwsConnectLambdaFunctionAssociationDataSource_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_lambda_function_association.test"
	datasourceName := "data.aws_connect_lambda_function_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsConnectLambdaFunctionAssociationDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLambdaFunctionAssociationDataSourceConfigBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "function_arn", resourceName, "function_arn"),
				),
			},
		},
	})
}

func testAccAwsConnectLambdaFunctionAssociationDataSourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_lambda_function_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Lambda Function Association ID not set")
		}
		instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		associatedFunctionArn, err := finder.LambdaFunctionAssociationByArnWithContext(context.Background(), conn, instanceID, functionArn)

		if err != nil && !isAWSErr(err, "ResourceNotFoundException", "") {
			return fmt.Errorf("error Connect Lambda Function Association (%s): still exists: %w", associatedFunctionArn, err)
		}

		if associatedFunctionArn != "" {
			return fmt.Errorf("error Connect Lambda Function Association (%s): still exists", associatedFunctionArn)
		}
	}
	return nil
}

func testAccAwsConnectLambdaFunctionAssociationDataSourceBaseConfig(rName string, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[2]q
  outbound_calls_enabled   = true
}

resource "aws_connect_lambda_function_association" "test" {
  instance_id  = aws_connect_instance.test.id
  function_arn = aws_lambda_function.test.arn
}
`, rName, rName2)
}

func testAccAwsConnectLambdaFunctionAssociationDataSourceConfigBasic(rName string, rName2 string) string {
	return fmt.Sprintf(testAccAwsConnectLambdaFunctionAssociationDataSourceBaseConfig(rName, rName2) + `
data "aws_connect_lambda_function_association" "test" {
  function_arn = aws_lambda_function.test.arn
  instance_id  = aws_connect_instance.test.id
}
`)
}
