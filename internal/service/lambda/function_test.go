package lambda_test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/signer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(lambda.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"InvalidParameterValueException: Unsupported source arn",
		"InvalidParameterValueException: CompatibleArchitectures are not",
	)
}

func TestAccLambdaFunction_basic(t *testing.T) {
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureX8664),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size", "512"),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, tflambda.FunctionVersionLatest)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
					resource.TestCheckResourceAttr(resourceName, "version", tflambda.FunctionVersionLatest),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_unpublishedCodeUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf1, conf2 lambda.GetFunctionOutput

	initialFilename := "test-fixtures/lambdatest.zip"
	updatedFilename, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(updatedFilename)

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_versioned_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_versioned_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_versioned_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_versioned_%s", rString)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFilenameConfig(initialFilename, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "version", tflambda.FunctionVersionLatest),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, tflambda.FunctionVersionLatest)),
				),
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFilenameConfig(updatedFilename, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "version", tflambda.FunctionVersionLatest),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, tflambda.FunctionVersionLatest)),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_disappears(t *testing.T) {
	var function lambda.GetFunctionOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConfig(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &function),
					acctest.CheckResourceDisappears(acctest.Provider, tflambda.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunction_codeSigning(t *testing.T) {
	if got, want := acctest.Partition(), endpoints.AwsUsGovPartitionID; got == want {
		t.Skipf("Lambda code signing config is not supported in %s partition", got)
	}

	// We are hardcoding the region here, because go aws sdk endpoints
	// package does not support Signer service
	for _, want := range []string{endpoints.ApNortheast3RegionID, endpoints.ApSoutheast3RegionID} {
		if got := acctest.Region(); got == want {
			t.Skipf("Lambda code signing config is not supported in %s region", got)
		}
	}

	var conf lambda.GetFunctionOutput
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_csc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_csc_%s", rString)
	resourceName := "aws_lambda_function.test"
	cscResourceName := "aws_lambda_code_signing_config.code_signing_config_1"
	cscUpdateResourceName := "aws_lambda_code_signing_config.code_signing_config_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSignerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCSCCreateConfig(roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttrPair(resourceName, "code_signing_config_arn", cscResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccCSCUpdateConfig(roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttrPair(resourceName, "code_signing_config_arn", cscUpdateResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccCSCDeleteConfig(roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "code_signing_config_arn", ""),
				),
			},
		},
	})
}

func TestAccLambdaFunction_concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_concurrency_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_concurrency_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_concurrency_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_concurrency_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConcurrencyConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "111"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccConcurrencyUpdateConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_concurrencyCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_concurrency_cycle_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_concurrency_cycle_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_concurrency_cycle_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_concurrency_cycle_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccConcurrencyUpdateConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_expectFilenameAndS3Attributes(t *testing.T) {
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_expect_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_expect_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_expect_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_expect_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccWithoutFilenameAndS3AttributesConfig(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile(`filename, s3_\* or image_uri attributes must be set`),
			},
		},
	})
}

func TestAccLambdaFunction_envVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_env_vars_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_env_vars_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_env_vars_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_env_vars_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "environment"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccEnvVariablesConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
				),
			},
			{
				Config: testAccEnvVariablesModifiedConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "baz"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo1", "bar1"),
				),
			},
			{
				Config: testAccEnvVariablesModifiedWithoutEnvironmentConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "environment"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_EnvironmentVariables_noValue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentVariablesNoValueConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.key1", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_encryptedEnvVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	keyDesc := fmt.Sprintf("tf_acc_key_lambda_func_encrypted_env_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_encrypted_env_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_encrypted_env_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_encrypted_env_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_encrypted_env_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptedEnvVariablesConfig(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_arn", "kms", regexp.MustCompile(`key/.+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccEncryptedEnvVariablesModifiedConfig(keyDesc, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
			},
		},
	})
}

func TestAccLambdaFunction_versioned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_versioned_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_versioned_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_versioned_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_versioned_%s", rString)
	resourceName := "aws_lambda_function.test"

	version := "1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishableConfig("test-fixtures/lambdatest.zip", funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, version)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_versionedUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_versioned_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_versioned_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_versioned_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_versioned_%s", rString)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	version := "2"
	versionUpdated := "3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishableConfig("test-fixtures/lambdatest.zip", funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, "1")),
				),
			},
			{
				// Test for changed code, will publish a new version
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccPublishableConfig(path, funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, version)),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
			{
				// Test for changed runtime, will publish a new version
				PreConfig: func() {
					timeBeforeUpdate = time.Now()
				},
				Config: testAccVersionedNodeJs14xRuntimeConfig(path, funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", versionUpdated),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, versionUpdated)),
					resource.TestCheckResourceAttr(resourceName, "runtime", lambda.RuntimeNodejs14X),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_enablePublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf1, conf2, conf3 lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_enable_publish_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_enable_publish_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_enable_publish_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_enable_publish_%s", rString)
	resourceName := "aws_lambda_function.test"
	fileName := "test-fixtures/lambdatest.zip"

	unpublishedVersion := tflambda.FunctionVersionLatest
	publishedVersion := "1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf1),
					testAccCheckFunctionName(&conf1, funcName),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", unpublishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, unpublishedVersion)),
				),
			},
			{
				// No changes, except to `publish`. This should publish a new version.
				Config: testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf2),
					testAccCheckFunctionName(&conf2, funcName),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, publishedVersion)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				// No changes, `publish` is true. This should not publish a new version.
				Config: testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf3),
					testAccCheckFunctionName(&conf3, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, publishedVersion)),
				),
			},
		},
	})
}

func TestAccLambdaFunction_disablePublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf1, conf2 lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_disable_publish_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_disable_publish_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_disable_publish_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_disable_publish_%s", rString)
	resourceName := "aws_lambda_function.test"
	fileName := "test-fixtures/lambdatest.zip"

	publishedVersion := "1"
	unpublishedVersion := publishedVersion // Should remain the last published version

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf1),
					testAccCheckFunctionName(&conf1, funcName),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, publishedVersion)),
				),
			},
			{
				// No changes, except to `publish`. This should not update the current version.
				Config: testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf2),
					testAccCheckFunctionName(&conf2, funcName),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "version", unpublishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", funcName, unpublishedVersion)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_deadLetter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_dlconfig_%s", rString)
	topicName := fmt.Sprintf("tf_acc_topic_lambda_func_dlconfig_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_dlconfig_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_dlconfig_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_dlconfig_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					func(s *terraform.State) error {
						if !strings.HasSuffix(*conf.Configuration.DeadLetterConfig.TargetArn, ":"+topicName) {
							return fmt.Errorf(
								"Expected DeadLetterConfig.TargetArn %s to have suffix %s", *conf.Configuration.DeadLetterConfig.TargetArn, ":"+topicName,
							)
						}
						return nil
					},
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure configuration can be removed
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
				),
			},
		},
	})
}

func TestAccLambdaFunction_deadLetterUpdated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_dlcfg_upd_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_dlcfg_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_dlcfg_upd_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_dlcfg_upd_%s", rString)
	topic1Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_%s", rString)
	topic2Name := fmt.Sprintf("tf_acc_topic_lambda_func_dlcfg_upd_2_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithDeadLetterConfig(funcName, topic1Name, policyName, roleName, sgName),
			},
			{
				Config: testAccWithDeadLetterUpdatedConfig(funcName, topic1Name, topic2Name, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					func(s *terraform.State) error {
						if !strings.HasSuffix(*conf.Configuration.DeadLetterConfig.TargetArn, ":"+topic2Name) {
							return fmt.Errorf(
								"Expected DeadLetterConfig.TargetArn %s to have suffix %s", *conf.Configuration.DeadLetterConfig.TargetArn, ":"+topic2Name,
							)
						}
						return nil
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_nilDeadLetter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_nil_dlcfg_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_nil_dlcfg_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_nil_dlcfg_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_nil_dlcfg_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithNilDeadLetterConfig(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf("nil dead_letter_config supplied for function: %s", funcName)),
			},
		},
	})
}

func TestAccLambdaFunction_fileSystem(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			// Ensure a function with lambda file system configuration can be created
			{
				Config: testAccFileSystemConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.0.local_mount_path", "/mnt/efs"),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure lambda file system configuration can be updated
			{
				Config: testAccFileSystemUpdateConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.0.local_mount_path", "/mnt/lambda"),
				),
			},
			// Ensure lambda file system configuration can be removed
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "0"),
				),
			},
		},
	})
}

func testAccImageLatestV1V2PreCheck(t *testing.T) {
	if (os.Getenv("AWS_LAMBDA_IMAGE_LATEST_ID") == "") || (os.Getenv("AWS_LAMBDA_IMAGE_V1_ID") == "") || (os.Getenv("AWS_LAMBDA_IMAGE_V2_ID") == "") {
		t.Skip("AWS_LAMBDA_IMAGE_LATEST_ID, AWS_LAMBDA_IMAGE_V1_ID and AWS_LAMBDA_IMAGE_V2_ID env vars must be set for Lambda Container Image Support acceptance tests. ")
	}
}

func TestAccLambdaFunction_image(t *testing.T) {
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	imageLatestID := os.Getenv("AWS_LAMBDA_IMAGE_LATEST_ID")
	imageV1ID := os.Getenv("AWS_LAMBDA_IMAGE_V1_ID")
	imageV2ID := os.Getenv("AWS_LAMBDA_IMAGE_V2_ID")

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccImageLatestV1V2PreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			// Ensure a function with lambda image configuration can be created
			{
				Config: testAccImageConfig(funcName, policyName, roleName, sgName, imageLatestID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeImage),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageLatestID),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.entry_point.0", "/bootstrap-with-handler"),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.command.0", "app.lambda_handler"),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.working_directory", "/var/task"),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure lambda image code can be updated
			{
				Config: testAccImageUpdateCodeConfig(funcName, policyName, roleName, sgName, imageV1ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageV1ID),
				),
			},
			// Ensure lambda image config can be updated
			{
				Config: testAccImageUpdateConfig(funcName, policyName, roleName, sgName, imageV2ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageV2ID),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.command.0", "app.another_handler"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architectures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccArchitecturesARM64(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure function's "architectures" attribute can be removed. The actual architecture remains unchanged.
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architecturesUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccArchitecturesARM64(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure function architecture can be updated
			{
				Config: testAccArchitecturesUpdate(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureX8664),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architecturesWithLayer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccArchitecturesARM64WithLayer(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			// Ensure function architecture can be updated
			{
				Config: testAccArchitecturesUpdateWithLayer(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureX8664),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_ephemeralStorage(t *testing.T) {
	var conf lambda.GetFunctionOutput
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_ephemeral_storage_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_ephemeral_storage_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_ephemeral_storage_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_ephemeral_storage_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccWithEphemeralStorage(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size", "1024"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithUpdateEphemeralStorage(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size", "2048"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_tracing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	if got, want := acctest.Partition(), endpoints.AwsUsGovPartitionID; got == want {
		t.Skipf("Lambda tracing config is not supported in %s partition", got)
	}

	var conf lambda.GetFunctionOutput
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_tracing_cfg_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tracing_cfg_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tracing_cfg_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tracing_cfg_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithTracingConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.0.mode", "Active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithTracingUpdatedConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.0.mode", "PassThrough"),
				),
			},
		},
	})
}

// This test is to verify the existing behavior in the Lambda API where the KMS Key ARN
// is not returned if environment variables are not in use. If the API begins saving this
// value and the kms_key_arn check begins failing, the documentation should be updated.
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6366
func TestAccLambdaFunction_KMSKeyARN_noEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function1 lambda.GetFunctionOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSKeyARNNoEnvironmentVariablesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &function1),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_layers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_layer_%s", rString)
	layerName := fmt.Sprintf("tf_acc_layer_lambda_func_layer_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_layer_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_layer_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_layer_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithLayersConfig(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_layersUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_layer_%s", rString)
	layerName := fmt.Sprintf("tf_acc_lambda_layer_%s", rString)
	layer2Name := fmt.Sprintf("tf_acc_lambda_layer2_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithLayersConfig(funcName, layerName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithLayersUpdatedConfig(funcName, layerName, layer2Name, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "2"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_vpc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "vpc_config.0.vpc_id", regexp.MustCompile("^vpc-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_vpcRemoval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCConfig(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccBasicConfig(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_vpcUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_upd_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_upd_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_upd_%s", rString)
	sgName2 := fmt.Sprintf("tf_acc_sg_lambda_func_2nd_vpc_upd_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithVPCUpdatedConfig(funcName, policyName, roleName, sgName, sgName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform/issues/5767
// and https://github.com/hashicorp/terraform/issues/10272
func TestAccLambdaFunction_VPC_withInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_w_invc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_w_invc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_w_invc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_w_invc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccInvokeFunction(&conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

// See https://github.com/hashicorp/terraform-provider-aws/issues/17385
// When the vpc config doesn't change the version shouldn't change
func TestAccLambdaFunction_VPCPublishNo_changes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_w_invc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_w_invc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_w_invc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_w_invc_%s", rString)
	resourceName := "aws_lambda_function.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCPublishConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithVPCPublishConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform-provider-aws/issues/17385
// When the vpc config changes the version should change
func TestAccLambdaFunction_VPCPublishHas_changes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_vpc_w_invc_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_vpc_w_invc_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_vpc_w_invc_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_vpc_w_invc_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,

		Steps: []resource.TestStep{
			{
				Config: testAccWithVPCPublishConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccWithVPCUpdatedPublishConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10044
func TestAccLambdaFunction_VPC_properIAMDependencies(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function lambda.GetFunctionOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCProperIAMDependenciesConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, rName, &function),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", vpcResourceName, "id"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_emptyVPC(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_empty_vpc_config_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_empty_vpc_config_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_empty_vpc_config_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_empty_vpc_config_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWithEmptyVPCConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_s3(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_%s", rString)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccS3Config(bucketName, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"s3_bucket", "s3_key", "publish"},
			},
		},
	})
}

func TestAccLambdaFunction_localUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_local_upd_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_local_upd_%s", rString)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFunctionConfig_local(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
			},
		},
	})
}

func TestAccLambdaFunction_LocalUpdate_nameOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_local_upd_name_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_local_upd_name_%s", rString)
	resourceName := "aws_lambda_function.test"

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	updatedPath, updatedZipFile, err := createTempFile("lambda_localUpdate_name_change")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(updatedPath)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_local_name_only(path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, updatedZipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_local_name_only(updatedPath, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_S3Update_basic(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := sdkacctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-upd-basic-%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_upd_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_upd_basic_%s", rString)
	resourceName := "aws_lambda_function.test"

	key := "lambda-func.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "s3_bucket", "s3_key", "s3_object_version"},
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3(bucketName, key, path, roleName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_S3Update_unversioned(t *testing.T) {
	var conf lambda.GetFunctionOutput

	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	rString := sdkacctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-bucket-lambda-func-s3-upd-unver-%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_func_s3_upd_unver_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_s3_upd_unver_%s", rString)
	resourceName := "aws_lambda_function.test"
	key := "lambda-func.zip"
	key2 := "lambda-func-modified.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "8DPiX+G1l2LQ8hjBkwRchQFf1TSCEvPrYGRKlM9UoyY="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "s3_bucket", "s3_key"},
			},
			{
				PreConfig: func() {
					// Upload 2nd version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_tags(t *testing.T) {
	var conf lambda.GetFunctionOutput

	rString := sdkacctest.RandString(8)
	resourceName := "aws_lambda_function.test"
	funcName := fmt.Sprintf("tf_acc_lambda_func_tags_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_tags_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_tags_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_tags_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBasicConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccTagsConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One"),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "Very interesting"),
				),
			},
			{
				Config: testAccTagsModifiedConfig(funcName, policyName, roleName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(resourceName, funcName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "lambda", fmt.Sprintf("function:%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value One Changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value Two"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value Three"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_runtimes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	steps := []resource.TestStep{
		{
			// Test invalid runtime.
			Config:      testAccRuntimeConfig(rName, rName),
			ExpectError: regexp.MustCompile(`expected runtime to be one of`),
		},
	}
	for _, runtime := range lambda.Runtime_Values() {
		// EOL runtimes.
		switch runtime {
		case lambda.RuntimeNodejs43Edge:
			fallthrough
		case lambda.RuntimeDotnetcore20:
			fallthrough
		case lambda.RuntimeDotnetcore10:
			fallthrough
		case lambda.RuntimeNodejs10X:
			fallthrough
		case lambda.RuntimeNodejs810:
			fallthrough
		case lambda.RuntimeNodejs610:
			fallthrough
		case lambda.RuntimeNodejs43:
			fallthrough
		case lambda.RuntimeNodejs:
			continue
		}

		steps = append(steps, resource.TestStep{
			Config: testAccRuntimeConfig(rName, runtime),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckFunctionExists(resourceName, rName, &v),
				resource.TestCheckResourceAttr(resourceName, "runtime", runtime),
			),
		})
	}
	steps = append(steps, resource.TestStep{
		ResourceName:            resourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"filename", "publish"},
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps:             steps,
	})
}

func TestAccLambdaFunction_Zip_validation(t *testing.T) {
	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_expect_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_expect_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_expect_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_expect_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccZipWithoutHandlerConfig(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
			{
				Config:      testAccZipWithoutRuntimeConfig(funcName, policyName, roleName, sgName),
				ExpectError: regexp.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
		},
	})
}

func testAccCheckFunctionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function" {
			continue
		}

		_, err := conn.GetFunction(&lambda.GetFunctionInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda Function still exists")
		}

	}

	return nil
}

func testAccCheckFunctionDisappears(function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		input := &lambda.DeleteFunctionInput{
			FunctionName: function.Configuration.FunctionName,
		}

		_, err := conn.DeleteFunction(input)

		return err
	}
}

func testAccCheckFunctionExists(res, funcName string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Lambda function not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda function ID not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		params := &lambda.GetFunctionInput{
			FunctionName: aws.String(funcName),
		}

		getFunction, err := conn.GetFunction(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckFunctionInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.StringValue(function.Configuration.FunctionArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccInvokeFunction(function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f := function.Configuration
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		// If the function is VPC-enabled this will create ENI automatically
		_, err := conn.Invoke(&lambda.InvokeInput{
			FunctionName: f.FunctionName,
		})

		return err
	}
}

func testAccCheckFunctionName(function *lambda.GetFunctionOutput, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.FunctionName != expectedName {
			return fmt.Errorf("Expected function name %s, got %s", expectedName, *c.FunctionName)
		}

		return nil
	}
}

// Rename to correctly identify as using API values
func testAccCheckFunctionVersion(function *lambda.GetFunctionOutput, expectedVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.Version != expectedVersion {
			return fmt.Errorf("Expected version %s, got %s", expectedVersion, *c.Version)
		}
		return nil
	}
}

func testAccCheckSourceCodeHash(function *lambda.GetFunctionOutput, expectedHash string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := function.Configuration
		if *c.CodeSha256 != expectedHash {
			return fmt.Errorf("Expected code hash %s, got %s", expectedHash, *c.CodeSha256)
		}

		return nil
	}
}

func testAccCheckAttributeIsDateAfter(s *terraform.State, name string, key string, before time.Time) error {
	rs, ok := s.RootModule().Resources[name]
	if !ok {
		return fmt.Errorf("Resource %s not found", name)
	}

	v, ok := rs.Primary.Attributes[key]
	if !ok {
		return fmt.Errorf("%s: Attribute '%s' not found", name, key)
	}

	const ISO8601UTC = "2006-01-02T15:04:05Z0700"
	timeValue, err := time.Parse(ISO8601UTC, v)
	if err != nil {
		return err
	}

	if !before.Before(timeValue) {
		return fmt.Errorf("Expected time attribute %s.%s with value %s was not before %s", name, key, v, before.Format(ISO8601UTC))
	}

	return nil
}

func testAccCreateZipFromFiles(files map[string]string, zipFile *os.File) error {
	if err := zipFile.Truncate(0); err != nil {
		return err
	}
	if _, err := zipFile.Seek(0, 0); err != nil {
		return err
	}

	w := zip.NewWriter(zipFile)

	for source, destination := range files {
		f, err := w.Create(destination)
		if err != nil {
			return err
		}

		fileContent, err := os.ReadFile(source)
		if err != nil {
			return err
		}

		_, err = f.Write(fileContent)
		if err != nil {
			return err
		}
	}

	err := w.Close()
	if err != nil {
		return err
	}

	return w.Flush()
}

func createTempFile(prefix string) (string, *os.File, error) {
	f, err := os.CreateTemp(os.TempDir(), prefix)
	if err != nil {
		return "", nil, err
	}

	pathToFile, err := filepath.Abs(f.Name())
	if err != nil {
		return "", nil, err
	}
	return pathToFile, f, nil
}

func testAccBasicConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}

func testAccCSCBasicConfig(roleName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "policy" {
  statement {
    sid    = ""
    effect = "Allow"

    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "iam_for_lambda" {
  name               = "%s"
  assume_role_policy = data.aws_iam_policy_document.policy.json
}

resource "aws_signer_signing_profile" "test1" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test2" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test3" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "test4" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "code_signing_config_1" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test1.version_arn,
      aws_signer_signing_profile.test2.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  description = "Code Signing Config for test account"
}

resource "aws_lambda_code_signing_config" "code_signing_config_2" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test3.version_arn,
      aws_signer_signing_profile.test4.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }

  description = "Code Signing Config for test account update"
}
`, roleName)
}

func testAccCSCCreateConfig(roleName, funcName string) string {
	return fmt.Sprintf(testAccCSCBasicConfig(roleName)+`
resource "aws_lambda_function" "test" {
  filename                = "test-fixtures/lambdatest.zip"
  function_name           = "%s"
  role                    = aws_iam_role.iam_for_lambda.arn
  handler                 = "exports.example"
  runtime                 = "nodejs12.x"
  code_signing_config_arn = aws_lambda_code_signing_config.code_signing_config_1.arn
}
`, funcName)
}

func testAccCSCUpdateConfig(roleName, funcName string) string {
	return fmt.Sprintf(testAccCSCBasicConfig(roleName)+`
resource "aws_lambda_function" "test" {
  filename                = "test-fixtures/lambdatest.zip"
  function_name           = "%s"
  role                    = aws_iam_role.iam_for_lambda.arn
  handler                 = "exports.example"
  runtime                 = "nodejs12.x"
  code_signing_config_arn = aws_lambda_code_signing_config.code_signing_config_2.arn
}
`, funcName)
}

func testAccCSCDeleteConfig(roleName, funcName string) string {
	return fmt.Sprintf(testAccCSCBasicConfig(roleName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}

func testAccBasicConcurrencyConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename                       = "test-fixtures/lambdatest.zip"
  function_name                  = "%s"
  role                           = aws_iam_role.iam_for_lambda.arn
  handler                        = "exports.example"
  runtime                        = "nodejs12.x"
  reserved_concurrent_executions = 111
}
`, funcName)
}

func testAccConcurrencyUpdateConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename                       = "test-fixtures/lambdatest.zip"
  function_name                  = "%s"
  role                           = aws_iam_role.iam_for_lambda.arn
  handler                        = "exports.example"
  runtime                        = "nodejs12.x"
  reserved_concurrent_executions = 222
}
`, funcName)
}

func testAccWithoutFilenameAndS3AttributesConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}

func testAccEnvVariablesConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, funcName)
}

func testAccEnvVariablesModifiedConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  environment {
    variables = {
      foo  = "baz"
      foo1 = "bar1"
    }
  }
}
`, funcName)
}

func testAccEnvVariablesModifiedWithoutEnvironmentConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}

func testAccEnvironmentVariablesNoValueConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"

  environment {
    variables = {
      key1 = ""
    }
  }
}
`, rName))
}

func testAccEncryptedEnvVariablesConfig(keyDesc, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_kms_key" "foo" {
  description = "%s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  kms_key_arn   = aws_kms_key.foo.arn
  runtime       = "nodejs12.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, keyDesc, funcName)
}

func testAccEncryptedEnvVariablesModifiedConfig(keyDesc, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_kms_key" "foo" {
  description = "%s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, keyDesc, funcName)
}

func testAccFilenameConfig(fileName, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  publish       = false
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, fileName, funcName)
}

func testAccPublishableConfig(fileName, funcName, policyName, roleName, sgName string, publish bool) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "%s"
  function_name = "%s"
  publish       = %t
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, fileName, funcName, publish)
}

func testAccFileSystemConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_efs_file_system" "efs_for_lambda" {
  tags = {
    Name = "efs_for_lambda"
  }
}

resource "aws_efs_mount_target" "mount_target_az1" {
  file_system_id  = aws_efs_file_system.efs_for_lambda.id
  subnet_id       = aws_subnet.subnet_for_lambda.id
  security_groups = [aws_security_group.sg_for_lambda.id]
}

resource "aws_efs_access_point" "access_point_1" {
  file_system_id = aws_efs_file_system.efs_for_lambda.id

  root_directory {
    path = "/lambda1"

    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "777"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.access_point_1.arn
    local_mount_path = "/mnt/efs"
  }

  depends_on = [aws_efs_mount_target.mount_target_az1]
}
`, funcName)
}

func testAccFileSystemUpdateConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_efs_file_system" "efs_for_lambda" {
  tags = {
    Name = "efs_for_lambda"
  }
}

resource "aws_efs_mount_target" "mount_target_az2" {
  file_system_id  = aws_efs_file_system.efs_for_lambda.id
  subnet_id       = aws_subnet.subnet_for_lambda_az2.id
  security_groups = [aws_security_group.sg_for_lambda.id]
}

resource "aws_efs_access_point" "access_point_2" {
  file_system_id = aws_efs_file_system.efs_for_lambda.id

  root_directory {
    path = "/lambda2"

    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "777"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda_az2.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.access_point_2.arn
    local_mount_path = "/mnt/lambda"
  }

  depends_on = [aws_efs_mount_target.mount_target_az2]
}
`, funcName)
}

func testAccImageConfig(funcName, policyName, roleName, sgName, imageID string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  image_uri     = "%s"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  image_config {
    entry_point       = ["/bootstrap-with-handler"]
    command           = ["app.lambda_handler"]
    working_directory = "/var/task"
  }
}
`, imageID, funcName)
}

func testAccImageUpdateCodeConfig(funcName, policyName, roleName, sgName, imageID string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  image_uri     = "%s"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  publish       = true
}
`, imageID, funcName)
}

func testAccImageUpdateConfig(funcName, policyName, roleName, sgName, imageID string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  image_uri     = "%s"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  image_config {
    command = ["app.another_handler"]
  }
}
`, imageID, funcName)
}

func testAccArchitecturesARM64(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
}
`, funcName)
}

func testAccArchitecturesUpdate(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  architectures = ["x86_64"]
}
`, funcName)
}

func testAccArchitecturesARM64WithLayer(funcName, layerName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = "%[1]s"
  compatible_runtimes      = ["nodejs12.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[2]s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  architectures = ["arm64"]
  layers        = [aws_lambda_layer_version.test.arn]
}
`, layerName, funcName)
}

func testAccArchitecturesUpdateWithLayer(funcName, layerName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = "%[1]s"
  compatible_runtimes      = ["nodejs12.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[2]s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  architectures = ["x86_64"]
  layers        = [aws_lambda_layer_version.test.arn]
}
`, layerName, funcName)
}

func testAccVersionedNodeJs14xRuntimeConfig(fileName, funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "%s"
  function_name = "%s"
  publish       = true
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs14.x"
}
`, fileName, funcName)
}

func testAccVPCProperIAMDependenciesConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
  role       = aws_iam_role.test.id
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

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName)
}

func testAccWithTracingConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  tracing_config {
    mode = "Active"
  }
}
`, funcName)
}

func testAccWithTracingUpdatedConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  tracing_config {
    mode = "PassThrough"
  }
}
`, funcName)
}

func testAccWithDeadLetterConfig(funcName, topicName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  dead_letter_config {
    target_arn = aws_sns_topic.test.arn
  }
}

resource "aws_sns_topic" "test" {
  name = "%s"
}
`, funcName, topicName)
}

func testAccWithDeadLetterUpdatedConfig(funcName, topic1Name, topic2Name, policyName,
	roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  dead_letter_config {
    target_arn = aws_sns_topic.test_2.arn
  }
}

resource "aws_sns_topic" "test" {
  name = "%s"
}

resource "aws_sns_topic" "test_2" {
  name = "%s"
}
`, funcName, topic1Name, topic2Name)
}

func testAccWithNilDeadLetterConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  dead_letter_config {
    target_arn = ""
  }
}
`, funcName)
}

func testAccKMSKeyARNNoEnvironmentVariablesConfig(rName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(rName, rName, rName)+`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  kms_key_arn   = aws_kms_key.test.arn
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccWithLayersConfig(funcName, layerName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = "%s"
  compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  layers        = [aws_lambda_layer_version.test.arn]
}
`, layerName, funcName)
}

func testAccWithLayersUpdatedConfig(funcName, layerName, layer2Name, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_layer_version" "test" {
  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = "%s"
  compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_layer_version" "test_2" {
  filename            = "test-fixtures/lambdatest_modified.zip"
  layer_name          = "%s"
  compatible_runtimes = ["nodejs12.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  layers = [
    aws_lambda_layer_version.test.arn,
    aws_lambda_layer_version.test_2.arn,
  ]
}
`, layerName, layer2Name, funcName)
}

func testAccWithVPCConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }
}
`, funcName)
}

func testAccWithVPCPublishConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  publish       = true
  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }
}
`, funcName)
}

func testAccWithVPCUpdatedPublishConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
  publish       = true
  vpc_config {
    security_group_ids = []
    subnet_ids         = []
  }
}
`, funcName)
}

func testAccWithVPCUpdatedConfig(funcName, policyName, roleName, sgName, sgName2 string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id, aws_subnet.subnet_for_lambda_az2.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id, aws_security_group.sg_for_lambda_2.id]
  }
}

resource "aws_security_group" "sg_for_lambda_2" {
  name        = "sg_for_lambda_%s"
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = aws_vpc.vpc_for_lambda.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, funcName, sgName2)
}

func testAccWithEmptyVPCConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  vpc_config {
    subnet_ids         = []
    security_group_ids = []
  }
}
`, funcName)
}

func testAccS3Config(bucketName, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = "%s"
}

resource "aws_s3_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.id
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
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

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_bucket.lambda_bucket.id
  s3_key        = aws_s3_object.lambda_code.id
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, bucketName, roleName, funcName)
}

func testAccTagsConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  tags = {
    Key1        = "Value One"
    Description = "Very interesting"
  }
}
`, funcName)
}

func testAccTagsModifiedConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  tags = {
    Key1 = "Value One Changed"
    Key2 = "Value Two"
    Key3 = "Value Three"
  }
}
`, funcName)
}

func testAccFunctionConfig_local(filePath, roleName, funcName string) string {
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

resource "aws_lambda_function" "test" {
  filename         = "%s"
  source_code_hash = filebase64sha256("%s")
  function_name    = "%s"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}
`, roleName, filePath, filePath, funcName)
}

func testAccFunctionConfig_local_name_only(filePath, roleName, funcName string) string {
	return testAccFunctionConfig_local_name_only_tpl(filePath, roleName, funcName)
}

func testAccFunctionConfig_local_name_only_tpl(filePath, roleName, funcName string) string {
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

resource "aws_lambda_function" "test" {
  filename      = "%s"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, roleName, filePath, funcName)
}

func testAccFunctionConfig_s3(bucketName, key, path, roleName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_acl" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  acl    = "private"
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "o" {
  # Must have versioning enabled first
  depends_on = [aws_s3_bucket_versioning.artifacts]

  bucket = aws_s3_bucket.artifacts.bucket
  key    = "%s"
  source = "%s"
  etag   = filemd5("%s")
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

resource "aws_lambda_function" "test" {
  s3_bucket         = aws_s3_object.o.bucket
  s3_key            = aws_s3_object.o.key
  s3_object_version = aws_s3_object.o.version_id
  function_name     = "%s"
  role              = aws_iam_role.iam_for_lambda.arn
  handler           = "exports.example"
  runtime           = "nodejs12.x"
}
`, bucketName, key, path, path, roleName, funcName)
}

func testAccFunctionConfig_s3_unversioned_tpl(bucketName, roleName, funcName, key, path string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = "%s"
  force_destroy = true
}

resource "aws_s3_bucket_acl" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  acl    = "private"
}

resource "aws_s3_object" "o" {
  bucket = aws_s3_bucket.artifacts.bucket
  key    = "%s"
  source = "%s"
  etag   = filemd5("%s")
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

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_object.o.bucket
  s3_key        = aws_s3_object.o.key
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, bucketName, key, path, path, roleName, funcName)
}

func testAccRuntimeConfig(rName, runtime string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = %[2]q
}
`, rName, runtime))
}

func testAccZipWithoutHandlerConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs12.x"
}
`, funcName)
}

func testAccZipWithoutRuntimeConfig(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
}
`, funcName)
}

func TestFlattenImageConfigShouldNotFailWithEmptyImageConfig(t *testing.T) {
	t.Parallel()
	response := lambda.ImageConfigResponse{}
	tflambda.FlattenImageConfig(&response)
}

func testAccPreCheckSignerSigningProfile(t *testing.T, platformID string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn

	var foundPlatform bool
	err := conn.ListSigningPlatformsPages(&signer.ListSigningPlatformsInput{}, func(page *signer.ListSigningPlatformsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, platform := range page.Platforms {
			if platform == nil {
				continue
			}

			if aws.StringValue(platform.PlatformId) == platformID {
				foundPlatform = true

				return false
			}
		}

		return !lastPage
	})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if !foundPlatform {
		t.Skipf("skipping acceptance testing: Signing Platform (%s) not found", platformID)
	}
}

func testAccWithEphemeralStorage(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  ephemeral_storage {
    size = 1024
  }
}
`, funcName)
}

func testAccWithUpdateEphemeralStorage(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"

  ephemeral_storage {
    size = 2048
  }
}
`, funcName)
}
