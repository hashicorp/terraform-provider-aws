package lambda_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaFunctionUrl_basic(t *testing.T) {
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionUrlBasicConfig(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionUrlExists(resourceName, funcName, nil, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", lambda.FunctionUrlAuthTypeNone),
					acctest.CheckResourceAttrRegionalARN(resourceName, "function_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, tflambda.FunctionVersionLatest)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"function_name", "qualifier"},
			},
		},
	})
}

func TestAccLambdaFunctionUrl_Cors(t *testing.T) {
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionUrlCoresConfig(funcName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionUrlExists(resourceName, funcName, nil, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", lambda.FunctionUrlAuthTypeAwsIam),
					acctest.CheckResourceAttrRegionalARN(resourceName, "function_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, tflambda.FunctionVersionLatest)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.0", "date"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.1", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_methods.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_origins.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.0", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.1", "date"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.max_age", "86400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"function_name", "qualifier"},
			},
		},
	})
}

func TestAccLambdaFunctionUrl_Alias(t *testing.T) {
	var conf lambda.GetFunctionUrlConfigOutput
	resourceName := "aws_lambda_function_url.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	aliasName := "live"
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionUrlAliasConfig(funcName, aliasName, policyName, roleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionUrlExists(resourceName, funcName, &aliasName, &conf),
					resource.TestCheckResourceAttr(resourceName, "authorization_type", lambda.FunctionUrlAuthTypeAwsIam),
					acctest.CheckResourceAttrRegionalARN(resourceName, "function_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, aliasName)),
					resource.TestCheckResourceAttr(resourceName, "cors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_credentials", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.0", "date"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_headers.1", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_methods.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.allow_origins.0", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.0", "keep-alive"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.expose_headers.1", "date"),
					resource.TestCheckResourceAttr(resourceName, "cors.0.max_age", "86400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"function_name", "qualifier"},
			},
		},
	})
}

func testAccCheckFunctionUrlExists(res string, funcName string, qualifier *string, function *lambda.GetFunctionUrlConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda function url not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda function url ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		params := &lambda.GetFunctionUrlConfigInput{
			FunctionName: aws.String(funcName),
		}

		if qualifier != nil {
			params.Qualifier = qualifier
		}

		output, err := conn.GetFunctionUrlConfig(params)
		if err != nil {
			return err
		}

		*function = *output

		return nil
	}
}

func testAccFunctionUrlBaseConfig(policyName, roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "%s"
  role = aws_iam_role.iam_for_lambda.id

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
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

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

`, policyName, roleName)
}

func testAccFunctionUrlBasicConfig(funcName, policyName, roleName string) string {
	return fmt.Sprintf(testAccFunctionUrlBaseConfig(policyName, roleName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs14.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  authorization_type = "NONE"
}
`, funcName)
}

func testAccFunctionUrlCoresConfig(funcName, policyName, roleName string) string {
	return fmt.Sprintf(testAccFunctionUrlBaseConfig(policyName, roleName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs14.x"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
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
`, funcName)
}

func testAccFunctionUrlAliasConfig(funcName, aliasName, policyName, roleName string) string {
	return fmt.Sprintf(testAccFunctionUrlBaseConfig(policyName, roleName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs14.x"
  publish       = true
}

resource "aws_lambda_alias" "live" {
  name             = %[2]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "1"
}

resource "aws_lambda_function_url" "test" {
  function_name      = aws_lambda_function.test.function_name
  qualifier          = aws_lambda_alias.live.name
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
`, funcName, aliasName)
}
