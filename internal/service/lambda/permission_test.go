package lambda_test

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestPermissionUnmarshalling(t *testing.T) {
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
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "event_source_token", "test-event-source-token"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_principalOrgID(t *testing.T) {
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionOrgIdConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_org_id", "data.aws_organizations_organization.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromCloudWatch"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "event_source_token", "test-event-source-token"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_statementIDDuplicate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionStatementIdDuplicateConfig(rName),
				ExpectError: regexp.MustCompile(`ResourceConflictException`),
			},
		},
	})
}

func TestAccLambdaPermission_rawFunctionName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withRawFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithRawFuncName"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_statementIDPrefix(t *testing.T) {
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withStatementIdPrefix(rName, "AllowExecutionWithStatementIdPrefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "statement_id", "AllowExecutionWithStatementIdPrefix-"),
					resource.TestCheckResourceAttr(resourceName, "statement_id_prefix", "AllowExecutionWithStatementIdPrefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_qualifier(t *testing.T) {
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withQualifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithQualifier"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(acctest.Provider, tflambda.ResourcePermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLambdaPermission_multiplePerms(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_multiplePerms(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckPermissionExists(resourceNameFirst, &firstStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "arn"),
					// 2nd
					testAccCheckPermissionExists(resourceNameSecond, &firstStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecond, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecond, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecond, "statement_id", "AllowExecutionSecond"),
					resource.TestCheckResourceAttrPair(resourceNameSecond, "function_name", functionResourceName, "arn"),
				),
			},
			{
				Config: testAccPermissionConfig_multiplePermsModified(funcName, roleName),
				Check: resource.ComposeTestCheckFunc(
					// 1st
					testAccCheckPermissionExists(resourceNameFirst, &secondStatement),
					resource.TestCheckResourceAttr(resourceNameFirst, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceNameFirst, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameFirst, "statement_id", "AllowExecutionFirst"),
					resource.TestCheckResourceAttrPair(resourceNameFirst, "function_name", functionResourceName, "arn"),
					// 2nd
					testAccCheckPermissionExists(resourceNameSecondModified, &secondStatementModified),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameSecondModified, "statement_id", "AllowExecutionSec0nd"),
					resource.TestCheckResourceAttrPair(resourceNameSecondModified, "function_name", functionResourceName, "arn"),
					// 3rd
					testAccCheckPermissionExists(resourceNameThird, &thirdStatement),
					resource.TestCheckResourceAttr(resourceNameThird, "action", "lambda:*"),
					resource.TestCheckResourceAttr(resourceNameThird, "principal", "events.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceNameThird, "statement_id", "AllowExecutionThird"),
					resource.TestCheckResourceAttrPair(resourceNameThird, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceNameFirst,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceNameFirst),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceNameSecondModified,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceNameSecondModified),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_s3(t *testing.T) {
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "s3.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromS3"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", bucketResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_sns(t *testing.T) {
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"
	snsTopicResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withSNS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttr(resourceName, "principal", "sns.amazonaws.com"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromSNS"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source_arn", snsTopicResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_iamRole(t *testing.T) {
	var statement tflambda.PolicyStatement

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	iamRoleResourceName := "aws_iam_role.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_withIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunction"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromIAMRole"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_FunctionURLs_iam(t *testing.T) {
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_FunctionURLs_iam(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunctionUrl"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionWithIAM"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "function_url_auth_type", lambda.FunctionUrlAuthTypeAwsIam),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLambdaPermission_FunctionURLs_none(t *testing.T) {
	var statement tflambda.PolicyStatement
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lambda_permission.test"
	functionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, lambda.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_FunctionUrls_None(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &statement),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:InvokeFunctionUrl"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "AllowExecutionFromWithoutAuth"),
					resource.TestCheckResourceAttr(resourceName, "qualifier", ""),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "function_url_auth_type", lambda.FunctionUrlAuthTypeNone),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckPermissionExists(n string, v *tflambda.PolicyStatement) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lambda Permission ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

		output, err := tflambda.FindPolicyStatementByTwoPartKey(conn, rs.Primary.Attributes["function_name"], rs.Primary.ID, rs.Primary.Attributes["qualifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_permission" {
			continue
		}

		_, err := tflambda.FindPolicyStatementByTwoPartKey(conn, rs.Primary.Attributes["function_name"], rs.Primary.ID, rs.Primary.Attributes["qualifier"])

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

func testAccPermissionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
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

func testAccPermissionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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

func testAccPermissionConfig(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  event_source_token = "test-event-source-token"
}
`)
}

func testAccPermissionStatementIdDuplicateConfig(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test1" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}

resource "aws_lambda_permission" "test2" {
  action             = "lambda:InvokeFunction"
  event_source_token = "test-event-source-token"
  function_name      = aws_lambda_function.test.arn
  principal          = "events.amazonaws.com"
  statement_id       = "AllowExecutionFromCloudWatch"
}
`)
}

func testAccPermissionConfig_withRawFunctionName(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionWithRawFuncName"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}
`)
}

func testAccPermissionConfig_withStatementIdPrefix(rName, prefix string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id_prefix = %[1]q
  action              = "lambda:InvokeFunction"
  function_name       = aws_lambda_function.test.arn
  principal           = "events.amazonaws.com"
}
`, prefix))
}

func testAccPermissionConfig_withQualifier(rName string) string {
	// lintignore:AWSAT003,AWSAT005 // ARN, region not actually used
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id   = "AllowExecutionWithQualifier"
  action         = "lambda:InvokeFunction"
  function_name  = aws_lambda_function.test.arn
  principal      = "events.amazonaws.com"
  source_account = "111122223333"
  source_arn     = "arn:aws:events:eu-west-1:111122223333:rule/RunDaily"
  qualifier      = aws_lambda_alias.test.name
}

resource "aws_lambda_alias" "test" {
  name             = %[1]q
  description      = "a sample description"
  function_name    = aws_lambda_function.test.arn
  function_version = "$LATEST"
}
`, rName))
}

var testAccPermissionConfig_multiplePerms_tpl = `
resource "aws_lambda_permission" "first" {
  statement_id  = "AllowExecutionFirst"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}

resource "aws_lambda_permission" "%s" {
  statement_id  = "%s"
  action        = "lambda:*"
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}
%s

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.handler"
  runtime       = "nodejs12.x"
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
  function_name = aws_lambda_function.test.arn
  principal     = "events.amazonaws.com"
}
`, funcName, roleName)
}

func testAccPermissionConfig_withS3(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromS3"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.test.arn
}

resource "aws_s3_bucket" "test" {
  bucket = "%s"
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
}
`, rName))
}

func testAccPermissionConfig_withSNS(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), fmt.Sprintf(`
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
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

func testAccPermissionConfig_withIAMRole(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test" {
  statement_id  = "AllowExecutionFromIAMRole"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = aws_iam_role.test.arn
}
`)
}

func testAccPermissionOrgIdConfig(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
data "aws_organizations_organization" "test" {}

resource "aws_lambda_permission" "test" {
  statement_id       = "AllowExecutionFromCloudWatch"
  action             = "lambda:InvokeFunction"
  function_name      = aws_lambda_function.test.arn
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

func testAccPermissionConfig_FunctionURLs_iam(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test" {
  statement_id           = "AllowExecutionWithIAM"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.test.arn
  principal              = "*"
  function_url_auth_type = "AWS_IAM"
}
`)
}

func testAccPermissionConfig_FunctionUrls_None(rName string) string {
	return acctest.ConfigCompose(testAccPermissionBaseConfig(rName), `
resource "aws_lambda_permission" "test" {
  statement_id           = "AllowExecutionFromWithoutAuth"
  action                 = "lambda:InvokeFunctionUrl"
  function_name          = aws_lambda_function.test.arn
  principal              = "*"
  function_url_auth_type = "NONE"
}
`)
}
