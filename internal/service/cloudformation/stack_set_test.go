// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFormationStackSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test.0"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudformation", regexp.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "AWSCloudFormationStackSetExecutionRole"),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.0.active", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "permission_model", "SELF_MANAGED"),
					resource.TestMatchResourceAttr(resourceName, "stack_set_id", regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName)+"\n"),
					resource.TestCheckNoResourceAttr(resourceName, "template_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSet_administrationRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_administrationRoleARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRole1ResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_administrationRoleARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRole2ResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_description(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_executionRoleName(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_executionRoleName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_executionRoleName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "name2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_managedExecution(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_managedExecution(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.0.active", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSet_name(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackSetConfig_name(""),
				ExpectError: regexp.MustCompile(`expected length`),
			},
			{
				Config:      testAccStackSetConfig_name(sdkacctest.RandStringFromCharSet(129, sdkacctest.CharSetAlpha)),
				ExpectError: regexp.MustCompile(`(cannot be longer|expected length)`),
			},
			{
				Config:      testAccStackSetConfig_name("1"),
				ExpectError: regexp.MustCompile(`must begin with alphabetic character`),
			},
			{
				Config:      testAccStackSetConfig_name("a_b"),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric and hyphen characters`),
			},
			{
				Config: testAccStackSetConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_operationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
					"operation_preferences",
				},
			},
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 3, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "12"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferencesUpdated(rName, 15, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "15"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "75"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 2, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "8"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 0, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferencesUpdated(rName, 0, 95),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "95"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parameters1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_parameters2(rName, "value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter2", "value2"),
				),
			},
			{
				Config: testAccStackSetConfig_parameters1(rName, "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
				),
			},
			{
				Config: testAccStackSetConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_Parameters_default(t *testing.T) {
	acctest.Skip(t, "this resource does not currently ignore unconfigured CloudFormation template parameters with the Default property")
	// Additional references:
	//  * https://github.com/hashicorp/terraform/issues/18863

	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "defaultvalue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_parametersDefault1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1"),
				),
			},
			{
				Config: testAccStackSetConfig_parametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "defaultvalue"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_Parameters_noEcho(t *testing.T) {
	acctest.Skip(t, "this resource does not currently ignore CloudFormation template parameters with the NoEcho property")
	// Additional references:
	//  * https://github.com/hashicorp/terraform-provider-aws/issues/55

	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parametersNoEcho1(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "****"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_parametersNoEcho1(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "****"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_PermissionModel_serviceManaged(t *testing.T) {
	acctest.Skip(t, "API does not support enabling Organizations access (in particular, creating the Stack Sets IAM Service-Linked Role)")

	ctx := acctest.Context(t)
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_permissionModel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudformation", regexp.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "permission_model", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.retain_stacks_on_account_removal", "false"),
					resource.TestMatchResourceAttr(resourceName, "stack_set_id", regexp.MustCompile(fmt.Sprintf("%s:.+", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccStackSetConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateBody(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_templateBody(rName, testAccStackSetTemplateBodyVPC(rName+"1")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName+"1")+"\n"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_templateBody(rName, testAccStackSetTemplateBodyVPC(rName+"2")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName+"2")+"\n"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateURL(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_templateURL1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
			{
				Config: testAccStackSetConfig_templateURL2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
				),
			},
		},
	})
}

func testAccCheckStackSetExists(ctx context.Context, resourceName string, v *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		callAs := rs.Primary.Attributes["call_as"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn(ctx)

		output, err := tfcloudformation.FindStackSetByName(ctx, conn, rs.Primary.ID, callAs)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set" {
				continue
			}

			callAs := rs.Primary.Attributes["call_as"]

			_, err := tfcloudformation.FindStackSetByName(ctx, conn, rs.Primary.ID, callAs)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation StackSet %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackSetNotRecreated(i, j *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackSetId) != aws.StringValue(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) recreated", aws.StringValue(i.StackSetName))
		}

		return nil
	}
}

func testAccCheckStackSetRecreated(i, j *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackSetId) == aws.StringValue(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) not recreated", aws.StringValue(i.StackSetName))
		}

		return nil
	}
}

func testAccPreCheckStackSet(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn(ctx)

	input := &cloudformation.ListStackSetsInput{}
	_, err := conn.ListStackSetsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "ValidationError", "AWS CloudFormation StackSets is not supported") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccStackSetTemplateBodyParameters1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccStackSetTemplateBodyParameters2(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
  Parameter2:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Parameter2Value:
    Value: !Ref Parameter2
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccStackSetTemplateBodyParametersDefault1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
    Default: defaultvalue
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccStackSetTemplateBodyParametersNoEcho1(rName string) string {
	return fmt.Sprintf(`
Parameters:
  Parameter1:
    Type: String
    NoEcho: true
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccStackSetTemplateBodyVPC(rName string) string {
	return fmt.Sprintf(`
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
`, rName)
}

func testAccStackSetConfig_baseAdministrationRoleARNs(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = %[2]d

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-${count.index}"
}
`, rName, count)
}

func testAccStackSetConfig_administrationRoleARN1(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 2), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName)))
}

func testAccStackSetConfig_administrationRoleARN2(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 2), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[1].arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName)))
}

func testAccStackSetConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  description             = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), description))
}

func testAccStackSetConfig_executionRoleName(rName, executionRoleName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  execution_role_name     = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), executionRoleName))
}

func testAccStackSetConfig_managedExecution(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  managed_execution {
    active = true
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName)))
}

func testAccStackSetConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName)))
}

func testAccStackSetConfig_parameters1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParameters1(rName), value1))
}

func testAccStackSetConfig_parameters2(rName, value1, value2 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
    Parameter2 = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParameters2(rName), value1, value2))
}

func testAccStackSetConfig_parametersDefault0(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersDefault1(rName)))
}

func testAccStackSetConfig_parametersDefault1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersDefault1(rName), value1))
}

func testAccStackSetConfig_parametersNoEcho1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersNoEcho1(rName), value1))
}

func testAccStackSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  tags = {
    %[3]q = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), tagKey1, tagValue1))
}

func testAccStackSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccStackSetConfig_templateBody(rName, templateBody string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, templateBody))
}

func testAccStackSetConfig_baseTemplateURL(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"

  depends_on = [
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]
}
`, rName))
}

func testAccStackSetConfig_templateURL1(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseTemplateURL(rName), fmt.Sprintf(`
resource "aws_s3_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template1.yml"

  depends_on = [aws_s3_bucket_acl.test]
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
}
`, rName, testAccStackSetTemplateBodyVPC(rName+"1")))
}

func testAccStackSetConfig_templateURL2(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseTemplateURL(rName), fmt.Sprintf(`
resource "aws_s3_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template2.yml"

  depends_on = [aws_s3_bucket_acl.test]
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
}
`, rName, testAccStackSetTemplateBodyVPC(rName+"2")))
}

func testAccStackSetConfig_permissionModel(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  name             = %[1]q
  permission_model = "SERVICE_MANAGED"

  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = false
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetConfig_operationPreferences(rName string, failureToleranceCount, maxConcurrentCount int) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  operation_preferences {
    failure_tolerance_count = %[2]d
    max_concurrent_count    = %[3]d
  }

  template_body = <<TEMPLATE
%[4]s
TEMPLATE
}
`, rName, failureToleranceCount, maxConcurrentCount, testAccStackSetTemplateBodyVPC(rName)))
}

func testAccStackSetConfig_operationPreferencesUpdated(rName string, failureTolerancePercentage, maxConcurrentPercentage int) string {
	return acctest.ConfigCompose(testAccStackSetConfig_baseAdministrationRoleARNs(rName, 1), fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test[0].arn
  name                    = %[1]q

  operation_preferences {
    failure_tolerance_percentage = %[2]d
    max_concurrent_percentage    = %[3]d
  }

  template_body = <<TEMPLATE
%[4]s
TEMPLATE
}
`, rName, failureTolerancePercentage, maxConcurrentPercentage, testAccStackSetTemplateBodyVPC(rName)))
}
