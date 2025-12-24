// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsOrganizationsAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_notifications_organizations_access.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationsAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsAccessExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccOrganizationsAccessConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsAccessExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				Config: testAccOrganizationsAccessConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationsAccessExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckOrganizationsAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_organizations_access" {
				continue
			}

			output, err := tfnotifications.WaitOrganizationsAccessStable(ctx, conn, tfnotifications.OrganizationsAccessStableTimeout)

			if err != nil {
				return fmt.Errorf("reading User Notifications Organizations Access (%s): %w", rs.Primary.ID, err)
			}

			if output == "" {
				return fmt.Errorf("reading User Notifications Organizations Access (%s): empty response", rs.Primary.ID)
			}

			return nil
		}

		return nil
	}
}

func testAccCheckOrganizationsAccessExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		output, err := tfnotifications.WaitOrganizationsAccessStable(ctx, conn, tfnotifications.OrganizationsAccessStableTimeout)

		if err != nil {
			return fmt.Errorf("reading User Notifications Organizations Access (%s): %w", rs.Primary.ID, err)
		}

		if output == "" {
			return fmt.Errorf("reading User Notifications Organizations Access (%s): empty response", rs.Primary.ID)
		}

		if output != string(awstypes.AccessStatusEnabled) && rs.Primary.Attributes[names.AttrEnabled] == acctest.CtTrue {
			return fmt.Errorf("User Notifications Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		if output == string(awstypes.AccessStatusEnabled) && rs.Primary.Attributes[names.AttrEnabled] == acctest.CtFalse {
			return fmt.Errorf("User Notifications Organizations Access (%s): wrong setting", rs.Primary.ID)
		}

		return nil
	}
}

func testAccOrganizationsAccessConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_notifications_organizations_access" "test" {
  enabled = %[1]t
}
`, enabled)
}
