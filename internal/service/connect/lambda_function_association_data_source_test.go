package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectLambdaFunctionAssociationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_lambda_function_association.test"
	datasourceName := "data.aws_connect_lambda_function_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLambdaFunctionAssociationDataSource_ConfigBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "function_arn", resourceName, "function_arn"),
				),
			},
		},
	})
}

func testAccLambdaFunctionAssociationDataSource_BaseConfig(rName string, rName2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs14.x"
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
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
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

func testAccLambdaFunctionAssociationDataSource_ConfigBasic(rName string, rName2 string) string {
	return fmt.Sprintf(testAccLambdaFunctionAssociationDataSource_BaseConfig(rName, rName2) + `
data "aws_connect_lambda_function_association" "test" {
  function_arn = aws_connect_lambda_function_association.test.function_arn
  instance_id  = aws_connect_instance.test.id
}
`)
}
