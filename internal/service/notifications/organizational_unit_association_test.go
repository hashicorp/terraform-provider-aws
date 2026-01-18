// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsOrganizationalUnitAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_organizational_unit_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.12.1",
			},
		},
		CheckDestroy: testAccCheckOrganizationalUnitAssociationDestroy(ctx),
		Steps: []resource.TestStep{
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
				ImportStateIdFunc:                    testAccOrganizationalUnitAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccNotificationsOrganizationalUnitAssociation_organization(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_organizational_unit_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.12.1",
			},
		},
		CheckDestroy: testAccCheckOrganizationalUnitAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitAssociationConfig_organization(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationalUnitAssociationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "organizational_unit_id",
				ImportStateIdFunc:                    testAccOrganizationalUnitAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccNotificationsOrganizationalUnitAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_organizational_unit_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.12.1",
			},
		},
		CheckDestroy: testAccCheckOrganizationalUnitAssociationDestroy(ctx),
		Steps: []resource.TestStep{
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

			err := tfnotifications.FindOrganizationalUnitAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes["organizational_unit_id"])

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

		err := tfnotifications.FindOrganizationalUnitAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["notification_configuration_arn"], rs.Primary.Attributes["organizational_unit_id"])

		return err
	}
}

func testAccOrganizationalUnitAssociationImportStateIDFunc(n string) func(*terraform.State) (string, error) {
	return func(state *terraform.State) (string, error) {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["notification_configuration_arn"] + intflex.ResourceIdSeparator + rs.Primary.Attributes["organizational_unit_id"], nil
	}
}

func testAccOrganizationalUnitAssociationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = "example"
}
`, rName)
}

func testAccOrganizationalUnitAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationalUnitAssociationConfig_base(rName),
		fmt.Sprintf(`
resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

# Allow time for organizational unit creation to propagate
resource "time_sleep" "wait" {
  depends_on = [
    aws_organizations_organizational_unit.test,
    aws_notifications_notification_configuration.test,
  ]

  create_duration = "5s"
}

resource "aws_notifications_organizational_unit_association" "test" {
  depends_on = [time_sleep.wait]

  organizational_unit_id         = aws_organizations_organizational_unit.test.id
  notification_configuration_arn = aws_notifications_notification_configuration.test.arn
}
`, rName))
}

func testAccOrganizationalUnitAssociationConfig_organization(rName string) string {
	return acctest.ConfigCompose(
		testAccOrganizationalUnitAssociationConfig_base(rName),
		`
resource "aws_notifications_organizational_unit_association" "test" {
  organizational_unit_id         = data.aws_organizations_organization.test.roots[0].id
  notification_configuration_arn = aws_notifications_notification_configuration.test.arn
}
`)
}
