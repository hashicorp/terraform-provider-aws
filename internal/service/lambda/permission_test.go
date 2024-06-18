// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestPermissionUnmarshalling(t *testing.T) {
	t.Parallel()

	v := tflambda.Policy{}
	err := json.Unmarshal(testPolicy, &v)
	if err != nil {
		t.Fatalf("Expected no error when unmarshalling: %s", err)
	}
	stmt := v.Statement[0]

	expectedSid := "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
	if stmt.Sid != expectedSid {
		t.Fatalf("Expected Sid to match (%q != %q)", stmt.Sid, expectedSid)
	}

	expectedFunctionName := "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction" // lintignore:AWSAT003,AWSAT005 // unit test
	if stmt.Resource != expectedFunctionName {
		t.Fatalf("Expected function name to match (%q != %q)", stmt.Resource, expectedFunctionName)
	}

	expectedAction := "lambda:InvokeFunction"
	if stmt.Action != expectedAction {
		t.Fatalf("Expected Action to match (%q != %q)", stmt.Action, expectedAction)
	}

	expectedPrincipal := "events.amazonaws.com"
	service := stmt.Principal.(map[string]interface{})["Service"]
	if service != expectedPrincipal {
		t.Fatalf("Expected Principal to match (%q != %q)", service, expectedPrincipal)
	}

	expectedSourceAccount := "319201112229"
	strEquals := stmt.Condition["StringEquals"]
	if strEquals["AWS:SourceAccount"] != expectedSourceAccount {
		t.Fatalf("Expected Source Account to match (%q != %q)", strEquals["AWS:SourceAccount"], expectedSourceAccount)
	}

	expectedEventSourceToken := "test-event-source-token"
	if strEquals["lambda:EventSourceToken"] != expectedEventSourceToken {
		t.Fatalf("Expected Event Source Token to match (%q != %q)", strEquals["lambda:EventSourceToken"], expectedEventSourceToken)
	}
}

func TestPermissionOrgUnmarshalling(t *testing.T) {
	t.Parallel()

	v := tflambda.Policy{}
	err := json.Unmarshal(testOrgPolicy, &v)
	if err != nil {
		t.Fatalf("Expected no error when unmarshalling: %s", err)
	}
	stmt := v.Statement[0]

	expectedSid := "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
	if stmt.Sid != expectedSid {
		t.Fatalf("Expected Sid to match (%q != %q)", stmt.Sid, expectedSid)
	}

	expectedFunctionName := "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction" // lintignore:AWSAT003,AWSAT005 // unit test
	if stmt.Resource != expectedFunctionName {
		t.Fatalf("Expected function name to match (%q != %q)", stmt.Resource, expectedFunctionName)
	}

	expectedAction := "lambda:InvokeFunction"
	if stmt.Action != expectedAction {
		t.Fatalf("Expected Action to match (%q != %q)", stmt.Action, expectedAction)
	}

	expectedPrincipal := "*"
	principal := stmt.Principal.(string)
	if principal != expectedPrincipal {
		t.Fatalf("Expected Principal to match (%q != %q)", principal, expectedPrincipal)
	}

	expectedOrgId := "o-1234567890"
	strEquals := stmt.Condition["StringEquals"]
	if strEquals["aws:PrincipalOrgID"] != expectedOrgId {
		t.Fatalf("Expected Principal Org ID to match (%q != %q)", strEquals["aws:PrincipalOrgID"], expectedOrgId)
	}
}

func TestPermissionGetQualifierFromAliasOrVersionARN_alias(t *testing.T) {
	t.Parallel()

	arnWithAlias := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:testalias" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "testalias"
	qualifier, err := tflambda.GetQualifierFromAliasOrVersionARN(arnWithAlias)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}

func TestPermissionGetQualifierFromAliasOrVersionARN_govcloud(t *testing.T) {
	t.Parallel()

	arnWithAlias := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name:testalias" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "testalias"
	qualifier, err := tflambda.GetQualifierFromAliasOrVersionARN(arnWithAlias)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}

func TestPermissionGetQualifierFromAliasOrVersionARN_version(t *testing.T) {
	t.Parallel()

	arnWithVersion := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:223" // lintignore:AWSAT003,AWSAT005 // unit test
	expectedQualifier := "223"
	qualifier, err := tflambda.GetQualifierFromAliasOrVersionARN(arnWithVersion)
	if err != nil {
		t.Fatalf("Expected no error when getting qualifier: %s", err)
	}
	if qualifier != expectedQualifier {
		t.Fatalf("Expected qualifier to match (%q != %q)", qualifier, expectedQualifier)
	}
}

func TestPermissionGetQualifierFromAliasOrVersionARN_invalid(t *testing.T) {
	t.Parallel()

	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	qualifier, err := tflambda.GetQualifierFromAliasOrVersionARN(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}

	// with trailing colon
	invalidArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:" // lintignore:AWSAT003,AWSAT005 // unit test
	qualifier, err = tflambda.GetQualifierFromAliasOrVersionARN(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when getting qualifier")
	}
	if qualifier != "" {
		t.Fatalf("Expected qualifier to be empty (%q)", qualifier)
	}
}

func TestPermissionGetFunctionNameFromARN_invalid(t *testing.T) {
	t.Parallel()

	invalidArn := "arn:aws:lambda:us-west-2:187636751137:function:" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := tflambda.GetFunctionNameFromARN(invalidArn)
	if err == nil {
		t.Fatalf("Expected error when parsing invalid ARN (%q)", invalidArn)
	}
	if fn != "" {
		t.Fatalf("Expected empty string when parsing invalid ARN (%q)", invalidArn)
	}
}

func TestPermissionGetFunctionNameFromARN_valid(t *testing.T) {
	t.Parallel()

	validArn := "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := tflambda.GetFunctionNameFromARN(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname := "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}

	// With qualifier
	validArn = "arn:aws:lambda:us-west-2:187636751137:function:lambda_function_name:12" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err = tflambda.GetFunctionNameFromARN(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname = "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}
}

func TestPermissionGetFunctionNameFromGovCloudARN(t *testing.T) {
	t.Parallel()

	validArn := "arn:aws-us-gov:lambda:us-gov-west-1:187636751137:function:lambda_function_name" // lintignore:AWSAT003,AWSAT005 // unit test
	fn, err := tflambda.GetFunctionNameFromARN(validArn)
	if err != nil {
		t.Fatalf("Expected no error (%q): %q", validArn, err)
	}
	expectedFunctionname := "lambda_function_name"
	if fn != expectedFunctionname {
		t.Fatalf("Expected Lambda function name to match (%q != %q)",
			validArn, expectedFunctionname)
	}
}

func TestAccLambdaPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "event_source_token", "test-event-source-token"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_principalOrgID(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_orgID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_org_id", "data.aws_organizations_organization.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "event_source_token", "test-event-source-token"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_statementIDDuplicate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionConfig_statementIDDuplicate(rName),
				ExpectError: regexache.MustCompile(`ResourceConflictException`),
			},
		},
	})
}

func TestAccLambdaPermission_rawFunctionName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_rawFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithRawFuncName"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_statementIDPrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_statementIDPrefix(rName, "AllowExecutionWithStatementIdPrefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "events.amazonaws.com"),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "statement_id", "AllowExecutionWithStatementIdPrefix-"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", "AllowExecutionWithStatementIdPrefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_qualifier(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_qualifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithQualifier"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflambda.ResourcePermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaPermission_multiplePerms(t *testing.T) {
	ctx := acctest.Context(t)
	var firstStatement tflambda.PolicyStatement
	var firstStatementModified tflambda.PolicyStatement
	var secondStatement tflambda.PolicyStatement
	var secondStatementModified tflambda.PolicyStatement
	var thirdStatement tflambda.PolicyStatement

	rString := sdkacctest.RandString(8)
	funcName := fmt.Sprintf("tf_acc_lambda_perm_multi_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_lambda_perm_multi_%s", rString)

	resourceNameFirst := "aws_lambda_permission.first"
	resourceNameSecond := "aws_lambda_permission.second"
	resourceNameSecondModified := "aws_lambda_permission.sec0nd"
	resourceNameThird := "aws_lambda_permission.third"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_multiplePerms(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckPermissionExists(ctx, resourceNameFirst, &firstStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "function_name"),
					// 2nd
					testAccCheckPermissionExists(ctx, resourceNameSecond, &firstStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecond, names.AttrAction, "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecond, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecond, "statement_id", "AllowExecutionSecond"),
					resource.TestCheckResourceAttrPair(resourceNameSecond, "function_name", functionResourceName, "function_name"),
				),
			},
			{
				Config: testAccPermissionConfig_multiplePermsModified(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckPermissionExists(ctx, resourceNameFirst, &secondStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "function_name"),
					// 2nd
					testAccCheckPermissionExists(ctx, resourceNameSecondModified, &secondStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecondModified, names.AttrAction, "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "statement_id", "AllowExecutionSec0nd"),
					resource.TestCheckResourceAttrPair(resourceNameSecondModified, "function_name", functionResourceName, "function_name"),
					// 3rd
					testAccCheckPermissionExists(ctx, resourceNameThird, &thirdStatement),
					resource.TestCheckResourceAttr(resourceNameThird, names.AttrAction, "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameThird, names.AttrPrincipal, "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameThird, "statement_id", "AllowExecutionThird"),
					resource.TestCheckResourceAttrPair(resourceNameThird, "function_name", functionResourceName, "function_name"),
				),
			},
			{
				ResourceName:      resourceNameFirst,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceNameFirst),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameSecondModified,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceNameSecondModified),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "s3.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromS3"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", bucketResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_sns(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_sns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "sns.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromSNS"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", snsTopicResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	iamRoleResourceName := "aws_iam_role.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunction"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPrincipal, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromIAMRole"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_FunctionURLs_iam(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_functionURLsIAM(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunctionUrl"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithIAM"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "function_url_auth_type", string(awstypes.FunctionUrlAuthTypeAwsIam)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_FunctionURLs_none(t *testing.T) {
	ctx := acctest.Context(t)
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_functionURLsNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "lambda:InvokeFunctionUrl"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromWithoutAuth"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "function_name"),
					resource.TestCheckResourceAttr(resourceName, "function_url_auth_type", string(awstypes.FunctionUrlAuthTypeNone)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIDFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPermissionExists(ctx context.Context, n string, v *tflambda.PolicyStatement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lambda Permission ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		output, err := tflambda.FindPolicyStatementByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.ID, rs.Primary.Attributes["qualifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_permission" {
				continue
			}

			_, err := tflambda.FindPolicyStatementByTwoPartKey(ctx, conn, rs.Primary.Attributes["function_name"], rs.Primary.ID, rs.Primary.Attributes["qualifier"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Permission (%s/%s) still exists", rs.Primary.Attributes["function_name"], rs.Primary.ID)
		}

		return nil
	}
}

func testAccPermissionImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		if v, ok := rs.Primary.Attributes["qualifier"]; ok && v != "" {
			return fmt.Sprintf("%s:%s/%s", rs.Primary.Attributes["function_name"], v, rs.Primary.Attributes["statement_id"]), nil
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["function_name"], rs.Primary.Attributes["statement_id"]), nil
	}
}

func testAccPermissionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs16.x"
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
`, rName)
}

func testAccPermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.function_name
  principal          = "events.amazonaws.com"
  event_source_token = "test-event-source-token"
}
`)
}

func testAccPermissionConfig_statementIDDuplicate(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test1" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.function_name
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_permission" "test2" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.function_name
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}
`)
}

func testAccPermissionConfig_rawFunctionName(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionWithRawFuncName"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "events.amazonaws.com"
}
`)
}

func testAccPermissionConfig_statementIDPrefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id_prefix = %[1]q
  action              = "lambda:InvokeFunction"
  function_name       = aws_lambda_function.test.function_name
  principal           = "events.amazonaws.com"
}
`, prefix))
}

func testAccPermissionConfig_qualifier(rName string) string {
	// lintignore:AWSAT003,AWSAT005 // ARN, region not actually used
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id   = "AllowExecutionWithQualifier"
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.test.function_name
  principal      = "events.amazonaws.com"
  source_account = "111122223333"
  source_arn     = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"
  qualifier      = aws_lambda_alias.test.name
}

resource "aws_lambda_alias" "test" {
  name             = %[1]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.function_name
  function_version = "$LATEST"
}
`, rName))
}

var testAccPermissionConfig_multiplePerms_tpl = `
resource "aws_lambda_permission" "first" {
  statement_id  = "AllowExecutionFirst"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "events.amazonaws.com"
}

resource "aws_lambda_permission" "%s" {
  statement_id  = "%s"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.function_name
  principal     = "events.amazonaws.com"
}
%s

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs16.x"
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
`

func testAccPermissionConfig_multiplePerms(funcName, roleName string) string {
	return fmt.Sprintf(testAccPermissionConfig_multiplePerms_tpl,
		"second", "AllowExecutionSecond", "", funcName, roleName)
}

func testAccPermissionConfig_multiplePermsModified(funcName, roleName string) string {
	return fmt.Sprintf(testAccPermissionConfig_multiplePerms_tpl,
		"sec0nd", "AllowExecutionSec0nd", `
resource "aws_lambda_permission" "third" {
  statement_id  = "AllowExecutionThird"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.function_name
  principal     = "events.amazonaws.com"
}
`, funcName, roleName)
}

func testAccPermissionConfig_s3(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName))
}

func testAccPermissionConfig_sns(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.test.arn
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.test.arn
}
`, rName))
}

func testAccPermissionConfig_iamRole(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromIAMRole"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = aws_iam_role.test.arn
}
`)
}

func testAccPermissionConfig_orgID(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
data "aws_organizations_organization" "test" {}

resource "aws_lambda_permission" "test" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.function_name
  principal          = "*"
  principal_org_id   = data.aws_organizations_organization.test.id
  event_source_token = "test-event-source-token"
}
`)
}

// lintignore:AWSAT003,AWSAT005 // unit test
var testPolicy = []byte(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Condition": {
        "StringEquals": {
          "AWS:SourceAccount": "319201112229",
          "lambda:EventSourceToken": "test-event-source-token"
        },
        "ArnLike": {
          "AWS:SourceArn": "arn:aws:events:eu-west-1:319201112229:rule/RunDaily"
        }
      },
      "Action": "lambda:InvokeFunction",
      "Resource": "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Sid": "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
    }
  ],
  "Id": "default"
}`)

// lintignore:AWSAT003,AWSAT005 // unit test
var testOrgPolicy = []byte(`{
	"Version": "2012-10-17",
	"Statement": [
	  {
		"Condition": {
		  "StringEquals": {
			"aws:PrincipalOrgID": "o-1234567890"
		  }
		},
		"Action": "lambda:InvokeFunction",
		"Resource": "arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction",
		"Effect": "Allow",
		"Principal": "*",
		"Sid": "36fe77d9-a4ae-13fb-8beb-5dc6821d5291"
	  }
	],
	"Id": "default"
  }`)

func testAccPermissionConfig_functionURLsIAM(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test" {
  statement_id           = "AllowExecutionWithIAM"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.test.function_name
  principal              = "*"
  function_url_auth_type = "AWS_IAM"
}
`)
}

func testAccPermissionConfig_functionURLsNone(rName string) string {
	return acctest.ConfigCompose(testAccPermissionConfig_base(rName), `
resource "aws_lambda_permission" "test" {
  statement_id           = "AllowExecutionFromWithoutAuth"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.test.function_name
  principal              = "*"
  function_url_auth_type = "NONE"
}
`)
}
