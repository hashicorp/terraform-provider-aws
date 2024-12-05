// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup"
	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssmquicksetup "github.com/hashicorp/terraform-provider-aws/internal/service/ssmquicksetup"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMQuickSetupConfigurationManager_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cm ssmquicksetup.GetConfigurationManagerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmquicksetup_configuration_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMQuickSetupEndpointID)
			testAccConfigurationManagerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_definition.*", map[string]string{
						names.AttrType: "AWSQuickSetupType-PatchPolicy",
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "manager_arn", "ssm-quicksetup", regexache.MustCompile(`configuration-manager/+.`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "manager_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "manager_arn",
				ImportStateVerifyIgnore:              []string{"status_summaries"},
			},
		},
	})
}

func TestAccSSMQuickSetupConfigurationManager_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cm ssmquicksetup.GetConfigurationManagerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmquicksetup_configuration_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMQuickSetupEndpointID)
			testAccConfigurationManagerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssmquicksetup.ResourceConfigurationManager, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSMQuickSetupConfigurationManager_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cm ssmquicksetup.GetConfigurationManagerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmquicksetup_configuration_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMQuickSetupEndpointID)
			testAccConfigurationManagerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_description(rName, "foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "foo"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "manager_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "manager_arn",
				ImportStateVerifyIgnore:              []string{"status_summaries"},
			},
			{
				Config: testAccConfigurationManagerConfig_description(rName, "bar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "bar"),
				),
			},
		},
	})
}

func TestAccSSMQuickSetupConfigurationManager_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cm ssmquicksetup.GetConfigurationManagerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmquicksetup_configuration_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMQuickSetupEndpointID)
			testAccConfigurationManagerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_parameters(rName, "10%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_definition.*", map[string]string{
						names.AttrType: "AWSQuickSetupType-PatchPolicy",
					}),
					resource.TestCheckResourceAttr(resourceName, "configuration_definition.0.parameters.PatchPolicyName", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration_definition.0.parameters.RateControlConcurrency", "10%"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "manager_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "manager_arn",
				ImportStateVerifyIgnore:              []string{"status_summaries"},
			},
			{
				Config: testAccConfigurationManagerConfig_parameters(rName, "15%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_definition.*", map[string]string{
						names.AttrType: "AWSQuickSetupType-PatchPolicy",
					}),
					resource.TestCheckResourceAttr(resourceName, "configuration_definition.0.parameters.PatchPolicyName", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration_definition.0.parameters.RateControlConcurrency", "15%"),
				),
			},
		},
	})
}

func TestAccSSMQuickSetupConfigurationManager_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cm ssmquicksetup.GetConfigurationManagerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmquicksetup_configuration_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMQuickSetupEndpointID)
			testAccConfigurationManagerPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "manager_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "manager_arn",
				ImportStateVerifyIgnore:              []string{"status_summaries"},
			},
			{
				Config: testAccConfigurationManagerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccConfigurationManagerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// Confirms the IAM roles required to execute configuration manager acceptance tests
// are present
//
// The following IAM roles must exist in the account to facilitate execution of
// the CloudFormation templates used to provision patch policies. The test
// configuration __could__ create customer managed roles with these same permissions,
// but due to the complexity of the permissions involved and potential for drift
// it was deemed preferable to rely on the AWS generated roles instead.
//
// To trigger creation of these roles, navigate to the	SSM service, select the
// QuickSetup item in the left navbar, and choose the "Patch Manager" type. Leave the
// default configuration in place. Under the "Local deployment roles" section, be
// sure that "User new IAM local deployment roles" is selected. After clicking
// "Create", the new roles will be created prior to the PatchPolicy CloudFormation
// template being executed. The roles will now be available in the account for use with
// acceptance testing.
func testAccConfigurationManagerPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckHasIAMRole(ctx, t, "AWS-QuickSetup-PatchPolicy-LocalAdministrationRole")
	acctest.PreCheckHasIAMRole(ctx, t, "AWS-QuickSetup-PatchPolicy-LocalExecutionRole")
}

func testAccCheckConfigurationManagerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMQuickSetupClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssmquicksetup_configuration_manager" {
				continue
			}
			managerARN := rs.Primary.Attributes["manager_arn"]

			_, err := tfssmquicksetup.FindConfigurationManagerByID(ctx, conn, managerARN)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSMQuickSetup, create.ErrActionCheckingDestroyed, tfssmquicksetup.ResNameConfigurationManager, managerARN, err)
			}

			return create.Error(names.SSMQuickSetup, create.ErrActionCheckingDestroyed, tfssmquicksetup.ResNameConfigurationManager, managerARN, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckConfigurationManagerExists(ctx context.Context, name string, configurationmanager *ssmquicksetup.GetConfigurationManagerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSMQuickSetup, create.ErrActionCheckingExistence, tfssmquicksetup.ResNameConfigurationManager, name, errors.New("not found"))
		}

		managerARN := rs.Primary.Attributes["manager_arn"]
		if managerARN == "" {
			return create.Error(names.SSMQuickSetup, create.ErrActionCheckingExistence, tfssmquicksetup.ResNameConfigurationManager, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMQuickSetupClient(ctx)

		out, err := tfssmquicksetup.FindConfigurationManagerByID(ctx, conn, managerARN)
		if err != nil {
			return create.Error(names.SSMQuickSetup, create.ErrActionCheckingExistence, tfssmquicksetup.ResNameConfigurationManager, managerARN, err)
		}

		*configurationmanager = *out

		return nil
	}
}

func testAccConfigurationManagerConfigBase_patchPolicy(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_ssm_patch_baselines" "test" {
  default_baselines = true
}

locals {
  # transform the output of the aws_ssm_patch_baselines data source
  # into the format expected by the SelectedPatchBaselines parameter
  selected_patch_baselines = jsonencode({
    for baseline in data.aws_ssm_patch_baselines.test.baseline_identities : baseline.operating_system => {
      "value" : baseline.baseline_id
      "label" : baseline.baseline_name
      "description" : baseline.baseline_description
      "disabled" : !baseline.default_baseline
    }
  })

  local_deployment_administration_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/AWS-QuickSetup-PatchPolicy-LocalAdministrationRole"
  local_deployment_execution_role_name     = "AWS-QuickSetup-PatchPolicy-LocalExecutionRole"

  # workaround - terrafmt cannot parse verbs inside nested maps
  patch_policy_name = %[1]q
}
`, rName)
}

func testAccConfigurationManagerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(rName),
		fmt.Sprintf(`
resource "aws_ssmquicksetup_configuration_manager" "test" {
  name = %[1]q

  configuration_definition {
    local_deployment_administration_role_arn = local.local_deployment_administration_role_arn
    local_deployment_execution_role_name     = local.local_deployment_execution_role_name
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : local.patch_policy_name,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : "10%%",
      "RateControlErrorThreshold" : "2%%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }
}
`, rName))
}

func testAccConfigurationManagerConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(rName),
		fmt.Sprintf(`
resource "aws_ssmquicksetup_configuration_manager" "test" {
  name        = %[1]q
  description = %[2]q

  configuration_definition {
    local_deployment_administration_role_arn = local.local_deployment_administration_role_arn
    local_deployment_execution_role_name     = local.local_deployment_execution_role_name
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : local.patch_policy_name,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : "10%%",
      "RateControlErrorThreshold" : "2%%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }
}
`, rName, description))
}

func testAccConfigurationManagerConfig_parameters(rName, rateControlConcurrency string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(rName),
		fmt.Sprintf(`
locals {
  # workaround - terrafmt cannot parse verbs inside nested maps
  rate_control_concurrency = %[2]q
}

resource "aws_ssmquicksetup_configuration_manager" "test" {
  name = %[1]q

  configuration_definition {
    local_deployment_administration_role_arn = local.local_deployment_administration_role_arn
    local_deployment_execution_role_name     = local.local_deployment_execution_role_name
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : local.patch_policy_name,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : local.rate_control_concurrency,
      "RateControlErrorThreshold" : "2%%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }
}
`, rName, rateControlConcurrency))
}

func testAccConfigurationManagerConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(rName),
		fmt.Sprintf(`
resource "aws_ssmquicksetup_configuration_manager" "test" {
  name = %[1]q

  configuration_definition {
    local_deployment_administration_role_arn = local.local_deployment_administration_role_arn
    local_deployment_execution_role_name     = local.local_deployment_execution_role_name
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : local.patch_policy_name,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : "10%%",
      "RateControlErrorThreshold" : "2%%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccConfigurationManagerConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(rName),
		fmt.Sprintf(`
resource "aws_ssmquicksetup_configuration_manager" "test" {
  name = %[1]q

  configuration_definition {
    local_deployment_administration_role_arn = local.local_deployment_administration_role_arn
    local_deployment_execution_role_name     = local.local_deployment_execution_role_name
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : local.patch_policy_name,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : "10%%",
      "RateControlErrorThreshold" : "2%%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
