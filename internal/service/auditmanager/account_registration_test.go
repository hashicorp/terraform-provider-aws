// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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

	acctest.Test(ctx, t, resource.TestCase{
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
					testAccCheckAccoountRegistrationExists(ctx, t, resourceName),
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

	acctest.Test(ctx, t, resource.TestCase{
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
					testAccCheckAccoountRegistrationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceAccountRegistration, resourceName),
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

	acctest.Test(ctx, t, resource.TestCase{
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
					testAccCheckAccoountRegistrationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKey),
				),
			},
			{
				Config: testAccAccountRegistrationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrKMSKey),
				),
			},
			{
				Config: testAccAccountRegistrationConfig_kmsKey(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccoountRegistrationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKey),
				),
			},
		},
	})
}

func testAccCheckAccoountRegistrationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		_, err := tfauditmanager.FindAccountRegistration(ctx, conn)

		return err
	}
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
