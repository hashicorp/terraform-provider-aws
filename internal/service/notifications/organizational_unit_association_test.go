// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationalUnitAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_organizational_unit_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitAssociationConfig_base(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckSleep(t, 30*time.Second),
				),
			},
			{
				Config: testAccOrganizationalUnitAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationalUnitAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organizational_unit_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "notification_configuration_arn", "organizational_unit_id"),
			},
		},
	})
}

func testAccOrganizationalUnitAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_organizational_unit_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitAssociationConfig_base(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckSleep(t, 30*time.Second),
				),
			},
			{
				Config: testAccOrganizationalUnitAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationalUnitAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnotifications.ResourceOrganizationalUnitAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckOrganizationalUnitAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_organizational_unit_association" {
				continue
			}

			_, err := tfnotifications.FindOrganizationalUnitAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes["organizational_unit_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("User Notifications Organizational Unit Association still exists")
		}

		return nil
	}
}

func testAccCheckOrganizationalUnitAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		_, err := tfnotifications.FindOrganizationalUnitAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes["organizational_unit_id"])

		return err
	}
}

func testAccOrganizationalUnitAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_notifications_organizations_access" "test" {
  enabled = true
}

data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = "test"
}
`, rName)
}

func testAccOrganizationalUnitAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOrganizationalUnitAssociationConfig_base(rName), `
resource "aws_notifications_organizational_unit_association" "test" {
  organizational_unit_id         = aws_organizations_organizational_unit.test.id
  notification_configuration_arn = aws_notifications_notification_configuration.test.arn
}
`)
}
