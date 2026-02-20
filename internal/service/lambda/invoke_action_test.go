// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaInvokeAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	expectedResult := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_basic(rName, testData, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeAction(ctx, t, rName, inputJSON, expectedResult),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_withQualifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "qualifier_test"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	expectedResult := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_withQualifier(rName, testData, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionWithQualifier(ctx, t, rName, inputJSON, expectedResult),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_invocationTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "invocation_types_test"
	inputJSON := `{"key1":"value1","key2":"value2"}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_invocationType(rName, testData, inputJSON, "RequestResponse"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionInvocationType(ctx, t, rName, inputJSON, awstypes.InvocationTypeRequestResponse),
				),
			},
			{
				Config: testAccInvokeActionConfig_invocationType(rName, testData, inputJSON, "Event"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionInvocationType(ctx, t, rName, inputJSON, awstypes.InvocationTypeEvent),
				),
			},
			{
				Config: testAccInvokeActionConfig_invocationType(rName, testData, inputJSON, "DryRun"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionInvocationType(ctx, t, rName, inputJSON, awstypes.InvocationTypeDryRun),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_logTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "log_types_test"
	inputJSON := `{"key1":"value1","key2":"value2"}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_logType(rName, testData, inputJSON, "None"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionLogType(ctx, t, rName, inputJSON, awstypes.LogTypeNone),
				),
			},
			{
				Config: testAccInvokeActionConfig_logType(rName, testData, inputJSON, "Tail"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionLogType(ctx, t, rName, inputJSON, awstypes.LogTypeTail),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_clientContext(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "client_context_test"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	clientContext := base64.StdEncoding.EncodeToString([]byte(`{"client":{"client_id":"test_client","app_version":"1.0.0"},"env":{"locale":"en_US"}}`))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_clientContext(rName, testData, inputJSON, clientContext),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeActionClientContext(ctx, t, rName, inputJSON, clientContext),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_complexPayload(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "complex_test"
	inputJSON := `{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a":"b"}}}`
	expectedResult := fmt.Sprintf(`{"key1":{"subkey1":"subvalue1"},"key2":{"subkey2":"subvalue2","subkey3":{"a":"b"}},"key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_basic(rName, testData, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeAction(ctx, t, rName, inputJSON, expectedResult),
				),
			},
		},
	})
}

func TestAccLambdaInvokeAction_tenantId(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	expectedResult := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInvokeActionConfig_tenantId(rName, testData, inputJSON),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvokeAction(ctx, t, rName, inputJSON, expectedResult),
				),
			},
		},
	})
}

// Test helper functions

// testAccCheckInvokeAction verifies that the action can successfully invoke a Lambda function
func testAccCheckInvokeAction(ctx context.Context, t *testing.T, functionName, inputJSON, expectedResult string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		// Invoke the function directly to verify it's working and compare results
		input := &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: awstypes.InvocationTypeRequestResponse,
			Payload:        []byte(inputJSON),
		}

		output, err := conn.Invoke(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function %s: %w", functionName, err)
		}

		if output.FunctionError != nil {
			return fmt.Errorf("Lambda function %s returned an error: %s", functionName, string(output.Payload))
		}

		actualResult := string(output.Payload)
		if actualResult != expectedResult {
			return fmt.Errorf("Lambda function %s result mismatch. Expected: %s, Got: %s", functionName, expectedResult, actualResult)
		}

		return nil
	}
}

// testAccCheckInvokeActionWithQualifier verifies action works with function qualifiers
func testAccCheckInvokeActionWithQualifier(ctx context.Context, t *testing.T, functionName, inputJSON, expectedResult string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		// Get the function to retrieve the version
		getFunc, err := conn.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: &functionName,
		})
		if err != nil {
			return fmt.Errorf("Failed to get Lambda function %s: %w", functionName, err)
		}

		// Invoke with the specific version
		input := &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: awstypes.InvocationTypeRequestResponse,
			Payload:        []byte(inputJSON),
			Qualifier:      getFunc.Configuration.Version,
		}

		output, err := conn.Invoke(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function %s with qualifier: %w", functionName, err)
		}

		if output.FunctionError != nil {
			return fmt.Errorf("Lambda function %s returned an error: %s", functionName, string(output.Payload))
		}

		actualResult := string(output.Payload)
		if actualResult != expectedResult {
			return fmt.Errorf("Lambda function %s result mismatch with qualifier. Expected: %s, Got: %s", functionName, expectedResult, actualResult)
		}

		return nil
	}
}

// testAccCheckInvokeActionInvocationType verifies different invocation types work
func testAccCheckInvokeActionInvocationType(ctx context.Context, t *testing.T, functionName, inputJSON string, invocationType awstypes.InvocationType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		input := &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: invocationType,
			Payload:        []byte(inputJSON),
		}

		output, err := conn.Invoke(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function %s with invocation type %s: %w", functionName, string(invocationType), err)
		}

		// For async invocations, we just verify the request was accepted
		if invocationType == awstypes.InvocationTypeEvent {
			if output.StatusCode != http.StatusAccepted {
				return fmt.Errorf("Expected status code 202 for async invocation, got %d", output.StatusCode)
			}
		}

		// For dry run, we verify the function would execute successfully
		if invocationType == awstypes.InvocationTypeDryRun {
			if output.StatusCode != http.StatusNoContent {
				return fmt.Errorf("Expected status code 204 for dry run, got %d", output.StatusCode)
			}
		}

		return nil
	}
}

// testAccCheckInvokeActionLogType verifies log type configuration works
func testAccCheckInvokeActionLogType(ctx context.Context, t *testing.T, functionName, inputJSON string, logType awstypes.LogType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		input := &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: awstypes.InvocationTypeRequestResponse,
			Payload:        []byte(inputJSON),
			LogType:        logType,
		}

		output, err := conn.Invoke(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function %s with log type %s: %w", functionName, string(logType), err)
		}

		if output.FunctionError != nil {
			return fmt.Errorf("Lambda function %s returned an error: %s", functionName, string(output.Payload))
		}

		// If log type is Tail, we should have log results
		if logType == awstypes.LogTypeTail {
			if output.LogResult == nil {
				return fmt.Errorf("Expected log result when log type is Tail, but got none")
			}
		}

		return nil
	}
}

// testAccCheckInvokeActionClientContext verifies client context is passed correctly
func testAccCheckInvokeActionClientContext(ctx context.Context, t *testing.T, functionName, inputJSON, clientContext string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		input := &lambda.InvokeInput{
			FunctionName:   &functionName,
			InvocationType: awstypes.InvocationTypeRequestResponse,
			Payload:        []byte(inputJSON),
			ClientContext:  &clientContext,
		}

		output, err := conn.Invoke(ctx, input)
		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function %s with client context: %w", functionName, err)
		}

		if output.FunctionError != nil {
			return fmt.Errorf("Lambda function %s returned an error: %s", functionName, string(output.Payload))
		}

		return nil
	}
}

// Configuration functions

func testAccInvokeActionConfig_base(rName string) string {
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
`, rName)
}

func testAccInvokeActionConfig_function(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs18.x"

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}
`, rName, testData))
}

func testAccInvokeActionConfig_basic(rName, testData, inputJSON string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_function(rName, testData),
		fmt.Sprintf(`
action "aws_lambda_invoke" "test" {
  config {
    function_name = aws_lambda_function.test.function_name
    payload       = %[1]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, inputJSON))
}

func testAccInvokeActionConfig_withQualifier(rName, testData, inputJSON string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs18.x"
  publish       = true

  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}

action "aws_lambda_invoke" "test" {
  config {
    function_name = aws_lambda_function.test.function_name
    payload       = %[3]q
    qualifier     = aws_lambda_function.test.version
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, rName, testData, inputJSON))
}

func testAccInvokeActionConfig_invocationType(rName, testData, inputJSON, invocationType string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_function(rName, testData),
		fmt.Sprintf(`
action "aws_lambda_invoke" "test" {
  config {
    function_name   = aws_lambda_function.test.function_name
    payload         = %[1]q
    invocation_type = %[2]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, inputJSON, invocationType))
}

func testAccInvokeActionConfig_logType(rName, testData, inputJSON, logType string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_function(rName, testData),
		fmt.Sprintf(`
action "aws_lambda_invoke" "test" {
  config {
    function_name = aws_lambda_function.test.function_name
    payload       = %[1]q
    log_type      = %[2]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, inputJSON, logType))
}

func testAccInvokeActionConfig_clientContext(rName, testData, inputJSON, clientContext string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_function(rName, testData),
		fmt.Sprintf(`
action "aws_lambda_invoke" "test" {
  config {
    function_name  = aws_lambda_function.test.function_name
    payload        = %[1]q
    client_context = %[2]q
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, inputJSON, clientContext))
}

func testAccInvokeActionConfig_function_tenant(rName, testData string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambda_invocation.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "lambda_invocation.handler"
  runtime       = "nodejs18.x"

  tenancy_config {
    tenant_isolation_mode = "PER_TENANT"
  }
  environment {
    variables = {
      TEST_DATA = %[2]q
    }
  }
}
`, rName, testData))
}

func testAccInvokeActionConfig_tenantId(rName, testData, inputJSON string) string {
	return acctest.ConfigCompose(
		testAccInvokeActionConfig_function_tenant(rName, testData),
		fmt.Sprintf(`
action "aws_lambda_invoke" "test" {
  config {
    function_name = aws_lambda_function.test.function_name
    payload       = %[1]q
    tenant_id     = "tenant-1"
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_lambda_invoke.test]
    }
  }
}
`, inputJSON))
}
