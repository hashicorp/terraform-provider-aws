// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
		CheckDestroy:             testAccCheckOrganizationsTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
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
		CheckDestroy:             testAccCheckOrganizationsTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsTagConfig(rName, "key1", "value1"),
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
		CheckDestroy:             testAccCheckOrganizationsTagDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationsTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationsTagConfig(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckOrganizationsTagDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_tag" {
				continue
			}

			identifier, key, err := tftags.GetResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tforganizations.FindTag(ctx, conn, identifier, key)

			if retry.NotFound(err) || errs.IsA[*awstypes.TargetNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("%s resource (%s) tag (%s) still exists", names.Organizations, identifier, key)
		}

		return nil
	}
}

func testAccOrganizationsTagConfig(rName string, key string, value string) string {
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
