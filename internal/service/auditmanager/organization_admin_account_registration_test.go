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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfauditmanager "github.com/hashicorp/terraform-provider-aws/internal/service/auditmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAuditManagerOrganizationAdminAccountRegistration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccOrganizationAdminAccountRegistration_basic,
		acctest.CtDisappears: testAccOrganizationAdminAccountRegistration_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOrganizationAdminAccountRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	adminAccountID := acctest.SkipIfEnvVarNotSet(t, "AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID")
	resourceName := "aws_auditmanager_organization_admin_account_registration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountRegistrationExists(ctx, t, resourceName),
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

func testAccOrganizationAdminAccountRegistration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	adminAccountID := acctest.SkipIfEnvVarNotSet(t, "AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID")
	resourceName := "aws_auditmanager_organization_admin_account_registration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountRegistrationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfauditmanager.ResourceOrganizationAdminAccountRegistration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_organization_admin_account_registration" {
				continue
			}

			_, err := tfauditmanager.FindOrganizationAdminAccount(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Audit Manager Organization Admin Account Registration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckOrganizationAdminAccountRegistrationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AuditManagerClient(ctx)

		_, err := tfauditmanager.FindOrganizationAdminAccount(ctx, conn)

		return err
	}
}

func testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_organization_admin_account_registration" "test" {
  admin_account_id = %[1]q
}
`, adminAccountID)
}
