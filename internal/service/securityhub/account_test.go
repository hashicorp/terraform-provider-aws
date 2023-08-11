// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "true"),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "SECURITY_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", "true"),
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
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
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
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_enableDefaultStandardsFalse,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "false"),
				),
			},
		},
	})
}

func testAccAccount_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securityhub_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "true"),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "SECURITY_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", "true"),
				),
			},
			{
				Config: testAccAccountConfig_full,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_default_standards", "false"),
					resource.TestCheckResourceAttr(resourceName, "control_finding_generator", "STANDARD_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "auto_enable_controls", "false"),
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
		ErrorCheck:   acctest.ErrorCheck(t, securityhub.EndpointsID),
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
				PlanOnly:                 true,
			},
		},
	})
}

func testAccCheckAccountExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub Account ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn(ctx)

		_, err := tfsecurityhub.FindStandardsSubscriptions(ctx, conn, &securityhub.GetEnabledStandardsInput{})

		return err
	}
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_account" {
				continue
			}

			_, err := tfsecurityhub.FindStandardsSubscriptions(ctx, conn, &securityhub.GetEnabledStandardsInput{})

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

const testAccAccountConfig_full = `
resource "aws_securityhub_account" "test" {
  enable_default_standards  = false
  control_finding_generator = "STANDARD_CONTROL"
  auto_enable_controls      = false
}
`
