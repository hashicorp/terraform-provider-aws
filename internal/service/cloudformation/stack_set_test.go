// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationStackSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test.0"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName, names.AttrARN),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudformation", regexache.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "AWSCloudFormationStackSetExecutionRole"),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.0.active", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "permission_model", "SELF_MANAGED"),
					resource.TestMatchResourceAttr(resourceName, "stack_set_id", regexache.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
			// Test import with call_as.
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s,SELF", rName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"template_url",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
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
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_administrationRoleARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRole1ResourceName, names.AttrARN),
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
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRole2ResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_description(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_executionRoleName(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
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
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_managedExecution(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "managed_execution.0.active", acctest.CtTrue),
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
	var stackSet1, stackSet2 awstypes.StackSet
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackSetConfig_name(""),
				ExpectError: regexache.MustCompile(`expected length`),
			},
			{
				Config:      testAccStackSetConfig_name(sdkacctest.RandStringFromCharSet(129, sdkacctest.CharSetAlpha)),
				ExpectError: regexache.MustCompile(`(cannot be longer|expected length)`),
			},
			{
				Config:      testAccStackSetConfig_name(acctest.Ct1),
				ExpectError: regexache.MustCompile(`must begin with alphabetic character`),
			},
			{
				Config:      testAccStackSetConfig_name("a_b"),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric and hyphen characters`),
			},
			{
				Config: testAccStackSetConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_operationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
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
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "12"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferencesUpdated(rName, 15, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "15"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "75"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 2, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "8"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferences(rName, 0, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetConfig_operationPreferencesUpdated(rName, 0, 95),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "95"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parameters1(rName, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", acctest.CtValue1),
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
				Config: testAccStackSetConfig_parameters2(rName, acctest.CtValue1Updated, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter2", acctest.CtValue2),
				),
			},
			{
				Config: testAccStackSetConfig_parameters1(rName, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", acctest.CtValue1Updated),
				),
			},
			{
				Config: testAccStackSetConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
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
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
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
				Config: testAccStackSetConfig_parametersDefault1(rName, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", acctest.CtValue1),
				),
			},
			{
				Config: testAccStackSetConfig_parametersDefault0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
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
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_parametersNoEcho1(rName, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
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
				Config: testAccStackSetConfig_parametersNoEcho1(rName, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "****"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_PermissionModel_serviceManaged(t *testing.T) {
	acctest.Skip(t, "API does not support enabling Organizations access (in particular, creating the Stack Sets IAM Service-Linked Role)")

	ctx := acctest.Context(t)
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_permissionModel(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudformation", regexache.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "permission_model", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.retain_stacks_on_account_removal", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "stack_set_id", regexache.MustCompile(fmt.Sprintf("%s:.+", rName))),
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
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
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
				Config: testAccStackSetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStackSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateBody(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_templateBody(rName, testAccStackSetTemplateBodyVPC(rName+acctest.Ct1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName+acctest.Ct1)+"\n"),
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
				Config: testAccStackSetConfig_templateBody(rName, testAccStackSetTemplateBodyVPC(rName+acctest.Ct2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName+acctest.Ct2)+"\n"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateURL(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet1, stackSet2 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/19015.
func TestAccCloudFormationStackSet_autoDeploymentEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_autoDeployment(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.retain_stacks_on_account_removal", acctest.CtFalse),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/19015.
func TestAccCloudFormationStackSet_autoDeploymentDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var stackSet awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetConfig_autoDeployment(rName, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.retain_stacks_on_account_removal", acctest.CtFalse),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/32536.
// Prerequisites:
// * Organizations management account
// * Organization member account
// * Delegated administrator not configured
// Authenticate with member account as target account and management account as alternate.
func TestAccCloudFormationStackSet_delegatedAdministrator(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var stackSet awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckStackSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers.
				Config: testAccStackSetConfig_delegatedAdministratorInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
				},
				Config: testAccStackSetConfig_delegatedAdministrator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_deployment.0.retain_stacks_on_account_removal", acctest.CtFalse),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}

					return fmt.Sprintf("%s,DELEGATED_ADMIN", rs.Primary.ID), nil
				},
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"call_as",
					"template_url",
				},
			},
		},
	})
}

func testAccCheckStackSetExists(ctx context.Context, resourceName string, v *awstypes.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		output, err := tfcloudformation.FindStackSetByName(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["call_as"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set" {
				continue
			}

			_, err := tfcloudformation.FindStackSetByName(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["call_as"])

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

func testAccCheckStackSetNotRecreated(i, j *awstypes.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StackSetId) != aws.ToString(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) recreated", aws.ToString(i.StackSetName))
		}

		return nil
	}
}

func testAccCheckStackSetRecreated(i, j *awstypes.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StackSetId) == aws.ToString(j.StackSetId) {
			return fmt.Errorf("CloudFormation StackSet (%s) not recreated", aws.ToString(i.StackSetName))
		}

		return nil
	}
}

func testAccPreCheckStackSet(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

	input := &cloudformation.ListStackSetsInput{}
	_, err := conn.ListStackSets(ctx, input)

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
`, rName, testAccStackSetTemplateBodyVPC(rName+acctest.Ct1)))
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
`, rName, testAccStackSetTemplateBodyVPC(rName+acctest.Ct2)))
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

func testAccStackSetConfig_autoDeployment(rName string, enabled, retainStacksOnAccountRemoval bool) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  name             = %[1]q
  permission_model = "SERVICE_MANAGED"

  auto_deployment {
    enabled                          = %[3]t
    retain_stacks_on_account_removal = %[4]t
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), enabled, retainStacksOnAccountRemoval)
}

// Initialize all the providers used by delegated administrator acceptance tests.
var testAccStackSetConfig_delegatedAdministratorInit = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "member" {}

data "aws_caller_identity" "management" {
  provider = awsalternate
}
`)

// Primary provider is Organizations member account that is made a delegated administrator.
// Alternate provider is the Organizations management account.
var testAccStackSetConfigDelegatedAdministratorConfig_base = acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "member" {}

data "aws_caller_identity" "management" {
  provider = awsalternate
}

resource "aws_organizations_delegated_administrator" "test" {
  provider = awsalternate

  account_id        = data.aws_caller_identity.member.account_id
  service_principal = "member.org.stacksets.cloudformation.amazonaws.com"
}
`)

func testAccStackSetConfig_delegatedAdministrator(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfigDelegatedAdministratorConfig_base, fmt.Sprintf(`
resource "aws_cloudformation_stack_set" "test" {
  name             = %[1]q
  permission_model = "SERVICE_MANAGED"
  call_as          = "DELEGATED_ADMIN"

  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = false
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE

  depends_on = [aws_organizations_delegated_administrator.test]

  lifecycle {
    ignore_changes = [administration_role_arn]
  }
}
`, rName, testAccStackSetTemplateBodyVPC(rName)))
}
