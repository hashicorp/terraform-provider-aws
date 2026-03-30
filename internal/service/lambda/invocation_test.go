// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseRecordID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input                               string
		FunctionName, Qualifier, ResultHash string
		ExpectError                         bool
	}{
		// Invalid cases
		{"ABCDEF", "", "", "", true},
		{"ABCDEF,42", "", "", "", true},
		{"ABCDEF,,", "", "", "", true},
		{"ABCDEF,invalid_qualifier,b326b5062b2f0e69046810717534cb09", "", "", "", true},
		{"ABCDEF,42,invalid_hash", "", "", "", true},
		// Valid cases
		{"ABCDEF,42,b326b5062b2f0e69046810717534cb09", "ABCDEF", "42", "b326b5062b2f0e69046810717534cb09", false},
		{"ABC_DEF,42,b326b5062b2f0e69046810717534cb09", "ABC_DEF", "42", "b326b5062b2f0e69046810717534cb09", false},
		{"ABCDEF,$LATEST,b326b5062b2f0e69046810717534cb09", "ABCDEF", "$LATEST", "b326b5062b2f0e69046810717534cb09", false},
		{"ABC_DEF,$LATEST,b326b5062b2f0e69046810717534cb09", "ABC_DEF", "$LATEST", "b326b5062b2f0e69046810717534cb09", false},
		{"ABC_DEF_1234,567,b326b5062b2f0e69046810717534cb09", "ABC_DEF_1234", "567", "b326b5062b2f0e69046810717534cb09", false},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()

			functionName, qualifier, resultHash, err := tflambda.InvocationParseResourceID(tc.Input)

			if tc.ExpectError {
				if err == nil {
					t.Fatalf("expected error for input: %s", tc.Input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %s: %v", tc.Input, err)
			}

			if functionName != tc.FunctionName {
				t.Fatalf("input: %s\nfunction_name: %s\nexpected:%s", tc.Input, functionName, tc.FunctionName)
			}
			if qualifier != tc.Qualifier {
				t.Fatalf("input: %s\nqualifier: %s\nexpected:%s", tc.Input, qualifier, tc.Qualifier)
			}
			if resultHash != tc.ResultHash {
				t.Fatalf("input: %s\nresult: %s\nexpected:%s", tc.Input, resultHash, tc.ResultHash)
			}
		})
	}
}

func TestInvocationResourceIDCreation(t *testing.T) {
	t.Parallel()

	functionName := "my_test_function"
	qualifier := "$LATEST"
	resultHash := "b326b5062b2f0e69046810717534cb09"

	expectedID := "my_test_function,$LATEST,b326b5062b2f0e69046810717534cb09"

	// Test parsing the expected ID format
	parsedFunctionName, parsedQualifier, parsedResultHash, err := tflambda.InvocationParseResourceID(expectedID)
	if err != nil {
		t.Fatalf("unexpected error parsing resource ID: %v", err)
	}

	if parsedFunctionName != functionName {
		t.Fatalf("expected function name: %s, got: %s", functionName, parsedFunctionName)
	}
	if parsedQualifier != qualifier {
		t.Fatalf("expected qualifier: %s, got: %s", qualifier, parsedQualifier)
	}
	if parsedResultHash != resultHash {
		t.Fatalf("expected result hash: %s, got: %s", resultHash, parsedResultHash)
	}
}

// TestBuildInputPanic verifies the fix for panic: assignment to entry in nil map
// This happens when input is empty and lifecycle_scope is CRUD
func TestBuildInputPanic(t *testing.T) {
	t.Parallel()

	// Create ResourceData with empty input and CRUD lifecycle scope
	resourceSchema := tflambda.ResourceInvocation().Schema
	d := schema.TestResourceDataRaw(t, resourceSchema, map[string]any{
		"function_name":   "test-function",
		"input":           "", // Empty input causes getObjectFromJSONString to return nil
		"lifecycle_scope": "CRUD",
		"terraform_key":   "tf",
	})
	d.SetId("test-id")

	// This should NOT panic after the fix
	result, err := tflambda.BuildInput(d, tflambda.InvocationActionDelete)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected result, but got nil")
	}
}

func TestAccLambdaInvocation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"input", "lifecycle_scope", "result", "terraform_key"},
			},
		},
	})
}

func TestAccLambdaInvocation_qualifier(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"input", "lifecycle_scope", "result", "terraform_key"},
			},
		},
	})
}

func TestAccLambdaInvocation_complex(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	testData2 := "value4"

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`

	extraArgs := `lifecycle_scope = "CRUD"`

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"input", "lifecycle_scope", "result", "terraform_key"},
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCRUDUpdateInput(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ssmParamResourceName := "aws_ssm_parameter.result_key1"

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`
	inputJSON2 := `{"key1":"valueB","key2":"value2"}`
	resultJSON2 := fmt.Sprintf(`{"key1":"valueB","key2":"value2","tf":{"action":"update", "prev_input": %s}}`, inputJSON)

	extraArgs := `lifecycle_scope = "CRUD"`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_dependency(rName, resourceName),
					testAccInvocationConfig_invocation(inputJSON, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
					testAccCheckInvocationResultUpdatedSSMParam(ssmParamResourceName, acctest.CtValue1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, ""),
					testAccInvocationConfig_dependency(rName, resourceName),
					testAccInvocationConfig_invocation(inputJSON2, extraArgs),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON2),
					testAccCheckInvocationResultUpdatedSSMParam(ssmParamResourceName, "valueB"),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCreateOnlyUpdateInput(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2"}`
	inputJSON2 := `{"key1":"valueB","key2":"value2"}`
	resultJSON2 := `{"key1":"valueB","key2":"value2"}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ssmParameterName := fmt.Sprintf("/tf-test/CRUD/%s", rName)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","tf":{"action":"create", "prev_input": null}}`
	destroyJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","tf":{"action":"delete","prev_input":%s}}`, inputJSON)

	dependsOnSSMPermissions := `depends_on = [aws_iam_role_policy_attachment.test_ssm]`
	crudLifecycle := `lifecycle_scope = "CRUD"`
	extraArgs := dependsOnSSMPermissions + "\n" + crudLifecycle

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
					testAccCheckCRUDDestroyResult(ctx, t, resourceName, ssmParameterName, destroyJSON),
				),
			},
		},
	})
}

func TestAccLambdaInvocation_lifecycle_scopeCreateOnlyToCRUD(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation_crud"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	ssmParameterName := fmt.Sprintf("/tf-test/CRUD/%s", rName)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2"}`
	resultJSONCRUD := fmt.Sprintf(`{"key1":"value1","key2":"value2","tf":{"action":"update", "prev_input": %s}}`, inputJSON)

	extraArgs := `lifecycle_scope = "CRUD"`

	acctest.ParallelTest(ctx, t, resource.TestCase{
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := `{"key1":"value1","key2":"value2","custom_key":{"action":"create", "prev_input": null}}`

	terraformKey := `terraform_key = "custom_key"`
	crudLifecycle := `lifecycle_scope = "CRUD"`
	extraArgs := terraformKey + "\n" + crudLifecycle

	acctest.ParallelTest(ctx, t, resource.TestCase{
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

// Tests the state upgrader coming from a version < v5.1.0 where the default values
// from the arguments added in 5.1.0 are not yet present.
//
// This causes unintentional invocations and/or issues processing input which is
// valid type for CREATE_ONLY lifecycle_scope.
//
// https://github.com/hashicorp/terraform-provider-aws/issues/40954
// https://github.com/hashicorp/terraform-provider-aws/issues/31786
func TestAccLambdaInvocation_UpgradeState_Pre_v5_1_0(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.65.0",
					},
				},
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, testData),
					testAccInvocationConfig_invocation(inputJSON, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

// Tests the state upgrader in cases where the default values from the arguments added
// in v5.1.0 are already present
func TestAccLambdaInvocation_UpgradeState_v5_83_0(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.LambdaServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.83.0",
					},
				},
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function(fName, rName, testData),
					testAccInvocationConfig_invocation(inputJSON, ""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

func TestAccLambdaInvocation_tenantID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lambda_invocation.test"
	fName := "lambda_invocation"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	testData := "value3"
	tenantID := "test-tenant-123"
	inputJSON := `{"key1":"value1","key2":"value2"}`
	resultJSON := fmt.Sprintf(`{"key1":"value1","key2":"value2","key3":%q}`, testData)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccInvocationConfig_function_with_tenant(fName, rName, testData),
					testAccInvocationConfig_tenantID(inputJSON, tenantID),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInvocationResult(resourceName, resultJSON),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"input", "lifecycle_scope", "result", "terraform_key", "tenant_id"},
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
func testAccCheckCRUDDestroyResult(ctx context.Context, t *testing.T, name, ssmParameterName, expectedResult string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("Still found resource in state: %s", name)
		}
		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)
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

func testAccCheckInvocationResultUpdatedSSMParam(name, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("SSM parameter %s not created", name)
		}

		value, ok := rs.Primary.Attributes[names.AttrValue]
		if !ok {
			return fmt.Errorf("SSM parameter attribute 'value' is empty, expected: %s", expectedValue)
		}

		if value != expectedValue {
			return fmt.Errorf("%s: Attribute 'value' expected %s, got %s", name, expectedValue, value)
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

func testAccInvocationConfig_function_with_tenant(fName, rName, testData string) string {
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
  tenancy_config {
    tenant_isolation_mode = "PER_TENANT"
  }
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

func testAccInvocationConfig_dependency(rName, resourceName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "result_key1" {
  name  = "/tf-test/CRUD/%[1]s/key1"
  type  = "String"
  value = try(jsondecode(%[2]s.result).key1, "")
}
`, rName, resourceName)
}

func testAccInvocationConfig_tenantID(inputJSON, tenantID string) string {
	return fmt.Sprintf(`
resource "aws_lambda_invocation" "test" {
  function_name = aws_lambda_function.test.function_name
  tenant_id     = %[2]q

  input = %[1]s
}
`, strconv.Quote(inputJSON), tenantID)
}
