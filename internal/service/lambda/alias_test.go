package lambda_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccLambdaAlias_basic(t *testing.T) {
	var conf lambda.AliasConfiguration
	resourceName := "aws_lambda_alias.test"

	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)

	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePart),
					testAccCheckAliasInvokeARN(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config:   testAccAliasUsingFunctionNameConfig(roleName, policyName, attachmentName, funcName, aliasName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccLambdaAlias_FunctionName_name(t *testing.T) {
	var conf lambda.AliasConfiguration

	resourceName := "aws_lambda_alias.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasUsingFunctionNameConfig(rName, rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s:%s", rName, rName)),
					testAccCheckAliasInvokeARN(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaAlias_nameUpdate(t *testing.T) {
	var conf lambda.AliasConfiguration
	resourceName := "aws_lambda_alias.test"

	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)
	aliasNameUpdate := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", sdkacctest.RandString(8))

	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)
	functionArnResourcePartUpdate := fmt.Sprintf("function:%s:%s", funcName, aliasNameUpdate)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePart),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePartUpdate),
				),
			},
		},
	})
}

func TestAccLambdaAlias_routing(t *testing.T) {
	var conf lambda.AliasConfiguration
	resourceName := "aws_lambda_alias.test"

	rString := sdkacctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_alias_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_alias_basic_%s", rString)
	attachmentName := fmt.Sprintf("tf_acc_attachment_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_lambda_alias_basic_%s", rString)

	functionArnResourcePart := fmt.Sprintf("function:%s:%s", funcName, aliasName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePart),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAliasImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAliasWithRoutingConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingExistsConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePart),
				),
			},
			{
				Config: testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAliasExists(resourceName, &conf),
					testAccCheckAliasAttributes(&conf),
					testAccCheckAliasRoutingDoesNotExistConfig(&conf),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", functionArnResourcePart),
				),
			},
		},
	})
}

func testAccCheckAliasDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_alias" {
			continue
		}

		_, err := conn.GetAlias(&lambda.GetAliasInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda alias was not deleted")
		}

	}

	return nil
}

func testAccCheckAliasExists(n string, mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Lambda alias not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda alias not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		params := &lambda.GetAliasInput{
			FunctionName: aws.String(rs.Primary.ID),
			Name:         aws.String(rs.Primary.Attributes["name"]),
		}

		getAliasConfiguration, err := conn.GetAlias(params)
		if err != nil {
			return err
		}

		*mapping = *getAliasConfiguration

		return nil
	}
}

func testAccCheckAliasAttributes(mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := *mapping.Name
		arn := *mapping.AliasArn
		if arn == "" {
			return fmt.Errorf("Could not read Lambda alias ARN")
		}
		if name == "" {
			return fmt.Errorf("Could not read Lambda alias name")
		}
		return nil
	}
}

func testAccCheckAliasInvokeARN(name string, mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.StringValue(mapping.AliasArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccCheckAliasRoutingExistsConfig(mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routingConfig := mapping.RoutingConfig

		if routingConfig == nil {
			return fmt.Errorf("Could not read Lambda alias routing config")
		}
		if len(routingConfig.AdditionalVersionWeights) != 1 {
			return fmt.Errorf("Could not read Lambda alias additional version weights")
		}
		return nil
	}
}

func testAccCheckAliasRoutingDoesNotExistConfig(mapping *lambda.AliasConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		routingConfig := mapping.RoutingConfig

		if routingConfig != nil {
			return fmt.Errorf("Lambda alias routing config still exists after removal")
		}
		return nil
	}
}

func testAccAliasImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccAliasBaseConfig(roleName, policyName, attachmentName string) string {
	return fmt.Sprintf(`
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

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for for Lamda alias testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}
`, roleName, policyName, attachmentName)
}

func testAccAliasConfig(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasBaseConfig(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = "%s"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = "%s"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"
}
`, funcName, aliasName))
}

func testAccAliasUsingFunctionNameConfig(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasBaseConfig(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  function_name    = "%s"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = "%s"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "1"
}
`, funcName, aliasName))
}

func testAccAliasWithRoutingConfig(roleName, policyName, attachmentName, funcName, aliasName string) string {
	return acctest.ConfigCompose(
		testAccAliasBaseConfig(roleName, policyName, attachmentName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest_modified.zip"
  function_name    = "%s"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest_modified.zip")
  publish          = "true"
}

resource "aws_lambda_alias" "test" {
  name             = "%s"
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "1"

  routing_config {
    additional_version_weights = {
      "2" = 0.5
    }
  }
}
`, funcName, aliasName))
}
