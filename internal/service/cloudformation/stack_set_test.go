package cloudformation_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCloudFormationStackSet_basic(t *testing.T) {
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "cloudformation", regexp.MustCompile(`stackset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "AWSCloudFormationStackSetExecutionRole"),
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
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudformation.ResourceStackSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSet_administrationRoleARN(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	iamRoleResourceName1 := "aws_iam_role.test1"
	iamRoleResourceName2 := "aws_iam_role.test2"
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetAdministrationRoleARN1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName1, "arn"),
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
				Config: testAccStackSetAdministrationRoleARN2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrPair(resourceName, "administration_role_arn", iamRoleResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_description(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_executionRoleName(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetExecutionRoleNameConfig(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetExecutionRoleNameConfig(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "execution_role_name", "name2"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_name(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackSetNameConfig(""),
				ExpectError: regexp.MustCompile(`expected length`),
			},
			{
				Config:      testAccStackSetNameConfig(sdkacctest.RandStringFromCharSet(129, sdkacctest.CharSetAlpha)),
				ExpectError: regexp.MustCompile(`(cannot be longer|expected length)`),
			},
			{
				Config:      testAccStackSetNameConfig("1"),
				ExpectError: regexp.MustCompile(`must begin with alphabetic character`),
			},
			{
				Config:      testAccStackSetNameConfig("a_b"),
				ExpectError: regexp.MustCompile(`must contain only alphanumeric and hyphen characters`),
			},
			{
				Config: testAccStackSetNameConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetNameConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_operationPreferences(t *testing.T) {
	var stackSet cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetOperationPreferencesConfig(rName, 1, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
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
				Config: testAccStackSetOperationPreferencesConfig(rName, 3, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "12"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetOperationPreferencesUpdatedConfig(rName, 15, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "15"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "75"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetOperationPreferencesConfig(rName, 2, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "8"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetOperationPreferencesConfig(rName, 0, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
			{
				Config: testAccStackSetOperationPreferencesUpdatedConfig(rName, 0, 95),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet),
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
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetParameters1Config(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetParameters2Config(rName, "value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter2", "value2"),
				),
			},
			{
				Config: testAccStackSetParameters1Config(rName, "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1updated"),
				),
			},
			{
				Config: testAccStackSetNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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

	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetParametersDefault0Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetParametersDefault1Config(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.Parameter1", "value1"),
				),
			},
			{
				Config: testAccStackSetParametersDefault0Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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

	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetParametersNoEcho1Config(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetParametersNoEcho1Config(rName, "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
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

	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckStackSet(t)
			acctest.PreCheckOrganizationsAccount(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID, "organizations"),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetPermissionModelConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetTags1Config(rName, "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
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
				Config: testAccStackSetTags2Config(rName, "value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
				),
			},
			{
				Config: testAccStackSetTags1Config(rName, "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1updated"),
				),
			},
			{
				Config: testAccStackSetNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateBody(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetTemplateBodyConfig(rName, testAccStackSetTemplateBodyVPC(rName+"1")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetTemplateBodyConfig(rName, testAccStackSetTemplateBodyVPC(rName+"2")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttr(resourceName, "template_body", testAccStackSetTemplateBodyVPC(rName+"2")+"\n"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSet_templateURL(t *testing.T) {
	var stackSet1, stackSet2 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStackSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetTemplateURL1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet1),
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
				Config: testAccStackSetTemplateURL2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(resourceName, &stackSet2),
					testAccCheckStackSetNotRecreated(&stackSet1, &stackSet2),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					resource.TestCheckResourceAttrSet(resourceName, "template_url"),
				),
			},
		},
	})
}

func testAccCheckStackSetExists(resourceName string, v *cloudformation.StackSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		callAs := rs.Primary.Attributes["call_as"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

		output, err := tfcloudformation.FindStackSetByName(conn, rs.Primary.ID, callAs)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set" {
			continue
		}

		callAs := rs.Primary.Attributes["call_as"]

		_, err := tfcloudformation.FindStackSetByName(conn, rs.Primary.ID, callAs)

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

func testAccPreCheckStackSet(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

	input := &cloudformation.ListStackSetsInput{}
	_, err := conn.ListStackSets(input)

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

func testAccStackSetAdministrationRoleARN1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
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

  name = "%[1]s1"
}

resource "aws_iam_role" "test2" {
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

  name = "%[1]s2"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test1.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetAdministrationRoleARN2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test1" {
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

  name = "%[1]s1"
}

resource "aws_iam_role" "test2" {
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

  name = "%[1]s2"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test2.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  description             = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), description)
}

func testAccStackSetExecutionRoleNameConfig(rName, executionRoleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  execution_role_name     = %[3]q
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), executionRoleName)
}

func testAccStackSetNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetParameters1Config(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParameters1(rName), value1)
}

func testAccStackSetParameters2Config(rName, value1, value2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
    Parameter2 = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParameters2(rName), value1, value2)
}

func testAccStackSetParametersDefault0Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersDefault1(rName))
}

func testAccStackSetParametersDefault1Config(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersDefault1(rName), value1)
}

func testAccStackSetParametersNoEcho1Config(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  parameters = {
    Parameter1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyParametersNoEcho1(rName), value1)
}

func testAccStackSetTags1Config(rName, value1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  tags = {
    Key1 = %[3]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), value1)
}

func testAccStackSetTags2Config(rName, value1, value2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  tags = {
    Key1 = %[3]q
    Key2 = %[4]q
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, testAccStackSetTemplateBodyVPC(rName), value1, value2)
}

func testAccStackSetTemplateBodyConfig(rName, templateBody string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  template_body = <<TEMPLATE
%[2]s
TEMPLATE
}
`, rName, templateBody)
}

func testAccStackSetTemplateURL1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template1.yml"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
}
`, rName, testAccStackSetTemplateBodyVPC(rName+"1"))
}

func testAccStackSetTemplateURL2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "public-read"
}

resource "aws_s3_object" "test" {
  acl    = "public-read"
  bucket = aws_s3_bucket.test.bucket

  content = <<CONTENT
%[2]s
CONTENT

  key = "%[1]s-template2.yml"
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q
  template_url            = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
}
`, rName, testAccStackSetTemplateBodyVPC(rName+"2"))
}

func testAccStackSetPermissionModelConfig(rName string) string {
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

func testAccStackSetOperationPreferencesConfig(rName string, failureToleranceCount, maxConcurrentCount int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  operation_preferences {
    failure_tolerance_count = %[2]d
    max_concurrent_count    = %[3]d
  }

  template_body = <<TEMPLATE
%[4]s
TEMPLATE
}
`, rName, failureToleranceCount, maxConcurrentCount, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetOperationPreferencesUpdatedConfig(rName string, failureTolerancePercentage, maxConcurrentPercentage int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
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

  name = %[1]q
}

resource "aws_cloudformation_stack_set" "test" {
  administration_role_arn = aws_iam_role.test.arn
  name                    = %[1]q

  operation_preferences {
    failure_tolerance_percentage = %[2]d
    max_concurrent_percentage    = %[3]d
  }

  template_body = <<TEMPLATE
%[4]s
TEMPLATE
}
`, rName, failureTolerancePercentage, maxConcurrentPercentage, testAccStackSetTemplateBodyVPC(rName))
}
