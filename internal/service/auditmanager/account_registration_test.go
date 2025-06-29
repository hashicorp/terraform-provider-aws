// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerAccountRegistration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccAccountRegistration_basic,
		acctest.CtDisappears: testAccAccountRegistration_disappears,
		"kms key":            testAccAccountRegistration_optionalKMSKey,
		"Identity":           testAccAuditManagerAccountRegistration_IdentitySerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_auditmanager_account_registration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountRegistrationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountRegistration_disappears(t *testing.T) {
	acctest.SkipIfEnvVarNotSet(t, "AUDITMANAGER_DEREGISTER_ACCOUNT_ON_DESTROY")
	ctx := acctest.Context(t)
	resourceName := "aws_auditmanager_account_registration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				// deregister_on_destroy must be enabled for the disappears helper to disable
				// audit manager on destroy and trigger the non-empty plan after state refresh
				Config: testAccAccountRegistrationConfig_deregisterOnDestroy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfauditmanager.ResourceAccountRegistration, resourceName),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccountRegistration_optionalKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_auditmanager_account_registration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountRegistrationConfig_kmsKey(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKey),
				),
			},
			{
				Config: testAccAccountRegistrationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKey),
				),
			},
			{
				Config: testAccAccountRegistrationConfig_kmsKey(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKey),
				),
			},
		},
	})
}

func testAccCheckAccoountRegistrationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)

		_, err := tfauditmanager.FindAccountRegistration(ctx, conn)

		return err
	}
}

func testAccAuditManagerAccountRegistration_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_auditmanager_account_registration.test"

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.AuditManagerServiceID),
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAccountRegistrationConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccAccountRegistrationConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountRegistrationConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
					}),
				},
			},
		},
	})
}

func testAccAccountRegistrationConfig_basic() string {
	return `
resource "aws_auditmanager_account_registration" "test" {}
`
}

func testAccAccountRegistrationConfig_deregisterOnDestroy() string {
	return `
resource "aws_auditmanager_account_registration" "test" {
  deregister_on_destroy = true
}
`
}

func testAccAccountRegistrationConfig_kmsKey() string {
	return `
resource "aws_kms_key" "test" {
  enable_key_rotation = true
}

resource "aws_auditmanager_account_registration" "test" {
  kms_key = aws_kms_key.test.arn
}
`
}
