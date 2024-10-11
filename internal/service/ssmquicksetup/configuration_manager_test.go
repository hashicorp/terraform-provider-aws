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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMQuickSetupServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationManagerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationManagerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationManagerExists(ctx, resourceName, &cm),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "configuration_definition.*", map[string]string{
						"type": "AWSQuickSetupType-PatchPolicy",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "manager_arn", "ssm-quicksetup", regexache.MustCompile(`configuration-manager/+.`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccConfigurationManagerImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "manager_arn",
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

func testAccConfigurationManagerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["manager_arn"], nil
	}
}

func testAccConfigurationManagerConfigBase_patchPolicy() string {
	return `
data "aws_caller_identity" "current" {}
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
}
`
}

func testAccConfigurationManagerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationManagerConfigBase_patchPolicy(),
		fmt.Sprintf(`
resource "aws_ssmquicksetup_configuration_manager" "test" {
  name = %[1]q

  configuration_definition {
    local_deployment_administration_role_arn = "arn:aws:iam::727561393803:role/AWS-QuickSetup-PatchPolicy-LocalAdministrationRole"
    local_deployment_execution_role_name     = "AWS-QuickSetup-PatchPolicy-LocalExecutionRole"
    type                                     = "AWSQuickSetupType-PatchPolicy"

    parameters = {
      "ConfigurationOptionsPatchOperation" : "Scan",
      "ConfigurationOptionsScanValue" : "cron(0 1 * * ? *)",
      "ConfigurationOptionsScanNextInterval" : "false",
      "PatchBaselineRegion" : data.aws_region.current.name,
      "PatchBaselineUseDefault" : "default",
      "PatchPolicyName" : %[1]q,
      "SelectedPatchBaselines" : local.selected_patch_baselines,
      "OutputLogEnableS3" : "false",
      "RateControlConcurrency" : "10%",
      "RateControlErrorThreshold" : "2%",
      "IsPolicyAttachAllowed" : "false",
      "TargetAccounts" : data.aws_caller_identity.current.account_id,
      "TargetRegions" : data.aws_region.current.name,
      "TargetType" : "*"
    }
  }
}
`, rName))
}
