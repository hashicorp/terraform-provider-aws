// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

func testAccOrganizationalUnit_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var unit organizations.OrganizationalUnit

	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_outest_%d", rInt)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", "0"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile(`ou/o-.+/ou-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccOrganizationalUnit_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var unit organizations.OrganizationalUnit

	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_outest_%d", rInt)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceOrganizationalUnit(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationalUnit_Name(t *testing.T) {
	ctx := acctest.Context(t)
	var unit organizations.OrganizationalUnit

	rInt := sdkacctest.RandInt()
	name1 := fmt.Sprintf("tf_outest_%d", rInt)
	name2 := fmt.Sprintf("tf_outest_%d", rInt+1)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitConfig_basic(name1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationalUnitConfig_basic(name2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "name", name2),
				),
			},
		},
	})
}

func testAccOrganizationalUnit_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	var unit organizations.OrganizationalUnit

	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_outest_%d", rInt)
	resourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationalUnitDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationalUnitConfig_tags1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationalUnitConfig_tags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOrganizationalUnitConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationalUnitExists(ctx, resourceName, &unit),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckOrganizationalUnitDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_organizational_unit" {
				continue
			}

			params := &organizations.DescribeOrganizationalUnitInput{
				OrganizationalUnitId: &rs.Primary.ID,
			}

			resp, err := conn.DescribeOrganizationalUnitWithContext(ctx, params)

			if err != nil {
				if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
					continue
				}
				if tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
					continue
				}
				return err
			}

			if resp != nil && resp.OrganizationalUnit != nil {
				return fmt.Errorf("Bad: Organizational Unit still exists: %q", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckOrganizationalUnitExists(ctx context.Context, n string, ou *organizations.OrganizationalUnit) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)
		params := &organizations.DescribeOrganizationalUnitInput{
			OrganizationalUnitId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeOrganizationalUnitWithContext(ctx, params)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
				return fmt.Errorf("Organizational Unit %q does not exist", rs.Primary.ID)
			}
			return err
		}

		if resp == nil {
			return fmt.Errorf("failed to DescribeOrganizationalUnit %q, response was nil", rs.Primary.ID)
		}

		ou = resp.OrganizationalUnit

		return nil
	}
}

func testAccOrganizationalUnitConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}
`, name)
}

func testAccOrganizationalUnitConfig_tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccOrganizationalUnitConfig_tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
