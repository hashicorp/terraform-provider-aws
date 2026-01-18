// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"fmt"
	"testing"

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
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
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

func testAccCheckOrganizationsAccessExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		_, err := tfnotifications.FindAccessForOrganization(ctx, conn)

		return err
	}
}

func testAccOrganizationsAccessConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_notifications_organizations_access" "test" {
  enabled = %[1]t
}
`, enabled)
}
