// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganizationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_macie2_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfigurationConfig_autoEnable(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "auto_enable", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccOrganizationConfigurationConfig_autoEnable(autoEnable bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_organization_admin_account" "test" {
  admin_account_id = data.aws_caller_identity.current.account_id
}

resource "aws_macie2_organization_configuration" "test" {
  auto_enable = %[1]t

  depends_on = [
	aws_macie2_organization_admin_account.test
  ]
}
`, autoEnable)
}
