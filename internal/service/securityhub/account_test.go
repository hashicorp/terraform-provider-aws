// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"
	controlFindingGeneratorDefaultValueFromAWS := "SECURITY_CONTROL"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		controlFindingGeneratorDefaultValueFromAWS = ""
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsecurityhub.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccount_enableDefaultStandardsFalse(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_enableDefaultStandardsFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccAccount_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		// control_finding_generator not supported in AWS GovCloud.
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_full(false, "STANDARD_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "STANDARD_CONTROL"),
				),
			},
			{
				Config: testAccAccountConfig_full(true, "SECURITY_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx),
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
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "enable_default_standards"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
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
	controlFindingGeneratorExpectedValue := "SECURITY_CONTROL"
	if acctest.Partition() == endpoints.AwsUsGovPartitionID {
		controlFindingGeneratorExpectedValue = ""
	}
	expectNonEmptyPlan := acctest.Partition() == endpoints.AwsUsGovPartitionID

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SecurityHubServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx),
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
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", controlFindingGeneratorExpectedValue),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", acctest.CtTrue),
				),
				ExpectNonEmptyPlan: expectNonEmptyPlan,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_basic,
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckAccountExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		awsClient := acctest.Provider.Meta().(*conns.AWSClient)
		conn := awsClient.SecurityHubClient(ctx)

		arn := tfsecurityhub.AccountHubARN(awsClient)
		_, err := tfsecurityhub.FindHubByARN(ctx, conn, arn)

		return err
	}
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		awsClient := acctest.Provider.Meta().(*conns.AWSClient)
		conn := awsClient.SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_account" {
				continue
			}

			arn := tfsecurityhub.AccountHubARN(awsClient)
			_, err := tfsecurityhub.FindHubByARN(ctx, conn, arn)

			if tfresource.NotFound(err) {
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
