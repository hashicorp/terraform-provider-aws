// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/account/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAlternateContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlternateContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlternateContactConfig_basic(rName1, emailAddress1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAlternateContactConfig_basic(rName2, emailAddress2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName2),
				),
			},
		},
	})
}

func testAccAlternateContact_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlternateContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlternateContactConfig_basic(rName, emailAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfaccount.ResourceAlternateContact(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAlternateContact_accountID(t *testing.T) { // nosemgrep:ci.account-in-func-name
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAlternateContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlternateContactConfig_organization(rName1, emailAddress1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAlternateContactConfig_organization(rName2, emailAddress2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName2),
				),
			},
		},
	})
}

func testAccCheckAlternateContactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_account_alternate_contact" {
				continue
			}

			accountID, contactType, err := tfaccount.AlternateContactParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfaccount.FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Account Alternate Contact %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAlternateContactExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account Alternate Contact ID is set")
		}

		accountID, contactType, err := tfaccount.AlternateContactParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

		_, err = tfaccount.FindAlternateContactByTwoPartKey(ctx, conn, accountID, contactType)

		return err
	}
}

func testAccAlternateContactConfig_basic(rName, emailAddress string) string {
	return fmt.Sprintf(`
resource "aws_account_alternate_contact" "test" {
  alternate_contact_type = "OPERATIONS"

  email_address = %[2]q
  name          = %[1]q
  phone_number  = "+17031235555"
  title         = %[1]q
}
`, rName, emailAddress)
}

func testAccAlternateContactConfig_organization(rName, emailAddress string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "test" {
  provider = "awsalternate"
}

resource "aws_account_alternate_contact" "test" {
  account_id             = data.aws_caller_identity.test.account_id
  alternate_contact_type = "OPERATIONS"

  email_address = %[2]q
  name          = %[1]q
  phone_number  = "+17031235555"
  title         = %[1]q
}
`, rName, emailAddress))
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountClient(ctx)

	_, err := tfaccount.FindAlternateContactByTwoPartKey(ctx, conn, "", string(types.AlternateContactTypeOperations))

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !tfresource.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
