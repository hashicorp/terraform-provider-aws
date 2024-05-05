// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaInvocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, testData),
					testAccInvocationConfig_invocation(inputJSON, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_qualifier(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_qualifier(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":"value1","key2":"value2","key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_complex(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_complex(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	testData := "value3"
	testData2 := "value4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvocationConfig_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccInvocationConfig_triggers(rName, testData),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData+`"}`),
				),
			},
			{
				Config: testAccInvocationConfig_triggers(rName, testData2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a": "b"}},"key3":"`+testData2+`"}`),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCRUDCreate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`

	extraArgs := `lifecycle_scope = "CRUD"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCRUDUpdateInput(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`
	inputJSON2 := `{"key1":"valueB","key2":"value2"}`
	resultJSON2 := fmt.Sprintf(`{"key1":"valueB","key2":"value2","tf":{"action":"update", "prev_input": %s}}`, inputJSON)

	extraArgs := `lifecycle_scope = "CRUD"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON2, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON2),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCreateOnlyUpdateInput(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2"}`
	inputJSON2 := `{"key1":"valueB","key2":"value2"}`
	resultJSON2 := `{"key1":"valueB","key2":"value2"}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON2, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON2),
				),
			},
		},
	})
}

// TestAccLambdaInvocation_lifecycle_scopeCRUDDestroy will check destroy is handled appropriately.
//
// In order to allow checking the deletion we use a custom lifecycle which will store it's JSON even when a delete action
// is passed. The Lambda function will create the SSM parameter and the check will verify the content.
func TestAccLambdaInvocation_lifecycle_scopeCRUDDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ssmParameterName := fmt.Sprintf("/tf-test/CRUD/%s", rName)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`
	destroyJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","tf":{"action":"delete","prev_input":%s}}`, inputJSON)

	dependsOnSSMPermissions := `depends_on = [aws_iam_role_policy_attachment.test_ssm]`
	crudLifecycle := `lifecycle_scope = "CRUD"`
	extraArgs := dependsOnSSMPermissions + "\n" + crudLifecycle

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ssmParameterName),
					testAccInvocationConfig_crudAllowSSM(rName, ssmParameterName),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ssmParameterName),
					testAccInvocationConfig_crudAllowSSM(rName, ssmParameterName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCRUDDestroyResult(ctx, resourceName, ssmParameterName, destroyJSON),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCreateOnlyToCRUD(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ssmParameterName := fmt.Sprintf("/tf-test/CRUD/%s", rName)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2"}`
	resultJSONCRUD := fmt.Sprintf(`{"key1":"value1","key2":"value2","tf":{"action":"update", "prev_input": %s}}`, inputJSON)

	extraArgs := `lifecycle_scope = "CRUD"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_crudAllowSSM(rName, ssmParameterName),
					testAccInvocationConfig_invocation(inputJSON, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_crudAllowSSM(rName, ssmParameterName),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSONCRUD),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_terraformKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","custom_key":{"action":"create", "prev_input": null}}`

	terraformKey := `terraform_key = "custom_key"`
	crudLifecycle := `lifecycle_scope = "CRUD"`
	extraArgs := terraformKey + "\n" + crudLifecycle

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
		},
	})
}

// testAccCheckCRUDDestroyResult verifies that when CRUD lifecycle is active that a destroyed resource
// triggers the lambda.
//
// Because a destroy implies the resource will be removed from the state we need another way to check
// how the lambda was invoked. The JSON used to invoke the lambda is stored in an SSM Parameter.
// We will read it out, compare with the expected result and clean up the SSM parameter.
func testAccCheckCRUDDestroyResult(ctx context.Context, name, ssmParameterName, expectedResult string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("Still found resource in state: %s", name)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMClient(ctx)
		res, err := conn.GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(ssmParameterName),
			WithDecryption: aws.Bool(true),
		})

		if cleanupErr := removeSSMParameter(ctx, conn, ssmParameterName); cleanupErr != nil {
			return fmt.Errorf("Could not cleanup SSM Parameter %s", ssmParameterName)
		}

		if err != nil {
			return fmt.Errorf("Could not get SSM Parameter %s", ssmParameterName)
		}

		if !verify.JSONStringsEqual(*res.Parameter.Value, expectedResult) {
			return fmt.Errorf("%s: input for destroy expected %s, got %s", name, expectedResult, *res.Parameter.Value)
		}

		return nil
	}
}

func removeSSMParameter(ctx context.Context, conn *ssm.Client, name string) error {
	_, err := conn.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	return err
}

func testAccInvocationConfig_base(roleName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role       = aws_iam_role.test.name
}
`, roleName)
}

func testAccInvocationConfig_crudAllowSSM(rName, ssmParameterName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "ssm:PutParameter",
        ]
        Effect   = "Allow"
        Resource = "arn:${data.aws_partition.current.partition}:ssm:*:*:parameter%[2]s"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test_ssm" {
  policy_arn = aws_iam_policy.test.arn
  role       = aws_iam_role.test.name
}
`, rName, ssmParameterName)
}

func testAccInvocationConfig_function(fName, rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvocationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/%[1]s.zip"
  function_name = %[2]q
  role          = aws_iam_role.test.arn
  handler       = "%[1]s.handler"
  runtime       = "nodejs18.x"

  environment {
    variables = {
      TEST_DATA = %[3]q
    }
  }
}
`, fName, rName, testData))
}

func testAccInvocationConfig_invocation(inputJSON, extraArgs string) string {
	return fmt.Sprintf(`
resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  input = %[1]s
  %[2]s
}
`, strconv.Quote(inputJSON), extraArgs)
}

func testAccInvocationConfig_qualifier(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvocationConfig_function("lambda_invocation", rName, testData),
		`
resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name
  qualifier     = aws_lambda_function.test.version

  input = jsonencode({
    key1 = "value1"
    key2 = "value2"
  })
}
`)
}

func testAccInvocationConfig_complex(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvocationConfig_function("lambda_invocation", rName, testData),
		`
resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  input = jsonencode({
    key1 = {
      subkey1 = "subvalue1"
    }
    key2 = {
      subkey2 = "subvalue2"
      subkey3 = {
        a = "b"
      }
    }
  })
}
`)
}

func testAccInvocationConfig_triggers(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvocationConfig_function("lambda_invocation", rName, testData),
		`
resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name

  triggers = {
    redeployment = sha1(jsonencode([
      aws_lambda_function.test.environment
    ]))
  }

  input = jsonencode({
    key1 = {
      subkey1 = "subvalue1"
    }
    key2 = {
      subkey2 = "subvalue2"
      subkey3 = {
        a = "b"
      }
    }
  })
}
`)
}
