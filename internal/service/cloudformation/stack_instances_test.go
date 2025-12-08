// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationStackInstances_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "accounts.0"),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "regions.0", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stacks", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "stack_instance_summaries.0.account_id"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.0.drift_status", "NOT_CHECKED"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.0.region", acctest.Region()),
					resource.TestCheckResourceAttrSet(resourceName, "stack_instance_summaries.0.stack_id"),
					resource.TestCheckResourceAttrSet(resourceName, "stack_instance_summaries.0.stack_set_id"),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.0.status", "CURRENT"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stacks",
				},
			},
		},
	})
}

func TestAccCloudFormationStackInstances_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackInstances(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackInstances_Disappears_stackSet(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1 tfcloudformation.StackInstances
	var stackSet1 awstypes.StackSet
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackSetExists(ctx, stackSetResourceName, &stackSet1),
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackInstances(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackInstances_Multi_increaseRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1, stackInstances2 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.Region()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "1"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "accounts.0"),
					resource.TestCheckResourceAttr(resourceName, "call_as", "SELF"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "operation_preferences.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stacks", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.Region(), acctest.AlternateRegion()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances2),
					testAccCheckStackInstancesNotRecreated(&stackInstances1, &stackInstances2),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.Region()),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_Multi_decreaseRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1, stackInstances2 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.Region(), acctest.AlternateRegion()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.Region()),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.Region()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances2),
					testAccCheckStackInstancesNotRecreated(&stackInstances1, &stackInstances2),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_Multi_swapRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1, stackInstances2 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.Region()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
			{
				Config: testAccStackInstancesConfig_regions(rName, []string{acctest.AlternateRegion()}),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances2),
					testAccCheckStackInstancesNotRecreated(&stackInstances1, &stackInstances2),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "stack_instance_summaries.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_parameterOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances1, stackInstances2, stackInstances3, stackInstances4 tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckStackSet(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_parameterOverrides1(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stacks",
					"call_as",
				},
			},
			{
				Config: testAccStackInstancesConfig_parameterOverrides2(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances2),
					testAccCheckStackInstancesNotRecreated(&stackInstances1, &stackInstances2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccStackInstancesConfig_parameterOverrides1(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances3),
					testAccCheckStackInstancesNotRecreated(&stackInstances2, &stackInstances3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccStackInstancesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesExists(ctx, resourceName, &stackInstances4),
					testAccCheckStackInstancesNotRecreated(&stackInstances3, &stackInstances4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_deploymentTargets(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.account_filter_type", "INTERSECTION"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.accounts_url", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stacks",
					"call_as",
					"deployment_targets",
				},
			},
			{
				Config: testAccStackInstancesConfig_deploymentTargets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_DeploymentTargets_emptyOU(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_targets.0.organizational_unit_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stacks",
					"call_as",
					"deployment_targets",
				},
			},
			{
				Config: testAccStackInstancesConfig_DeploymentTargets_emptyOU(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
				),
			},
		},
	})
}

func TestAccCloudFormationStackInstances_operationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_operationPreferences(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
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

func TestAccCloudFormationStackInstances_concurrencyMode(t *testing.T) {
	ctx := acctest.Context(t)
	var stackInstances tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackInstancesConfig_concurrencyMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
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
func TestAccCloudFormationStackInstances_delegatedAdministrator(t *testing.T) {
	ctx := acctest.Context(t)
	providers := make(map[string]*schema.Provider)
	var stackInstances tfcloudformation.StackInstances
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack_instances.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckStackSet(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationMemberAccount(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/member.org.stacksets.cloudformation.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationEndpointID, "organizations"),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesNamedAlternate(ctx, t, providers),
		CheckDestroy:             testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx),
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
				Config: testAccStackInstancesConfig_delegatedAdministrator(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStackInstancesForOrganizationalUnitExists(ctx, resourceName, stackInstances),
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
					"retain_stacks",
					"call_as",
				},
			},
		},
	})
}

func testAccCheckStackInstancesExists(ctx context.Context, resourceName string, v *tfcloudformation.StackInstances) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackInstancesResourceIDPartCount, true)
		if err != nil {
			return err
		}

		stackSetName := parts[0]
		callAs := rs.Primary.Attributes["call_as"]

		var accounts []string
		for i := range attributeLength(rs.Primary.Attributes["accounts.#"]) {
			accounts = append(accounts, rs.Primary.Attributes[fmt.Sprintf("accounts.%d", i)])
		}

		var regions []string
		for i := range attributeLength(rs.Primary.Attributes["regions.#"]) {
			regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("regions.%d", i)])
		}

		deployedByOU := false
		if rs.Primary.Attributes["deployment_targets.#"] != "0" && rs.Primary.Attributes["deployment_targets.0.organizational_unit_ids.#"] != "0" {
			deployedByOU = true
		}

		output, err := tfcloudformation.FindStackInstancesByNameCallAs(ctx, acctest.Provider.Meta(), stackSetName, callAs, deployedByOU, accounts, regions)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func attributeLength(attribute string) int {
	return errs.Must(strconv.Atoi(attribute)) // nosemgrep: ci.avoid-errs-Must
}

// testAccCheckStackInstancesForOrganizationalUnitExists is a variant of the
// standard CheckExistsFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackInstancesForOrganizationalUnitExists(ctx context.Context, resourceName string, v tfcloudformation.StackInstances) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackInstancesResourceIDPartCount, false)
		if err != nil {
			return err
		}

		stackSetName := parts[0]
		callAs := rs.Primary.Attributes["call_as"]
		var accounts []string
		for i := range attributeLength(rs.Primary.Attributes["accounts.#"]) {
			accounts = append(accounts, rs.Primary.Attributes[fmt.Sprintf("accounts.%d", i)])
		}

		var regions []string
		for i := range attributeLength(rs.Primary.Attributes["regions.#"]) {
			regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("regions.%d", i)])
		}

		deployedByOU := false
		if rs.Primary.Attributes["deployment_targets.#"] != "0" && rs.Primary.Attributes["deployment_targets.0.organizational_unit_ids.#"] != "0" {
			deployedByOU = true
		}

		output, err := tfcloudformation.FindStackInstancesByNameCallAs(ctx, acctest.Provider.Meta(), stackSetName, callAs, deployedByOU, accounts, regions)

		if err != nil {
			return err
		}

		v = output

		return nil
	}
}

// testAccCheckStackInstancesForOrganizationalUnitDestroy is a variant of the
// standard CheckDestroyFunc which expects the resource ID to contain organizational
// unit IDs rather than an account ID
func testAccCheckStackInstancesForOrganizationalUnitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_instances" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackInstancesResourceIDPartCount, false)
			if err != nil {
				return err
			}

			stackSetName := parts[0]
			callAs := rs.Primary.Attributes["call_as"]
			var accounts []string
			for i := range attributeLength(rs.Primary.Attributes["accounts.#"]) {
				accounts = append(accounts, rs.Primary.Attributes[fmt.Sprintf("accounts.%d", i)])
			}

			var regions []string
			for i := range attributeLength(rs.Primary.Attributes["regions.#"]) {
				regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("regions.%d", i)])
			}

			deployedByOU := false
			if rs.Primary.Attributes["deployment_targets.#"] != "0" && rs.Primary.Attributes["deployment_targets.0.organizational_unit_ids.#"] != "0" {
				deployedByOU = true
			}

			output, err := tfcloudformation.FindStackInstancesByNameCallAs(ctx, acctest.Provider.Meta(), stackSetName, callAs, deployedByOU, accounts, regions)

			if tfresource.NotFound(err) {
				continue
			}
			if output.StackSetID == "" {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation Stack Instances %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackInstancesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack_instances" {
				continue
			}

			parts, err := flex.ExpandResourceId(rs.Primary.ID, tfcloudformation.StackInstancesResourceIDPartCount, true)
			if err != nil {
				return err
			}

			stackSetName := parts[0]
			callAs := rs.Primary.Attributes["call_as"]
			var accounts []string
			for i := range attributeLength(rs.Primary.Attributes["accounts.#"]) {
				accounts = append(accounts, rs.Primary.Attributes[fmt.Sprintf("accounts.%d", i)])
			}

			var regions []string
			for i := range attributeLength(rs.Primary.Attributes["regions.#"]) {
				regions = append(regions, rs.Primary.Attributes[fmt.Sprintf("regions.%d", i)])
			}

			deployedByOU := false
			if rs.Primary.Attributes["deployment_targets.#"] != "0" && rs.Primary.Attributes["deployment_targets.0.organizational_unit_ids.#"] != "0" {
				deployedByOU = true
			}

			_, err = tfcloudformation.FindStackInstancesByNameCallAs(ctx, acctest.Provider.Meta(), stackSetName, callAs, deployedByOU, accounts, regions)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation Stack Instances %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckStackInstancesNotRecreated(i, j *tfcloudformation.StackInstances) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(&i.StackSetID) != aws.ToString(&j.StackSetID) {
			return fmt.Errorf("CloudFormation Stack Instances (%s vs %s) recreated", i.StackSetID, j.StackSetID)
		}
		for _, v := range i.Summaries {
			for _, w := range j.Summaries {
				if aws.ToString(v.Region) != aws.ToString(w.Region) {
					continue
				}
				if aws.ToString(v.StackId) != aws.ToString(w.StackId) {
					return fmt.Errorf("CloudFormation Stack Instances (%s) recreated:\n\tregions:\n\t\t%s\n\t\t%s\n\tstack_ids:\n\t\t%s\n\t\t%s", i.StackSetID, aws.ToString(v.Region), aws.ToString(w.Region), aws.ToString(v.StackId), aws.ToString(w.StackId))
				}
			}
		}

		return nil
	}
}

func testAccStackInstancesBaseConfig(rName string) string {
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

  depends_on = [aws_iam_role_policy.Execution]
}
`, rName)
}

func testAccStackInstancesConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig(rName), `
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`)
}

func testAccStackInstancesConfig_regions(rName string, regions []string) string {
	return acctest.ConfigCompose(
		testAccStackInstancesBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudformation_stack_instances" "test" {
  regions        = ["%[1]s"]
  stack_set_name = aws_cloudformation_stack_set.test.name

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`, strings.Join(regions, `", "`)))
}

func testAccStackInstancesConfig_parameterOverrides1(rName, value1 string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`, value1))
}

func testAccStackInstancesConfig_parameterOverrides2(rName, value1, value2 string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`, value1, value2))
}

func testAccStackInstancesBaseConfig_ServiceManagedStackSet(rName string) string {
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
  name             = %[1]q
  permission_model = "SERVICE_MANAGED"

  auto_deployment {
    enabled                          = true
    retain_stacks_on_account_removal = false
  }

  template_body = <<TEMPLATE
%[2]s
TEMPLATE

  depends_on = [data.aws_organizations_organization.test]

  lifecycle {
    ignore_changes = [administration_role_arn]
  }
}
`, rName, testAccStackSetTemplateBodyVPC(rName))
}

func testAccStackInstancesConfig_deploymentTargets(rName string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
    account_filter_type     = "INTERSECTION"
    accounts                = [data.aws_organizations_organization.test.non_master_accounts[0].id]
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`)
}

func testAccStackInstancesConfig_DeploymentTargets_emptyOU(rName string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig_ServiceManagedStackSet(rName), fmt.Sprintf(`
resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  deployment_targets {
    organizational_unit_ids = [aws_organizations_organizational_unit.test.id]
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`, rName))
}

func testAccStackInstancesConfig_operationPreferences(rName string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`)
}

func testAccStackInstancesConfig_concurrencyMode(rName string) string {
	return acctest.ConfigCompose(testAccStackInstancesBaseConfig_ServiceManagedStackSet(rName), `
resource "aws_cloudformation_stack_instances" "test" {
  stack_set_name = aws_cloudformation_stack_set.test.name

  operation_preferences {
    failure_tolerance_count = 1
    max_concurrent_count    = 10
    concurrency_mode        = "SOFT_FAILURE_TOLERANCE"
  }

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]
}
`)
}

func testAccStackInstancesConfig_delegatedAdministrator(rName string) string {
	return acctest.ConfigCompose(testAccStackSetConfig_delegatedAdministrator(rName), `
data "aws_organizations_organization" "test" {}

resource "aws_cloudformation_stack_instances" "test" {
  call_as = "DELEGATED_ADMIN"

  deployment_targets {
    organizational_unit_ids = [data.aws_organizations_organization.test.roots[0].id]
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`)
}
