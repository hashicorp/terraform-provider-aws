// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"
	controlFindingGeneratorDefaultValueFromAWS := "SECURITY_CONTROL"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		controlFindingGeneratorDefaultValueFromAWS = ""
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", controlFindingGeneratorDefaultValueFromAWS),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_default_standards"},
			},
		},
	})
}

func testAccAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsecurityhub.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccount_enableDefaultStandardsFalse(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_enableDefaultStandardsFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccAccount_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		// control_finding_generator not supported in AWS GovCloud.
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_full(false, "STANDARD_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "STANDARD_CONTROL"),
				),
			},
			{
				Config: testAccAccountConfig_full(true, "SECURITY_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "SECURITY_CONTROL"),
				),
			},
		},
	})
}

func testAccAccount_migrateV0(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.59.0",
					},
				},
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "enable_default_standards"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", acctest.CtTrue),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/33039 et al.
func testAccAccount_removeControlFindingGeneratorDefaultValue(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.13.0",
					},
				},
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("auto_enable_controls"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("control_finding_generator"), knownvalue.StringExact("SECURITY_CONTROL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enable_default_standards"), knownvalue.Bool(true)),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckAccountExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		awsClient := acctest.ProviderMeta(ctx, t)
		conn := awsClient.SecurityHubClient(ctx)

		arn := tfsecurityhub.AccountHubARN(ctx, awsClient)
		_, err := tfsecurityhub.FindHubByARN(ctx, conn, arn)

		return err
	}
}

func testAccCheckAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		awsClient := acctest.ProviderMeta(ctx, t)
		conn := awsClient.SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_account" {
				continue
			}

			arn := tfsecurityhub.AccountHubARN(ctx, awsClient)
			_, err := tfsecurityhub.FindHubByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccAccountConfig_basic = `
resource "aws_securityhub_account" "test" {}
`

const testAccAccountConfig_enableDefaultStandardsFalse = `
resource "aws_securityhub_account" "test" {
  enable_default_standards = false
}
`

func testAccAccountConfig_full(autoEnableControls bool, controlFindingGenerator string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {
  control_finding_generator = %[2]q
  auto_enable_controls      = %[1]t
}
`, autoEnableControls, controlFindingGenerator)
}
