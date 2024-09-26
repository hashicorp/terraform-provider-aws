// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
	adminAccountID := os.Getenv("AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID")
	if adminAccountID == "" {
		t.Skip("Environment variable AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID is not set")
	}

	resourceName := "aws_auditmanager_organization_admin_account_registration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountRegistrationExists(ctx, resourceName),
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
	adminAccountID := os.Getenv("AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID")
	if adminAccountID == "" {
		t.Skip("Environment variable AUDITMANAGER_ORGANIZATION_ADMIN_ACCOUNT_ID is not set")
	}

	resourceName := "aws_auditmanager_organization_admin_account_registration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AuditManagerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AuditManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationAdminAccountRegistrationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfauditmanager.ResourceOrganizationAdminAccountRegistration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOrganizationAdminAccountRegistrationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_auditmanager_organization_admin_account_registration" {
				continue
			}

			out, err := conn.GetOrganizationAdminAccount(ctx, &auditmanager.GetOrganizationAdminAccountInput{})
			if err != nil {
				return err
			}
			if out.AdminAccountId != nil {
				return create.Error(names.AuditManager, create.ErrActionCheckingDestroyed, tfauditmanager.ResNameOrganizationAdminAccountRegistration, rs.Primary.ID, errors.New("not destroyed"))
			}
		}

		return nil
	}
}

func testAccCheckOrganizationAdminAccountRegistrationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameOrganizationAdminAccountRegistration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameOrganizationAdminAccountRegistration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AuditManagerClient(ctx)
		out, err := conn.GetOrganizationAdminAccount(ctx, &auditmanager.GetOrganizationAdminAccountInput{})
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameOrganizationAdminAccountRegistration, rs.Primary.ID, err)
		}
		if out == nil || aws.ToString(out.AdminAccountId) != rs.Primary.ID {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tfauditmanager.ResNameOrganizationAdminAccountRegistration, rs.Primary.ID, errors.New("unexpected admin account ID"))
		}

		return nil
	}
}

func testAccOrganizationAdminAccountRegistrationConfig_basic(adminAccountID string) string {
	return fmt.Sprintf(`
resource "aws_auditmanager_organization_admin_account_registration" "test" {
  admin_account_id = %[1]q
}
`, adminAccountID)
}
