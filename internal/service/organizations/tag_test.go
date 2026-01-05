// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOrganizationsTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
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

func TestAccOrganizationsTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOrganizationsTag_Value(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagConfig(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, acctest.CtKey1),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, acctest.CtValue1Updated),
				),
			},
		},
	})
}

func testAccTagConfig(rName string, key string, value string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_organizations_tag" "test" {
  resource_id = aws_organizations_organizational_unit.test.id
  key         = %[2]q
  value       = %[3]q
}
`, rName, key, value)
}
