// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMSharingWithOrganization_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSharingWithOrganization_basic,
		acctest.CtDisappears: testAccSharingWithOrganization_disappears,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSharingWithOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_sharing_with_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSharingWithOrganizationConfig_basic(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSharingWithOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ram_sharing_with_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSharingWithOrganizationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfram.ResourceSharingWithOrganization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSharingWithOrganizationConfig_basic() string {
	return `
resource "aws_ram_sharing_with_organization" "test" {}
`
}
