package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaFunctionURLDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambda_function_url.test"
	resourceName := "aws_lambda_function_url.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccFunctionURLPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionURLDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "authorization_type", resourceName, "authorization_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.#", resourceName, "cors.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.allow_credentials", resourceName, "cors.0.allow_credentials"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.allow_headers.#", resourceName, "cors.0.allow_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.allow_methods.#", resourceName, "cors.0.allow_methods.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.allow_origins.#", resourceName, "cors.0.allow_origins.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.expose_headers.#", resourceName, "cors.0.expose_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors.0.max_age", resourceName, "cors.0.max_age"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_arn", resourceName, "function_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_name", resourceName, "function_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "function_url", resourceName, "function_url"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "qualifier", resourceName, "qualifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url_id", resourceName, "url_id"),
				),
			},
		},
	})
}

func testAccFunctionURLDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "lambda" {
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

resource "aws_iam_role_policy" "lambda" {
  name = %[1]q
  role = aws_iam_role.lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.lambda.arn
  runtime       = "nodejs14.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.arn
  authorization_type = "AWS_IAM"

  cors {
    allow_credentials = true
    allow_origins     = ["*"]
    allow_methods     = ["*"]
    allow_headers     = ["date", "keep-alive"]
    expose_headers    = ["keep-alive", "date"]
    max_age           = 86400
  }
}

`, rName)
}

func testAccFunctionURLDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionURLDataSourceConfig_base(rName), `
data "aws_lambda_function_url" "test" {
  function_name = aws_lambda_function_url.test.function_name
}
`)
}
