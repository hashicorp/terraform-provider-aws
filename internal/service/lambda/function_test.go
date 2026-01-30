// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	signertypes "github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.LambdaServiceID, testAccErrorCheckSkip)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(funcName, policyName, roleName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionResponseStreamingInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					testAccCheckFunctionName(&conf, funcName),
					resource.TestCheckResourceAttr(resourceName, "architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureX8664)),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					resource.TestCheckResourceAttrSet(resourceName, "code_sha256"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ephemeral_storage.0.size", "512"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "Text"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_group", fmt.Sprintf("/aws/lambda/%s", funcName)),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, "reserved_concurrent_executions", "-1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, tflambda.FunctionVersionLatest),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function),
					acctest.CheckSDKResourceDisappears(ctx, t, tflambda.ResourceFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_filename(initialFilename, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, tflambda.FunctionVersionLatest),
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
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, tflambda.FunctionVersionLatest),
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
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSignerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_cscCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "code_signing_config_arn", cscResourceName, names.AttrARN),
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
					resource.TestCheckResourceAttrPair(resourceName, "code_signing_config_arn", cscUpdateResourceName, names.AttrARN),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
			{
				Config: testAccFunctionConfig_concurrencyPublished(rName),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_noFilenameAndS3Attributes(rName),
				ExpectError: regexache.MustCompile("one of `filename,image_uri,s3_bucket` must be specified"),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_encryptedEnvVariablesKey1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKey1ResourceName, names.AttrARN),
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
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, kmsKey2ResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccFunctionConfig_envVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "environment.0.variables.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_basic(badFuncName, rName, rName, rName),
				ExpectError: regexache.MustCompile(`invalid value for function_name \(must be valid function name or function ARN\)`),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable("test-fixtures/lambdatest.zip", rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable("test-fixtures/lambdatest.zip", rName, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.RuntimeNodejs20x)),
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
				Config: testAccFunctionConfig_versionedNodeJs22xRuntime(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
					resource.TestCheckResourceAttr(resourceName, "runtime", string(awstypes.RuntimeNodejs22x)),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable(fileName, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtFalse),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, unpublishedVersion),
				),
			},
			{
				// No changes, except to `publish`. This should publish a new version.
				Config: testAccFunctionConfig_publishable(fileName, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_publishable(fileName, rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				// No changes, except to `publish`. This should not update the current version.
				Config: testAccFunctionConfig_publishable(fileName, rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf2),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtFalse),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, "qualified_arn", "lambda", "function:{function_name}:{version}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_deadLetter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.0", names.AttrARN),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_deadLetter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.0", names.AttrARN),
				),
			},
			{
				Config: testAccFunctionConfig_deadLetterUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "dead_letter_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "dead_letter_config.0.target_arn", "aws_sns_topic.test.1", names.AttrARN),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_nilDeadLetter(rName),
				ExpectError: regexache.MustCompile(
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttrPair(resourceName, "file_system_config.0.arn", "aws_efs_access_point.test1", names.AttrARN),
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
					resource.TestCheckResourceAttrPair(resourceName, "file_system_config.0.arn", "aws_efs_access_point.test2", names.AttrARN),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeImage)),
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

func TestAccLambdaFunction_imageConfigNull(t *testing.T) {
	ctx := acctest.Context(t)
	key := "AWS_LAMBDA_IMAGE_LATEST_ID"
	imageLatestID := os.Getenv(key)
	if imageLatestID == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			// Ensure a function with lambda image configuration can be created
			{
				Config: testAccFunctionConfig_imageConfigNull(rName, imageLatestID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeImage)),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageLatestID),
				),
			},
			// Ensure configuration can be imported
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_image(rName, imageLatestID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeImage)),
					resource.TestCheckResourceAttr(resourceName, "image_uri", imageLatestID),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.entry_point.0", "/bootstrap-with-handler"),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.command.0", "app.lambda_handler"),
					resource.TestCheckResourceAttr(resourceName, "image_config.0.working_directory", "/var/task"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureArm64)),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureArm64)),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureArm64)),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureX8664)),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureArm64)),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
					resource.TestCheckResourceAttr(resourceName, "architectures.0", string(awstypes.ArchitectureX8664)),
					resource.TestCheckResourceAttr(resourceName, "layers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "package_type", string(awstypes.PackageTypeZip)),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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

func TestAccLambdaFunction_loggingConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_loggingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "Text"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_group", fmt.Sprintf("/aws/lambda/%s_custom", rName)),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_updateLoggingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", "TRACE"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", "DEBUG"),
				),
			},
			{
				Config: testAccFunctionConfig_updateLoggingConfigLevelsUnspecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", "TRACE"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", "DEBUG"),
				),
			},
			{
				Config: testAccFunctionConfig_loggingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "Text"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", ""),
				),
			},
		},
	})
}

func TestAccLambdaFunction_loggingConfigWithPublish(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "Text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "Text"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "JSON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", "INFO"),
				),
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "JSON"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublishUpdated1(rName, "JSON", "DEBUG"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "3"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", "INFO"),
				),
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublishUpdated2(rName, "JSON", "WARN"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "4"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", "WARN"),
				),
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "JSON"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "Text"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "publish", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.application_log_level", ""),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.log_format", "Text"),
					resource.TestCheckResourceAttr(resourceName, "logging_config.0.system_log_level", ""),
				),
			},
			{
				Config: testAccFunctionConfig_loggingConfigWithPublish(rName, "Text"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_kmsKeyARNNoEnvironmentVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function1),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),

		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcPublish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "2"),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", vpcResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccLambdaFunction_VPC_replaceSGWithDefault(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcReplaceSGWithDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "replace_security_groups_on_destroy", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLambdaFunction_VPC_replaceSGWithCustom(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var function lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"
	vpcResourceName := "aws_vpc.test"
	replacementSGName := "aws_security_group.test_replacement"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_vpcReplaceSGWithCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &function),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "replace_security_groups_on_destroy", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "replacement_security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "replacement_security_group_ids.0", replacementSGName, names.AttrID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
				ImportStateVerifyIgnore: []string{"publish", names.AttrS3Bucket, "s3_key"},
			},
		},
	})
}

func TestAccLambdaFunction_LocalUpdate_sourceCodeHash(t *testing.T) {
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "source_code_hash"},
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
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccLambdaFunction_LocalUpdate_codeSha256(t *testing.T) {
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_local_codeSHA256(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "source_code_hash"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFunctionConfig_local_codeSHA256(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccLambdaFunction_LocalUpdate_sourceCodeHashToCodeSha256(t *testing.T) {
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "source_code_hash"},
			},
			// Switch from source_code_hash to code_sha256 for tracking source code changes
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.js": "lambda.js"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFunctionConfig_local_codeSHA256(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
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
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
				),
			},
		},
	})
}

func TestAccLambdaFunction_LocalUpdate_publish(t *testing.T) {
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func.py": "lambda_handler.py"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
				},
				Config: testAccFunctionConfig_localPublish(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "dLPb9UCUTa8WVNATdCYpZIcIxLWEoR4TLDWvr9rajBw="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "source_code_hash"},
			},
			{
				PreConfig: func() {
					if err := testAccCreateZipFromFiles(map[string]string{"test-fixtures/lambda_func_modified.py": "lambda_handler.py"}, zipFile); err != nil {
						t.Fatalf("error creating zip from files: %s", err)
					}
					timeBeforeUpdate = time.Now()
				},
				Config: testAccFunctionConfig_localPublish(path, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckSourceCodeHash(&conf, "7x43uxhWHTejc6xUvJlAcRvdVmRpqwGIYHpok5qDiYs="),
					func(s *terraform.State) error {
						return testAccCheckAttributeIsDateAfter(s, resourceName, "last_modified", timeBeforeUpdate)
					},
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", names.AttrS3Bucket, "s3_key", "s3_object_version"},
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
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
					testAccCheckSourceCodeHash(&conf, "MbW0T1Pcy1QPtrFC9dT7hUfircj1NXss2uXgakqzAbk="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", names.AttrS3Bucket, "s3_key"},
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
					testAccCheckSourceCodeHash(&conf, "7qn3LZOWCpWK5nm49qjw+VrbPQHfdu2ZrDjBsSUveKM="),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
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
			ExpectError: regexache.MustCompile(`expected runtime to be one of`),
		},
	}
	for _, runtime := range awstypes.Runtime("").Values() {
		// EOL runtimes.
		// https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtime-support-policy.
		switch runtime {
		case awstypes.RuntimeDotnetcore21:
			fallthrough
		case awstypes.RuntimePython27:
			fallthrough
		case awstypes.RuntimePython36:
			fallthrough
		case awstypes.RuntimeRuby25:
			fallthrough
		case awstypes.RuntimeNodejs14x:
			fallthrough
		case awstypes.RuntimeNodejs12x:
			fallthrough
		case awstypes.RuntimeNodejs10x:
			fallthrough
		case awstypes.RuntimeNodejs810:
			fallthrough
		case awstypes.RuntimeNodejs610:
			fallthrough
		case awstypes.RuntimeNodejs43edge:
			fallthrough
		case awstypes.RuntimeNodejs43:
			fallthrough
		case awstypes.RuntimeNodejs:
			fallthrough
		case awstypes.RuntimeDotnetcore31:
			fallthrough
		case awstypes.RuntimeDotnetcore20:
			fallthrough
		case awstypes.RuntimeDotnetcore10:
			continue
		}

		steps = append(steps, resource.TestStep{
			Config: testAccFunctionConfig_runtime(rName, string(runtime)),
			Check: resource.ComposeTestCheckFunc(
				testAccCheckFunctionExists(ctx, resourceName, &v),
				resource.TestCheckResourceAttr(resourceName, "runtime", string(runtime)),
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps:                    steps,
	})
}

func TestAccLambdaFunction_Zip_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFunctionConfig_zipNoHandler(rName),
				ExpectError: regexache.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
			{
				Config:      testAccFunctionConfig_zipNoRuntime(rName),
				ExpectError: regexache.MustCompile("handler and runtime must be set when PackageType is Zip"),
			},
		},
	})
}

func TestAccLambdaFunction_ipv6AllowedForDualStack(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_ipv6AllowedForDualStackDisabled(rName),
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
				Config: testAccFunctionConfig_ipv6AllowedForDualStackEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.ipv6_allowed_for_dual_stack", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLambdaFunction_sourceKMSKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_sourceKMSKeyARN(rName, "test"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					testAccCheckFunctionName(&conf, rName),
					resource.TestCheckResourceAttrPair(resourceName, "source_kms_key_arn", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_sourceKMSKeyARN(rName, "test2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					testAccCheckFunctionInvokeARN(resourceName, &conf),
					testAccCheckFunctionQualifiedInvokeARN(resourceName, &conf),
					testAccCheckFunctionName(&conf, rName),
					resource.TestCheckResourceAttrPair(resourceName, "source_kms_key_arn", "aws_kms_key.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccLambdaFunction_tenancyConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_tenancyConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tenancy_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenancy_config.0.tenant_isolation_mode", "PER_TENANT"),
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

func TestAccLambdaFunction_tenancyConfigForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tenancy_config.#", "0"),
				),
			},
			{
				Config: testAccFunctionConfig_tenancyConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tenancy_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenancy_config.0.tenant_isolation_mode", "PER_TENANT"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_durableConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast2RegionID) // Durable Functions is only available in us-east-2
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_durableConfig(rName, "", 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "durable_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.execution_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.retention_period", "7"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish"},
			},
			{
				Config: testAccFunctionConfig_durableConfig(rName, "Updated description", 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "durable_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.execution_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.retention_period", "7"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
				),
			},
			{
				Config: testAccFunctionConfig_durableConfig(rName, "Updated description", 600, 14),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "durable_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.execution_timeout", "600"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.retention_period", "14"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_durableConfigForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast2RegionID) // Durable Functions is only available in us-east-2
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName, rName, rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "durable_config.#", "0"),
				),
			},
			{
				Config: testAccFunctionConfig_durableConfig(rName, "", 300, 7),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "durable_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.execution_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "durable_config.0.retention_period", "7"),
				),
			},
		},
	})
}

func TestAccLambdaFunction_resetNonRefreshableAttributesAfterUpdateFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	resourceName := "aws_lambda_function.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_resetNonRefreshableAttributesAfterUpdateFailure(rName, "lambdatest.zip", "lambdatest.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "s3_key", "lambdatest.zip"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				// Update with a non-existent S3 key to force an error
				Config:      testAccFunctionConfig_resetNonRefreshableAttributesAfterUpdateFailure(rName, "lambdatest.zip", "lambdatest_not_exist.zip"),
				ExpectError: regexache.MustCompile(`The specified key does not exist`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Revert to previous configuration to ensure non-refreshable attributes were reset
				// This step would fail if s3_key was not reset to "lambdatest.zip"
				Config: testAccFunctionConfig_resetNonRefreshableAttributesAfterUpdateFailure(rName, "lambdatest.zip", "lambdatest.zip"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("s3_key"), knownvalue.StringExact("lambdatest.zip")),
					},
				},
			},
			{
				Config: testAccFunctionConfig_resetNonRefreshableAttributesAfterUpdateFailure(rName, "lambdatest_modified.zip", "lambdatest_modified.zip"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "s3_key", "lambdatest_modified.zip"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccLambdaFunction_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionNoDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLambdaFunction_capacityProvider(t *testing.T) {
	ctx := acctest.Context(t)
	var conf lambda.GetFunctionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_capacityProvider(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, resourceName, &conf),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lambda", "function:{function_name}"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filename", "publish", "publish_to"},
			},
		},
	})
}

func testAccCheckFunctionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function" {
				continue
			}

			_, err := tflambda.FindFunctionByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckFunctionNoDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_function" {
				continue
			}

			_, err := tflambda.FindFunctionByName(ctx, conn, rs.Primary.ID)

			return err
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		output, err := tflambda.FindFunctionByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckSignerSigningProfile(ctx context.Context, t *testing.T, platformID string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SignerClient(ctx)

	input := &signer.ListSigningPlatformsInput{}

	pages := signer.NewListSigningPlatformsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if acctest.PreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		if page == nil {
			t.Skip("skipping acceptance testing: empty response")
		}

		for _, platform := range page.Platforms {
			if platform == (signertypes.SigningPlatform{}) {
				continue
			}

			if aws.ToString(platform.PlatformId) == platformID {
				return
			}
		}
	}

	t.Skipf("skipping acceptance testing: Signing Platform (%s) not found", platformID)
}

func testAccCheckFunctionQualifiedInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		qualifiedArn := fmt.Sprintf("%s:%s", aws.ToString(function.Configuration.FunctionArn), aws.ToString(function.Configuration.Version))
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "qualified_invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", qualifiedArn))(s)
	}
}

func testAccCheckFunctionInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.ToString(function.Configuration.FunctionArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2015-03-31/functions/%s/invocations", arn))(s)
	}
}

func testAccCheckFunctionResponseStreamingInvokeARN(name string, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		arn := aws.ToString(function.Configuration.FunctionArn)
		return acctest.CheckResourceAttrRegionalARNAccountID(name, "response_streaming_invoke_arn", "apigateway", "lambda", fmt.Sprintf("path/2021-11-15/functions/%s/response-streaming-invocations", arn))(s)
	}
}

func testAccInvokeFunction(ctx context.Context, function *lambda.GetFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f := function.Configuration
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		// If the function is VPC-enabled this will create ENI automatically
		_, err := conn.Invoke(ctx, &lambda.InvokeInput{
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

func testAccFunctionConfigBase_properIAMDependencies(rName string) string {
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
`, rName)
}

func testAccFunctionConfig_basic(funcName, policyName, roleName, sgName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(policyName, roleName, sgName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}
`, funcName))
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
  runtime       = "nodejs20.x"
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
  runtime                 = "nodejs20.x"
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
  runtime                 = "nodejs20.x"
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
  runtime       = "nodejs20.x"
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
  runtime                        = "nodejs20.x"
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
  runtime                        = "nodejs20.x"
  reserved_concurrent_executions = 222
}
`, rName))
}

func testAccFunctionConfig_concurrencyPublished(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename                       = "test-fixtures/lambdatest.zip"
  function_name                  = %[1]q
  role                           = aws_iam_role.iam_for_lambda.arn
  handler                        = "exports.example"
  publish                        = true
  runtime                        = "nodejs20.x"
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
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"

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
  description             = "%[1]s-1"
  deletion_window_in_days = 7
  enable_key_rotation     = true

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
  description             = "%[1]s-2"
  deletion_window_in_days = 7
  enable_key_rotation     = true

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
  runtime       = "nodejs20.x"

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
  description             = "%[1]s-2"
  deletion_window_in_days = 7
  enable_key_rotation     = true

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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"
}
`, fileName, rName, publish))
}

func testAccFunctionConfig_versionedNodeJs22xRuntime(fileName, rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  publish       = true
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs22.x"
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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

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

resource "aws_efs_mount_target" "test" {
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
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.test1.arn
    local_mount_path = "/mnt/efs"
  }

  depends_on = [aws_efs_mount_target.test]
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

resource "aws_efs_mount_target" "test" {
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
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda_az2.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.test2.arn
    local_mount_path = "/mnt/lambda"
  }

  depends_on = [aws_efs_mount_target.test]
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

func testAccFunctionConfig_imageConfigNull(rName, imageID string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  image_uri     = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  package_type  = "Image"
  image_config {
    entry_point       = null
    command           = null
    working_directory = null
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
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"
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
  compatible_runtimes      = ["nodejs20.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
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
  compatible_runtimes      = ["nodejs20.x"]
  compatible_architectures = ["arm64", "x86_64"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

  ephemeral_storage {
    size = 2048
  }
}
`, rName))
}

func testAccFunctionConfig_loggingConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  logging_config {
    log_format = "Text"
    log_group  = %[2]q
  }
}
`, rName, fmt.Sprintf("/aws/lambda/%s_custom", rName)))
}

func testAccFunctionConfig_updateLoggingConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  logging_config {
    application_log_level = "TRACE"
    log_format            = "JSON"
    system_log_level      = "DEBUG"
  }
}
`, rName))
}

func testAccFunctionConfig_updateLoggingConfigLevelsUnspecified(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  logging_config {
    log_format = "JSON"
  }
}
`, rName))
}

func testAccFunctionConfig_loggingConfigWithPublish(rName, logFormat string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  publish       = true

  logging_config {
    log_format = %[2]q
  }
}
`, rName, logFormat))
}

func testAccFunctionConfig_loggingConfigWithPublishUpdated1(rName, logFormat, appLogLevel string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  publish       = true

  logging_config {
    log_format            = %[2]q
    application_log_level = %[3]q
  }
}
`, rName, logFormat, appLogLevel))
}

func testAccFunctionConfig_loggingConfigWithPublishUpdated2(rName, logFormat, sysLogLevel string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  publish       = true

  logging_config {
    log_format       = %[2]q
    system_log_level = %[3]q
  }
}
`, rName, logFormat, sysLogLevel))
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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

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
  enable_key_rotation     = true

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
  runtime       = "nodejs20.x"
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
  compatible_runtimes = ["nodejs20.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
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
  compatible_runtimes = ["nodejs20.x"]
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"

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
  runtime       = "nodejs20.x"
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
  runtime       = "nodejs20.x"
  publish       = true
  vpc_config {
    security_group_ids = []
    subnet_ids         = []
  }
}
`, rName))
}

func testAccFunctionConfig_vpcProperIAMDependencies(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_properIAMDependencies(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName))
}

func testAccFunctionConfig_vpcReplaceSGWithDefault(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_properIAMDependencies(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  replace_security_groups_on_destroy = true

  vpc_config {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName))
}

func testAccFunctionConfig_vpcReplaceSGWithCustom(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_properIAMDependencies(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test_replacement" {
  depends_on = [aws_iam_role_policy_attachment.test]

  name   = "%[1]s-replacement"
  vpc_id = aws_vpc.test.id
}

resource "aws_lambda_function" "test" {
  depends_on = [aws_iam_role_policy_attachment.test]

  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  replace_security_groups_on_destroy = true
  replacement_security_group_ids     = [aws_security_group.test_replacement.id]

  vpc_config {
    subnet_ids         = [aws_subnet.test.id]
    security_group_ids = [aws_security_group.test.id]
  }
}
`, rName))
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
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = []
    security_group_ids = []
  }
}
`, rName))
}

func testAccFunctionConfigBase_iamRole(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccFunctionConfig_s3Simple(rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "lambda_bucket" {
  bucket = %[1]q
}

resource "aws_s3_object" "lambda_code" {
  bucket = aws_s3_bucket.lambda_bucket.bucket
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_object.lambda_code.bucket
  s3_key        = aws_s3_object.lambda_code.key
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}
`, rName))
}

func testAccFunctionConfig_local(filePath, rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = %[1]q
  source_code_hash = filebase64sha256(%[1]q)
  function_name    = %[2]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs20.x"
}
`, filePath, rName))
}

func testAccFunctionConfig_localNameOnly(filePath, rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}
`, filePath, rName))
}

func testAccFunctionConfig_localPublish(filePath, rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename         = %[1]q
  source_code_hash = filebase64sha256(%[1]q)
  function_name    = %[2]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "lambda_handler.lambda_handler"
  runtime          = "python3.13"
  publish          = true

  snap_start {
    apply_on = "PublishedVersions"
  }
}
`, filePath, rName))
}

func testAccFunctionConfig_local_codeSHA256(filePath, rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = %[1]q
  code_sha256   = filebase64sha256(%[1]q)
  function_name = %[2]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}
`, filePath, rName))
}

func testAccFunctionConfig_s3(key, path, rName string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = %[3]q
  force_destroy = true
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

resource "aws_lambda_function" "test" {
  s3_bucket         = aws_s3_object.o.bucket
  s3_key            = aws_s3_object.o.key
  s3_object_version = aws_s3_object.o.version_id
  function_name     = %[3]q
  role              = aws_iam_role.iam_for_lambda.arn
  handler           = "exports.example"
  runtime           = "nodejs20.x"
}
`, key, path, rName))
}

func testAccFunctionConfig_s3UnversionedTPL(rName, key, path string) string {
	return acctest.ConfigCompose(
		testAccFunctionConfigBase_iamRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "artifacts" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "o" {
  bucket = aws_s3_bucket.artifacts.bucket
  key    = %[2]q
  source = %[3]q
  etag   = filemd5(%[3]q)
}

resource "aws_lambda_function" "test" {
  s3_bucket     = aws_s3_object.o.bucket
  s3_key        = aws_s3_object.o.key
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
}
`, rName, key, path))
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
  runtime       = "nodejs20.x"
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

func testAccFunctionConfig_ipv6AllowedForDualStackDisabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  vpc_config {
    subnet_ids         = [aws_subnet.subnet_for_lambda.id]
    security_group_ids = [aws_security_group.sg_for_lambda.id]
  }
}
`, rName))
}

func testAccFunctionConfig_ipv6AllowedForDualStackEnabled(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  vpc_config {
    ipv6_allowed_for_dual_stack = true
    subnet_ids                  = [aws_subnet.subnet_for_lambda.id]
    security_group_ids          = [aws_security_group.sg_for_lambda.id]
  }
}
`, rName))
}

func testAccFunctionConfig_sourceKMSKeyARN(rName, kmsIdentifier string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "%[1]s-1"
  deletion_window_in_days = 7
  enable_key_rotation     = true

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
  description             = "%[1]s-2"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-2",
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
  filename           = "test-fixtures/lambdatest.zip"
  function_name      = %[1]q
  role               = aws_iam_role.iam_for_lambda.arn
  handler            = "exports.example"
  runtime            = "nodejs20.x"
  source_kms_key_arn = aws_kms_key.%[2]s.arn
}
`, rName, kmsIdentifier))
}

func testAccFunctionConfig_tenancyConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"

  tenancy_config {
    tenant_isolation_mode = "PER_TENANT"
  }
}
`, rName))
}

func testAccFunctionConfig_skipDestroy(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs20.x"
  skip_destroy  = true
}
`, rName))
}

func testAccFunctionConfig_capacityProvider(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_basic(rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/capacityprovider.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "index.handler"
  runtime       = "python3.14"
  memory_size   = 32768

  publish    = true
  publish_to = "LATEST_PUBLISHED"

  capacity_provider_config {
    lambda_managed_instances_capacity_provider_config {
      capacity_provider_arn = aws_lambda_capacity_provider.test.arn
    }
  }
}
`, rName))
}

func testAccFunctionConfig_resetNonRefreshableAttributesAfterUpdateFailure(rName, zipFileS3, zipFileLambda string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket             = aws_s3_bucket.test.bucket
  key                = %[2]q
  source             = "test-fixtures/%[2]s"
  checksum_algorithm = "SHA256"
}

resource "aws_lambda_function" "test" {
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "exports.example"
  runtime          = "nodejs20.x"
  s3_bucket        = aws_s3_bucket.test.bucket
  s3_key           = %[3]q
  source_code_hash = aws_s3_object.test.checksum_sha256
}
`, rName, zipFileS3, zipFileLambda))
}

func testAccFunctionConfig_durableConfig(rName, description string, executionTimeout, retentionPeriod int) string {
	descriptionLine := ""
	if description != "" {
		descriptionLine = fmt.Sprintf("  description   = %q", description)
	}

	return acctest.ConfigCompose(
		acctest.ConfigLambdaBase(rName, rName, rName),
		fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs22.x"
%[2]s
  durable_config {
    execution_timeout = %[3]d
    retention_period  = %[4]d
  }

  timeouts {
    delete = "60m"
  }
}
`, rName, descriptionLine, executionTimeout, retentionPeriod))
}
