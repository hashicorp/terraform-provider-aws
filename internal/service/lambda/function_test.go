package lambda_test

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_func_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_func_basic_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_func_basic_%s", rString)
	sgName := fmt.Sprintf("tf_acc_sg_lambda_func_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
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

func TestAccLambdaFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaFunction_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFunctionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_unpublishedCodeUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	initialFilename := "test-fixtures/lambdatest.zip"
	updatedFilename, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(updatedFilename)

	var conf1, conf2 lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_filename(initialFilename, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "version", tflambda.FunctionVersionLatest),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, tflambda.FunctionVersionLatest)),
				),
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFunctionConfig_filename(updatedFilename, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "version", tflambda.FunctionVersionLatest),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, tflambda.FunctionVersionLatest)),
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

func TestAccLambdaFunction_codeSigning(t *testing.T) {
	ctx := acctest.Context(t)
	if curr := acctest.Region(); !tflambda.SignerServiceIsAvailable(curr) {
		t.Skipf("Lambda code signing config is not supported in %s region", curr)
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	cscResourceName := "aws_lambda_code_signing_config.code_signing_config_1"
	cscUpdateResourceName := "aws_lambda_code_signing_config.code_signing_config_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckSignerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA") },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_cscCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_cscUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_cscDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "code_signing_config_arn", ""),
				),
			},
		},
	})
}

func TestAccLambdaFunction_concurrency(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basicConcurrency(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_concurrencyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_concurrencyCycle(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_concurrencyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "222"),
				),
			},
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_expectFilenameAndS3Attributes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_noFilenameAndS3Attributes(rName),
				ExpectError: regexp.MustCompile("one of `filename,image_uri,s3_bucket` must be specified"),
			},
		},
	})
}

func TestAccLambdaFunction_envVariables(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_envVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
				),
			},
			{
				Config: testAccFunctionConfig_envVariablesModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "baz"),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo1", "bar1"),
				),
			},
			{
				Config: testAccFunctionConfig_envVariablesModifiedNoEnvironment(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_EnvironmentVariables_noValue(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_environmentVariablesNoValue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	kmsKey1ResourceName := "aws_kms_key.test1"
	kmsKey2ResourceName := "aws_kms_key.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_encryptedEnvVariablesKey1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKey1ResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_encryptedEnvVariablesKey2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", kmsKey2ResourceName, "arn"),
				),
			},
			{
				Config: testAccFunctionConfig_envVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
				),
			},
		},
	})
}

func TestAccLambdaFunction_nameValidation(t *testing.T) {
	ctx := acctest.Context(t)
	badFuncName := "prefix.viewer_request_lambda"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_basic(badFuncName, rName, rName, rName),
				ExpectError: regexp.MustCompile(`invalid value for function_name \(must be valid function name or function ARN\)`),
			},
		},
	})
}

func TestAccLambdaFunction_versioned(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	version := "1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable("test-fixtures/lambdatest.zip", rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, version)),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	version := "2"
	versionUpdated := "3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable("test-fixtures/lambdatest.zip", rName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, "1")),
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
				Config: testAccFunctionConfig_publishable(path, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, version)),
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
				Config: testAccFunctionConfig_versionedNodeJs14xRuntime(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", versionUpdated),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, versionUpdated)),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf1, conf2, conf3 lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	fileName := "test-fixtures/lambdatest.zip"

	unpublishedVersion := tflambda.FunctionVersionLatest
	publishedVersion := "1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable(fileName, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "version", unpublishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, unpublishedVersion)),
				),
			},
			{
				// No changes, except to `publish`. This should publish a new version.
				Config: testAccFunctionConfig_publishable(fileName, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, publishedVersion)),
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
				Config: testAccFunctionConfig_publishable(fileName, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf3),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, publishedVersion)),
				),
			},
		},
	})
}

func TestAccLambdaFunction_disablePublish(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf1, conf2 lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	fileName := "test-fixtures/lambdatest.zip"

	publishedVersion := "1"
	unpublishedVersion := publishedVersion // Should remain the last published version

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable(fileName, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "version", publishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, publishedVersion)),
				),
			},
			{
				// No changes, except to `publish`. This should not update the current version.
				Config: testAccFunctionConfig_publishable(fileName, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "version", unpublishedVersion),
					acctest.CheckResourceAttrRegionalARN(resourceName, "qualified_arn", "lambda", fmt.Sprintf("function:%s:%s", rName, unpublishedVersion)),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_deadLetter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.0", "arn"),
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
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_deadLetterUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_deadLetter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.0", "arn"),
				),
			},
			{
				Config: testAccFunctionConfig_deadLetterUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.1", "arn"),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_nilDeadLetter(rName),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf("nil dead_letter_config supplied for function: %s", rName)),
			},
		},
	})
}

func TestAccLambdaFunction_fileSystem(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure a function with lambda file system configuration can be created
			{
				Config: testAccFunctionConfig_fileSystem(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_config.0.arn", "aws_efs_access_point.test1", "arn"),
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
				Config: testAccFunctionConfig_fileSystemUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_config.0.arn", "aws_efs_access_point.test2", "arn"),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.0.local_mount_path", "/mnt/lambda"),
				),
			},
			// Ensure lambda file system configuration can be removed
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "file_system_config.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_image(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AWS_LAMBDA_IMAGE_LATEST_ID"
	imageLatestID := os.Getenv(key)
	if imageLatestID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	key = "AWS_LAMBDA_IMAGE_V1_ID"
	imageV1ID := os.Getenv(key)
	if imageV1ID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	key = "AWS_LAMBDA_IMAGE_V2_ID"
	imageV2ID := os.Getenv(key)
	if imageV2ID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure a function with lambda image configuration can be created
			{
				Config: testAccFunctionConfig_image(rName, imageLatestID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
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
				Config: testAccFunctionConfig_imageUpdateCode(rName, imageV1ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageV1ID),
				),
			},
			// Ensure lambda image config can be updated
			{
				Config: testAccFunctionConfig_imageUpdate(rName, imageV2ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageV2ID),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.command.0", "app.another_handler"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architectures(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccFunctionConfig_architecturesARM64(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
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
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architecturesUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccFunctionConfig_architecturesARM64(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
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
				Config: testAccFunctionConfig_architecturesUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureX8664),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
				),
			},
		},
	})
}

func TestAccLambdaFunction_architecturesWithLayer(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure function with arm64 architecture can be created
			{
				Config: testAccFunctionConfig_architecturesARM64Layer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureArm64),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
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
				Config: testAccFunctionConfig_architecturesUpdateLayer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", lambda.ArchitectureX8664),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", lambda.PackageTypeZip),
				),
			},
		},
	})
}

func TestAccLambdaFunction_ephemeralStorage(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_ephemeralStorage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_updateEphemeralStorage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size", "2048"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_tracing(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_tracing(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.#", "1"),
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
				Config: testAccFunctionConfig_tracingUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tracing_config.#", "1"),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function1 lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_kmsKeyARNNoEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function1),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_layers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_layers(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_layersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "2"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_config.0.vpc_id"),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_vpcUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_config.0.vpc_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_vpcUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionVersion(&conf, tflambda.FunctionVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_config.0.vpc_id"),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform/issues/5767
// and https://github.com/hashicorp/terraform/issues/10272
func TestAccLambdaFunction_VPC_withInvocation(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccInvokeFunction(ctx, &conf),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_vpcPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

// See https://github.com/hashicorp/terraform-provider-aws/issues/17385
// When the vpc config changes the version should change
func TestAccLambdaFunction_VPCPublishHas_changes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_vpcUpdatedPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10044
func TestAccLambdaFunction_VPC_properIAMDependencies(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcProperIAMDependencies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function),
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
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_emptyVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_s3Simple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	path, zipFile, err := createTempFile("lambda_localUpdate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	var timeBeforeUpdate time.Time

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_local(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_local(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_localNameOnly(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_localNameOnly(updatedPath, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_S3Update_basic(t *testing.T) {
	ctx := acctest.Context(t)
	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	key := "lambda-func.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3(key, path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_s3(key, path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_S3Update_unversioned(t *testing.T) {
	ctx := acctest.Context(t)
	path, zipFile, err := createTempFile("lambda_s3Update")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	key := "lambda-func.zip"
	key2 := "lambda-func-modified.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Upload 1st version
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_s3UnversionedTPL(rName, key, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
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
				Config: testAccFunctionConfig_s3UnversionedTPL(rName, key2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "0tdaP9H9hsk9c2CycSwOG/sa/x5JyAmSYunA/ce99Pg="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_snapStart(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_snapStartEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "snap_start.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "snap_start.0.apply_on", "PublishedVersions"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_snapStartDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "snap_start.#", "0"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_runtimes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	steps := []resource.TestStep{
		{
			// Test invalid runtime.
			Config:      testAccFunctionConfig_runtime(rName, rName),
			ExpectError: regexp.MustCompile(`expected runtime to be one of`),
		},
	}
	for _, runtime := range lambda.Runtime_Values() {
		// EOL runtimes.
		// https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtime-support-policy.
		switch runtime {
		case lambda.RuntimeDotnetcore21:
			fallthrough
		case lambda.RuntimePython27:
			fallthrough
		case lambda.RuntimePython36:
			fallthrough
		case lambda.RuntimeRuby25:
			fallthrough
		case lambda.RuntimeNodejs10X:
			fallthrough
		case lambda.RuntimeNodejs810:
			fallthrough
		case lambda.RuntimeNodejs610:
			fallthrough
		case lambda.RuntimeNodejs43Edge:
			fallthrough
		case lambda.RuntimeNodejs43:
			fallthrough
		case lambda.RuntimeNodejs:
			fallthrough
		case lambda.RuntimeDotnetcore20:
			fallthrough
		case lambda.RuntimeDotnetcore10:
			continue
		}

		steps = append(steps, resource.TestStep{
			Config: testAccFunctionConfig_runtime(rName, runtime),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckFunctionExists(ctx, resourceName, &v),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps:                    steps,
	})
}

func TestAccLambdaFunction_Zip_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, lambda.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_zipNoHandler(rName),
				ExpectError: regexp.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
			{
				Config:      testAccFunctionConfig_zipNoRuntime(rName),
				ExpectError: regexp.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
		},
	})
}

func testAccCheckFunctionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function" {
				continue
			}

			_, err := tflambda.FindFunctionByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Function %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFunctionExists(ctx context.Context, n string, v *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lambda Function ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn()

		output, err := tflambda.FindFunctionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFunctionQualifiedInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		qualifiedArn := fmt.Sprintf("%s:%s", aws.StringValue(function.Configuration.FunctionArn), aws.StringValue(function.Configuration.Version))
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "qualified_invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", qualifiedArn))(s)
	}
}

func testAccCheckFunctionInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.StringValue(function.Configuration.FunctionArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccInvokeFunction(ctx context.Context, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f := function.Configuration
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn()

		// If the function is VPC-enabled this will create ENI automatically
		_, err := conn.InvokeWithContext(ctx, &lambda.InvokeInput{
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

func testAccFunctionConfig_basic(funcName, policyName, roleName, sgName string) string {
	return fmt.Sprintf(acctest.ConfigLambdaBase(policyName, roleName, sgName)+`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, funcName)
}

func testAccFunctionConfig_snapStartEnabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda_java11.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "example.Hello::handleRequest"
  runtime       = "java11"

  snap_start {
    apply_on = "PublishedVersions"
  }
}
`, rName))
}

func testAccFunctionConfig_snapStartDisabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambda_java11.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "example.Hello::handleRequest"
  runtime       = "java11"
}
`, rName))
}

func testAccFunctionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccFunctionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccFunctionConfig_filename(fileName, rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  publish       = false
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, fileName, rName))
}

func testAccFunctionConfig_cscBase(rName string) string {
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
  name               = %[1]q
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
`, rName)
}

func testAccFunctionConfig_cscCreate(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfig_cscBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename                = "test-fixtures/lambdatest.zip"
  function_name           = %[1]q
  role                    = aws_iam_role.iam_for_lambda.arn
  handler                 = "exports.example"
  runtime                 = "nodejs16.x"
  code_signing_config_arn = aws_lambda_code_signing_config.code_signing_config_1.arn
}
`, rName))
}

func testAccFunctionConfig_cscUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfig_cscBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename                = "test-fixtures/lambdatest.zip"
  function_name           = %[1]q
  role                    = aws_iam_role.iam_for_lambda.arn
  handler                 = "exports.example"
  runtime                 = "nodejs16.x"
  code_signing_config_arn = aws_lambda_code_signing_config.code_signing_config_2.arn
}
`, rName))
}

func testAccFunctionConfig_cscDelete(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfig_cscBase(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccFunctionConfig_basicConcurrency(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename                       = "test-fixtures/lambdatest.zip"
  function_name                  = %[1]q
  role                           = aws_iam_role.iam_for_lambda.arn
  handler                        = "exports.example"
  runtime                        = "nodejs16.x"
  reserved_concurrent_executions = 111
}
`, rName))
}

func testAccFunctionConfig_concurrencyUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename                       = "test-fixtures/lambdatest.zip"
  function_name                  = %[1]q
  role                           = aws_iam_role.iam_for_lambda.arn
  handler                        = "exports.example"
  runtime                        = "nodejs16.x"
  reserved_concurrent_executions = 222
}
`, rName))
}

func testAccFunctionConfig_noFilenameAndS3Attributes(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccFunctionConfig_envVariables(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, rName))
}

func testAccFunctionConfig_envVariablesModified(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  environment {
    variables = {
      foo  = "baz"
      foo1 = "bar1"
    }
  }
}
`, rName))
}

func testAccFunctionConfig_envVariablesModifiedNoEnvironment(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccFunctionConfig_environmentVariablesNoValue(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"

  environment {
    variables = {
      key1 = ""
    }
  }
}
`, rName))
}

func testAccFunctionConfig_encryptedEnvVariablesKey1(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  description = "%[1]s-1"

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

resource "aws_kms_key" "test2" {
  description = "%[1]s-2"

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
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  kms_key_arn   = aws_kms_key.test1.arn
  runtime       = "nodejs16.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, rName))
}

func testAccFunctionConfig_encryptedEnvVariablesKey2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
# Delete aws_kms_key.test1.

resource "aws_kms_key" "test2" {
  description = "%[1]s-2"

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
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  kms_key_arn   = aws_kms_key.test2.arn
  runtime       = "nodejs16.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}
`, rName))
}

func testAccFunctionConfig_publishable(fileName, rName string, publish bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  publish       = %[3]t
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, fileName, rName, publish))
}

func testAccFunctionConfig_versionedNodeJs14xRuntime(fileName, rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  publish       = true
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs14.x"
}
`, fileName, rName))
}

func testAccFunctionConfig_deadLetter(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  dead_letter_config {
    target_arn = aws_sns_topic.test[0].arn
  }
}

resource "aws_sns_topic" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}
`, rName))
}

func testAccFunctionConfig_deadLetterUpdated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  dead_letter_config {
    target_arn = aws_sns_topic.test[1].arn
  }
}

resource "aws_sns_topic" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}
`, rName))
}

func testAccFunctionConfig_nilDeadLetter(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  dead_letter_config {
    target_arn = ""
  }
}
`, rName))
}

func testAccFunctionConfig_fileSystem(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test1" {
  file_system_id  = aws_efs_file_system.test.id
  subnet_id       = aws_subnet.subnet_for_lambda.id
  security_groups = [aws_security_group.sg_for_lambda.id]
}

resource "aws_efs_access_point" "test1" {
  file_system_id = aws_efs_file_system.test.id

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
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.test1.arn
    local_mount_path = "/mnt/efs"
  }

  depends_on = [aws_efs_mount_target.test1]
}
`, rName))
}

func testAccFunctionConfig_fileSystemUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test2" {
  file_system_id  = aws_efs_file_system.test.id
  subnet_id       = aws_subnet.subnet_for_lambda_az2.id
  security_groups = [aws_security_group.sg_for_lambda.id]
}

resource "aws_efs_access_point" "test2" {
  file_system_id = aws_efs_file_system.test.id

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
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda_az2.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.test2.arn
    local_mount_path = "/mnt/lambda"
  }

  depends_on = [aws_efs_mount_target.test2]
}
`, rName))
}

func testAccFunctionConfig_image(rName, imageID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  image_uri     = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  image_config {
    entry_point       = ["/bootstrap-with-handler"]
    command           = ["app.lambda_handler"]
    working_directory = "/var/task"
  }
}
`, imageID, rName))
}

func testAccFunctionConfig_imageUpdateCode(rName, imageID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  image_uri     = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  publish       = true
}
`, imageID, rName))
}

func testAccFunctionConfig_imageUpdate(rName, imageID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  image_uri     = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  image_config {
    command = ["app.another_handler"]
  }
}
`, imageID, rName))
}

func testAccFunctionConfig_architecturesARM64(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  architectures = ["arm64"]
}
`, rName))
}

func testAccFunctionConfig_architecturesUpdate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  architectures = ["x86_64"]
}
`, rName))
}

func testAccFunctionConfig_architecturesARM64Layer(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs16.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  architectures = ["arm64"]
  layers        = [aws_lambda_layer_version.test.arn]
}
`, rName))
}

func testAccFunctionConfig_architecturesUpdateLayer(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  filename                 = "test-fixtures/lambdatest.zip"
  layer_name               = %[1]q
  compatible_runtimes      = ["nodejs16.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  architectures = ["x86_64"]
  layers        = [aws_lambda_layer_version.test.arn]
}
`, rName))
}

func testAccFunctionConfig_ephemeralStorage(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  ephemeral_storage {
    size = 1024
  }
}
`, rName))
}

func testAccFunctionConfig_updateEphemeralStorage(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  ephemeral_storage {
    size = 2048
  }
}
`, rName))
}

func testAccFunctionConfig_tracing(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  tracing_config {
    mode = "Active"
  }
}
`, rName))
}

func testAccFunctionConfig_tracingUpdated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  tracing_config {
    mode = "PassThrough"
  }
}
`, rName))
}

func testAccFunctionConfig_kmsKeyARNNoEnvironmentVariables(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
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
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccFunctionConfig_layers(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  count = 2

  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = "%[1]s-${count.index}"
  compatible_runtimes = ["nodejs16.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  layers        = [aws_lambda_layer_version.test[0].arn]
}
`, rName))
}

func testAccFunctionConfig_layersUpdated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_layer_version" "test" {
  count = 2

  filename            = "test-fixtures/lambdatest.zip"
  layer_name          = "%[1]s-${count.index}"
  compatible_runtimes = ["nodejs16.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  layers        = aws_lambda_layer_version.test[*].arn
}
`, rName))
}

func testAccFunctionConfig_vpc(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }
}
`, rName))
}

func testAccFunctionConfig_vpcUpdated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id, aws_subnet.subnet_for_lambda_az2.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id, aws_security_group.sg_for_lambda_2.id]
  }
}

resource "aws_security_group" "sg_for_lambda_2" {
  name        = "%[1]s-2"
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
`, rName))
}

func testAccFunctionConfig_vpcPublish(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  publish       = true
  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }
}
`, rName))
}

func testAccFunctionConfig_vpcUpdatedPublish(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
  publish       = true
  vpc_config {
    security_group_ids = []
    subnet_ids         = []
  }
}
`, rName))
}

func testAccFunctionConfig_vpcProperIAMDependencies(rName string) string {
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
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName)
}

func testAccFunctionConfig_emptyVPC(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"

  vpc_config {
    subnet_ids         = []
    security_group_ids = []
  }
}
`, rName))
}

func testAccFunctionConfig_s3Simple(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = %[1]q
}

resource "aws_s3_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.id
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_iam_role" "iam_for_lambda" {
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

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_bucket.lambda_bucket.id
  s3_key        = aws_s3_object.lambda_code.id
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName)
}

func testAccFunctionConfig_local(filePath, rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = %[2]q

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
  filename         = %[1]q
  source_code_hash = filebase64sha256(%[1]q)
  function_name    = %[2]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs16.x"
}
`, filePath, rName)
}

func testAccFunctionConfig_localNameOnly(filePath, rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = %[2]q

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
  filename      = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, filePath, rName)
}

func testAccFunctionConfig_s3(key, path, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = %[3]q
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
  key    = %[1]q
  source = %[2]q
  etag   = filemd5(%[2]q)
}

resource "aws_iam_role" "iam_for_lambda" {
  name = %[3]q

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
  function_name     = %[3]q
  role              = aws_iam_role.iam_for_lambda.arn
  handler           = "exports.example"
  runtime           = "nodejs16.x"
}
`, key, path, rName)
}

func testAccFunctionConfig_s3UnversionedTPL(rName, key, path string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  acl    = "private"
}

resource "aws_s3_object" "o" {
  bucket = aws_s3_bucket.artifacts.bucket
  key    = %[2]q
  source = %[3]q
  etag   = filemd5(%[3]q)
}

resource "aws_iam_role" "iam_for_lambda" {
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

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_object.o.bucket
  s3_key        = aws_s3_object.o.key
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs16.x"
}
`, rName, key, path)
}

func testAccFunctionConfig_runtime(rName, runtime string) string {
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

func testAccFunctionConfig_zipNoHandler(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccFunctionConfig_zipNoRuntime(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
}
`, rName))
}

func TestFlattenImageConfigShouldNotFailWithEmptyImageConfig(t *testing.T) {
	t.Parallel()
	response := lambda.ImageConfigResponse{}
	tflambda.FlattenImageConfig(&response)
}

func testAccPreCheckSignerSigningProfile(ctx context.Context, t *testing.T, platformID string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerConn()

	var foundPlatform bool
	err := conn.ListSigningPlatformsPagesWithContext(ctx, &signer.ListSigningPlatformsInput{}, func(page *signer.ListSigningPlatformsOutput, lastPage bool) bool {
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
