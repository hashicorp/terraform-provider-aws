// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkterraform "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationStackSetInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_id", ""),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "stack_id"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stack_set_instance_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
			// Test import with call_as.
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", resourceName)
					}

					return fmt.Sprintf("%s,SELF", rs.Primary.ID), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudformation.ResourceStackSetInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_Disappears_stackSet(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1 awstypes.StackInstance
	var stackSet1 awstypes.StackSet
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, t, stackSetResourceName, &stackSet1),
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudformation.ResourceStackSetInstance(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcloudformation.ResourceStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_parameterOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1, stackInstance2, stackInstance3, stackInstance4 awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides1(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides2(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance2),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccStackSetInstanceConfig_parameterOverrides1(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance3),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccStackSetInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance4),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance3, &stackInstance4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_deploymentTargets(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1, stackInstance2 awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// A single-account INTERSECTION target is recorded by account ID.
				Config: testAccStackSetInstanceConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.account_filter_type", "INTERSECTION"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts_url", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAccountID, "data.aws_organizations_organization.test", "non_master_accounts.0.id"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
					"deployment_targets",
					"stack_instance_summaries",
				},
			},
			{
				// operation_preferences is updated in place with DeploymentTargets.
				Config: testAccStackSetInstanceConfig_deploymentTargetsOperationPreferences(rName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance2),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "2"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/40675.
func TestAccCloudFormationStackSetInstance_DeploymentTargets_importedAccountID(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance1, stackInstance2 awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance1),
				),
			},
			{
				// Forget the stack instance so it can be imported again.
				Config: testAccStackSetInstanceConfig_deploymentTargetsRemoved(rName),
			},
			{
				// Imported state has no deployment_targets.
				Config:       testAccStackSetInstanceConfig_deploymentTargets(rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return fmt.Sprintf("%s,%s,%s", rName, aws.ToString(stackInstance1.Account), acctest.Region()), nil
				},
				ImportStatePersist: true,
			},
			{
				// The configured block selects the imported instance: no replacement.
				// call_as and retain_stack are not imported.
				Config: testAccStackSetInstanceConfig_deploymentTargets(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// An in-place update of the imported instance.
				Config: testAccStackSetInstanceConfig_deploymentTargetsOperationPreferences(rName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance2),
					testAccCheckStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_DeploymentTargets_multipleOUs(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
			acctest.PreCheckMultipleOrganizationalUnitsWithAccounts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_multipleOUs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
					"deployment_targets",
				},
			},
			{
				// operation_preferences is updated in place with DeploymentTargets.
				Config: testAccStackSetInstanceConfig_multipleOUsOperationPreferences(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.concurrency_mode", "SOFT_FAILURE_TOLERANCE"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", "PARALLEL"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_DeploymentTargets_emptyOU(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
					"deployment_targets",
				},
			},
			{
				Config: testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_operationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_operationPreferences(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.concurrency_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_concurrencyMode(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_concurrencyMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.concurrency_mode", "SOFT_FAILURE_TOLERANCE"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.failure_tolerance_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.max_concurrent_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.0.region_concurrency_type", ""),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/32536.
func TestAccCloudFormationStackSetInstance_delegatedAdministrator(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var stackInstanceSummaries []awstypes.StackInstanceSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/member.org.stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Run a simple configuration to initialize the alternate providers
				Config: testAccStackSetConfig_delegatedAdministratorInit,
			},
			{
				PreConfig: func() {
					// Can only run check here because the provider is not available until the previous step.
					acctest.PreCheckOrganizationManagementAccountWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers))
					acctest.PreCheckIAMServiceLinkedRoleWithProvider(ctx, t, acctest.NamedProviderFunc(acctest.ProviderNameAlternate, providers), "/aws-service-role/stacksets.cloudformation.amazonaws.com")
				},
				Config: testAccStackSetInstanceConfig_delegatedAdministrator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx, t, resourceName, stackInstanceSummaries),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
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
					"retain_stack",
					"call_as",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_stackSetInstanceRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckStackSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_stackSetInstanceRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stack_set_instance_region", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_deprecatedRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstance awstypes.StackInstance
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_set_instance.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckStackSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackSetInstanceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig_deprecatedRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackSetInstanceExists(ctx, t, resourceName, &stackInstance),
					resource.TestCheckResourceAttr(resourceName, names.AttrRegion, acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stack_set_instance_region", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
					"call_as",
				},
			},
		},
	})
}

func testAccCheckStackSetInstanceExists(ctx context.Context, t *testing.T, resourceName string, v *awstypes.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFormationClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
		if err != nil {
			return err
		}

		stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
		callAs := rs.Primary.Attributes["call_as"]
		output, err := tfcloudformation.FindStackInstanceByFourPartKey(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccCheckStackSetInstanceForOrganizationalUnitExists is a variant of the
// standard CheckExistsFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackSetInstanceForOrganizationalUnitExists(ctx context.Context, t *testing.T, resourceName string, v []awstypes.StackInstanceSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudFormationClient(ctx)

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
		if err != nil {
			return err
		}

		stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
		orgIDs := strings.Split(accountOrOrgID, "/")
		callAs := rs.Primary.Attributes["call_as"]
		output, err := tfcloudformation.FindStackInstanceSummariesByFourPartKey(ctx, conn, stackSetName, region, callAs, orgIDs)

		if err != nil {
			return err
		}

		v = output

		return nil
	}
}

// testAccCheckStackSetInstanceForOrganizationalUnitDestroy is a variant of the
// standard CheckDestroyFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackSetInstanceForOrganizationalUnitDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set_instance" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
			if err != nil {
				return err
			}

			stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
			orgIDs := strings.Split(accountOrOrgID, "/")
			callAs := rs.Primary.Attributes["call_as"]
			output, err := tfcloudformation.FindStackInstanceSummariesByFourPartKey(ctx, conn, stackSetName, region, callAs, orgIDs)

			if retry.NotFound(err) {
				continue
			}
			if len(output) == 0 {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation StackSet Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackSetInstanceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_set_instance" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackSetInstanceResourceIDPartCount, false)
			if err != nil {
				return err
			}

			stackSetName, accountOrOrgID, region := parts[0], parts[1], parts[2]
			callAs := rs.Primary.Attributes["call_as"]
			_, err = tfcloudformation.FindStackInstanceByFourPartKey(ctx, conn, stackSetName, accountOrOrgID, region, callAs)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation StackSet Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackSetInstanceNotRecreated(i, j *awstypes.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.StackId) != aws.ToString(j.StackId) {
			return fmt.Errorf("CloudFormation StackSet Instance (%s,%s,%s) recreated", aws.ToString(i.StackSetId), aws.ToString(i.Account), aws.ToString(i.Region))
		}

		return nil
	}
}

func testAccStackSetInstanceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "Administration" {
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

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [aws_iam_role_policy.Execution]

  administration_role_arn = aws_iam_role.Administration.arn
  execution_role_name     = aws_iam_role.Execution.name
  name                    = %[1]q

  parameters = {
    Parameter1 = "stacksetvalue1"
    Parameter2 = "stacksetvalue2"
  }

  template_body = <<TEMPLATE
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
TEMPLATE
}
`, rName)
}

func testAccStackSetInstanceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceConfig_base(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_parameterOverrides1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1))
}

func testAccStackSetInstanceConfig_parameterOverrides2(rName, value1, value2 string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1, value2))
}

func testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "Administration" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

data "aws_organizations_organization" "test" {}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [data.aws_organizations_organization.test]

  name             = %[1]q
  permission_model = "SERVICE_MANAGED"

  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = false
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE

  lifecycle {
    ignore_changes = [administration_role_arn]
  }
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackSetInstanceConfig_deploymentTargets(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
    account_filter_type     = "INTERSECTION"
    accounts                = [data.aws_organizations_organization.test.non_master_accounts[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_deploymentTargetsRemoved(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
removed {
  from = aws_cloudformation_stack_set_instance.test

  lifecycle {
    destroy = false
  }
}
`)
}

func testAccStackSetInstanceConfig_deploymentTargetsOperationPreferences(rName string, maxConcurrentCount int) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
    account_filter_type     = "INTERSECTION"
    accounts                = [data.aws_organizations_organization.test.non_master_accounts[0].id]
  }

  operation_preferences {
    max_concurrent_count = %[1]d
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, maxConcurrentCount))
}

func testAccStackSetInstanceConfig_multipleOUs(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [
      data.aws_organizations_organizational_units.test.children[0].id,
      data.aws_organizations_organizational_units.test.children[1].id,
    ]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_multipleOUsOperationPreferences(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  operation_preferences {
    failure_tolerance_count   = 1
    max_concurrent_percentage = 100
    concurrency_mode          = "SOFT_FAILURE_TOLERANCE"
    region_concurrency_type   = "PARALLEL"
  }

  deployment_targets {
    organizational_unit_ids = [
      data.aws_organizations_organizational_units.test.children[0].id,
      data.aws_organizations_organizational_units.test.children[1].id,
    ]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_DeploymentTargets_emptyOU(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), fmt.Sprintf(`
resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  deployment_targets {
    organizational_unit_ids = [aws_organizations_organizational_unit.test.id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, rName))
}

func testAccStackSetInstanceConfig_operationPreferences(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_concurrencyMode(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
    concurrency_mode        = "SOFT_FAILURE_TOLERANCE"
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_delegatedAdministrator(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_delegatedAdministrator(rName), `
data "aws_organizations_organization" "test" {}

resource "aws_cloudformation_stack_set_instance" "test" {
  call_as = "DELEGATED_ADMIN"

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}

func testAccStackSetInstanceConfig_stackSetInstanceRegion(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_instance_region = %[1]q
  stack_set_name            = aws_cloudformation_stack_set.test.name
}
`, acctest.AlternateRegion()))
}

func testAccStackSetInstanceConfig_deprecatedRegion(rName string) string {
	return acctest.ConfigCompose(testAccStackSetInstanceConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  region         = %[1]q
  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, acctest.AlternateRegion()))
}

const (
	testStackSetInstanceAccount1 = "111111111111"
	testStackSetInstanceAccount2 = "222222222222"
	testStackSetInstanceOU1      = "ou-abcd-11111111"
	testStackSetInstanceOU2      = "ou-abcd-22222222"
)

func TestStackSetInstanceIDTarget(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		accountID string
		dt        *awstypes.DeploymentTargets
		want      string
	}{
		"account_id (no deployment_targets)": {
			accountID: testStackSetInstanceAccount1,
			want:      testStackSetInstanceAccount1,
		},
		"single account INTERSECTION": {
			accountID: testStackSetInstanceAccount2, // caller account; ignored
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			want: testStackSetInstanceAccount1,
		},
		"single account INTERSECTION across multiple OUs": {
			accountID: testStackSetInstanceAccount2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			want: testStackSetInstanceAccount1,
		},
		"single account DIFFERENCE": {
			accountID: testStackSetInstanceAccount2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeDifference,
			},
			want: testStackSetInstanceOU1,
		},
		"multiple accounts INTERSECTION": {
			accountID: testStackSetInstanceAccount2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1, testStackSetInstanceAccount2},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			want: testStackSetInstanceOU1,
		},
		"single account INTERSECTION with accounts_url": {
			accountID: testStackSetInstanceAccount2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountsUrl:           aws.String("s3://bucket/accounts.csv"),
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			want: testStackSetInstanceOU1,
		},
		"multiple OUs": {
			accountID: testStackSetInstanceAccount2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
			},
			want: testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfcloudformation.StackSetInstanceIDTarget(tc.accountID, tc.dt); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExpandStackSetInstanceUpdateTargets(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		accountOrOrgID string
		ouID           string
		dt             *awstypes.DeploymentTargets
		wantAccounts   []string
		wantTargets    *awstypes.DeploymentTargets
	}{
		"account_id on SELF_MANAGED (no deployment_targets)": {
			accountOrOrgID: testStackSetInstanceAccount1,
			wantAccounts:   []string{testStackSetInstanceAccount1},
		},
		"imported account ID on SERVICE_MANAGED (no deployment_targets)": {
			accountOrOrgID: testStackSetInstanceAccount1,
			ouID:           testStackSetInstanceOU1,
			dt:             &awstypes.DeploymentTargets{},
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
		},
		"imported account ID, organizational unit unknown (no deployment_targets)": {
			accountOrOrgID: testStackSetInstanceAccount1,
			dt:             &awstypes.DeploymentTargets{},
			wantAccounts:   []string{testStackSetInstanceAccount1},
		},
		"single account INTERSECTION": {
			accountOrOrgID: testStackSetInstanceAccount1,
			ouID:           testStackSetInstanceOU1,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
		},
		"single account INTERSECTION created as an OU ID": {
			accountOrOrgID: testStackSetInstanceOU1,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
		},
		"multiple OUs": {
			accountOrOrgID: testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
			},
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
			},
		},
		"imported multiple OUs (no deployment_targets)": {
			accountOrOrgID: testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2,
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			accounts, targets := tfcloudformation.ExpandStackSetInstanceUpdateTargets(tc.accountOrOrgID, tc.ouID, tc.dt)
			if diff := cmp.Diff(tc.wantAccounts, accounts); diff != "" {
				t.Errorf("Accounts mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantTargets, targets, cmp.AllowUnexported(awstypes.DeploymentTargets{})); diff != "" {
				t.Errorf("DeploymentTargets mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExpandStackSetInstanceDeleteTargets(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		accountOrOrgID string
		ouID           string
		dt             *awstypes.DeploymentTargets
		wantAccounts   []string
		wantTargets    *awstypes.DeploymentTargets
	}{
		"account_id on SELF_MANAGED": {
			accountOrOrgID: testStackSetInstanceAccount1,
			wantAccounts:   []string{testStackSetInstanceAccount1},
		},
		"imported account ID on SERVICE_MANAGED": {
			accountOrOrgID: testStackSetInstanceAccount1,
			ouID:           testStackSetInstanceOU1,
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
		},
		"single account INTERSECTION": {
			accountOrOrgID: testStackSetInstanceAccount1,
			ouID:           testStackSetInstanceOU1,
			dt: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
				Accounts:              []string{testStackSetInstanceAccount1},
				AccountFilterType:     awstypes.AccountFilterTypeIntersection,
			},
		},
		"imported multiple OUs": {
			accountOrOrgID: testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2,
			wantTargets: &awstypes.DeploymentTargets{
				OrganizationalUnitIds: []string{testStackSetInstanceOU1, testStackSetInstanceOU2},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			accounts, targets := tfcloudformation.ExpandStackSetInstanceDeleteTargets(tc.accountOrOrgID, tc.ouID, tc.dt)
			if diff := cmp.Diff(tc.wantAccounts, accounts); diff != "" {
				t.Errorf("Accounts mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantTargets, targets, cmp.AllowUnexported(awstypes.DeploymentTargets{})); diff != "" {
				t.Errorf("DeploymentTargets mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDeploymentTargetsEquivalentToID(t *testing.T) {
	t.Parallel()

	single := func(acct string, ous ...string) *awstypes.DeploymentTargets {
		return &awstypes.DeploymentTargets{
			OrganizationalUnitIds: ous,
			Accounts:              []string{acct},
			AccountFilterType:     awstypes.AccountFilterTypeIntersection,
		}
	}

	testCases := map[string]struct {
		accountOrOrgID string
		ouID           string
		dt             *awstypes.DeploymentTargets
		want           bool
	}{
		"no deployment_targets":                 {testStackSetInstanceAccount1, "", nil, false},
		"same account, same OU":                 {testStackSetInstanceAccount1, testStackSetInstanceOU1, single(testStackSetInstanceAccount1, testStackSetInstanceOU1), true},
		"same account, OU among several":        {testStackSetInstanceAccount1, testStackSetInstanceOU2, single(testStackSetInstanceAccount1, testStackSetInstanceOU1, testStackSetInstanceOU2), true},
		"same account, OU unknown":              {testStackSetInstanceAccount1, "", single(testStackSetInstanceAccount1, testStackSetInstanceOU1), true},
		"same account, different OU":            {testStackSetInstanceAccount1, testStackSetInstanceOU2, single(testStackSetInstanceAccount1, testStackSetInstanceOU1), false},
		"different account":                     {testStackSetInstanceAccount1, testStackSetInstanceOU1, single(testStackSetInstanceAccount2, testStackSetInstanceOU1), false},
		"account ID, OU-only deployment target": {testStackSetInstanceAccount1, testStackSetInstanceOU1, &awstypes.DeploymentTargets{OrganizationalUnitIds: []string{testStackSetInstanceOU1}}, false},
		"same OUs, different order": {
			testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2, "",
			&awstypes.DeploymentTargets{OrganizationalUnitIds: []string{testStackSetInstanceOU2, testStackSetInstanceOU1}}, true,
		},
		"OU subset": {
			testStackSetInstanceOU1 + "/" + testStackSetInstanceOU2, "",
			&awstypes.DeploymentTargets{OrganizationalUnitIds: []string{testStackSetInstanceOU1}}, false,
		},
		"OU ID, accounts filter": {
			testStackSetInstanceOU1, "", single(testStackSetInstanceAccount1, testStackSetInstanceOU1), false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := tfcloudformation.DeploymentTargetsEquivalentToID(tc.accountOrOrgID, tc.ouID, tc.dt); got != tc.want {
				t.Errorf("got %t, want %t", got, tc.want)
			}
		})
	}
}

func TestStackSetInstanceDiff_importedDeploymentTargets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	state := func(id, accountID, ouID string) *sdkterraform.InstanceState {
		return &sdkterraform.InstanceState{
			ID: id,
			Attributes: map[string]string{
				names.AttrID:                id,
				names.AttrAccountID:         accountID,
				"call_as":                   "SELF",
				"deployment_targets.#":      "0",
				"organizational_unit_id":    ouID,
				names.AttrRegion:            "us-east-1",
				"retain_stack":              acctest.CtTrue,
				"stack_set_instance_region": "us-east-1",
				"stack_set_name":            "example",
			},
		}
	}
	config := func(ous []any, accounts []any, filter string) *sdkterraform.ResourceConfig {
		dt := map[string]any{"organizational_unit_ids": ous}
		if accounts != nil {
			dt["accounts"] = accounts
		}
		if filter != "" {
			dt["account_filter_type"] = filter
		}
		return sdkterraform.NewResourceConfigRaw(map[string]any{
			"stack_set_name":            "example",
			"stack_set_instance_region": "us-east-1",
			"retain_stack":              true,
			"deployment_targets":        []any{dt},
		})
	}

	testCases := map[string]struct {
		state       *sdkterraform.InstanceState
		config      *sdkterraform.ResourceConfig
		wantReplace bool
	}{
		"imported account ID, equivalent single account INTERSECTION": {
			state:  state("example,"+testStackSetInstanceAccount1+",us-east-1", testStackSetInstanceAccount1, testStackSetInstanceOU1),
			config: config([]any{testStackSetInstanceOU1}, []any{testStackSetInstanceAccount1}, "INTERSECTION"),
		},
		"imported account ID, different account": {
			state:       state("example,"+testStackSetInstanceAccount1+",us-east-1", testStackSetInstanceAccount1, testStackSetInstanceOU1),
			config:      config([]any{testStackSetInstanceOU1}, []any{testStackSetInstanceAccount2}, "INTERSECTION"),
			wantReplace: true,
		},
		"imported account ID, different OU": {
			state:       state("example,"+testStackSetInstanceAccount1+",us-east-1", testStackSetInstanceAccount1, testStackSetInstanceOU1),
			config:      config([]any{testStackSetInstanceOU2}, []any{testStackSetInstanceAccount1}, "INTERSECTION"),
			wantReplace: true,
		},
		"imported OU IDs, equivalent multi-OU target": {
			state:  state("example,"+testStackSetInstanceOU1+"/"+testStackSetInstanceOU2+",us-east-1", "", ""),
			config: config([]any{testStackSetInstanceOU2, testStackSetInstanceOU1}, nil, ""),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diff, err := tfcloudformation.ResourceStackSetInstance().Diff(ctx, tc.state, tc.config, nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got := diff != nil && diff.RequiresNew(); got != tc.wantReplace {
				t.Errorf("RequiresNew: got %t, want %t (diff: %v)", got, tc.wantReplace, diff)
			}
		})
	}
}
