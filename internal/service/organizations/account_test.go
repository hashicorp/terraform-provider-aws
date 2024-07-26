// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"close_on_deletion",
			"create_govcloud",
			"govcloud_id",
		},
	}
}

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_CloseOnDeletion(t *testing.T) {
	ctx := acctest.Context(t)
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttr(resourceName, "govcloud_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_ParentID(t *testing.T) {
	ctx := acctest.Context(t)
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	parentIdResourceName1 := "aws_organizations_organizational_unit.test1"
	parentIdResourceName2 := "aws_organizations_organizational_unit.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_parentId1(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName1, names.AttrID),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_parentId2(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName2, names.AttrID),
				),
			},
		},
	})
}

func testAccAccount_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_tags1(name, email, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_tags2(name, email, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccountConfig_tags1(name, email, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccount_govCloud(t *testing.T) {
	ctx := acctest.Context(t)
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_govCloud(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "govcloud_id"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_account" {
				continue
			}

			_, err := tforganizations.FindAccountByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AWS Organizations Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccountExists(ctx context.Context, n string, v *awstypes.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		output, err := tforganizations.FindAccountByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccountConfig_basic(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name  = %[1]q
  email = %[2]q
}
`, name, email)
}

func testAccAccountConfig_closeOnDeletion(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_parentId1(name, email string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test1.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_parentId2(name, email string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test2.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_tags1(name, email, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true

  tags = {
    %[3]q = %[4]q
  }
}
`, name, email, tagKey1, tagValue1)
}

func testAccAccountConfig_tags2(name, email, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, name, email, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAccountConfig_govCloud(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name            = %[1]q
  email           = %[2]q
  create_govcloud = true
}
`, name, email)
}
