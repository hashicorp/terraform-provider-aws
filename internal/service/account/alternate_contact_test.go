package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/account"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccount "github.com/hashicorp/terraform-provider-aws/internal/service/account"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAccountAlternateContact_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, account.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAlternateContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlternateContactConfig_basic(rName1, emailAddress1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
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
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName2),
				),
			},
		},
	})
}

func TestAccAccountAlternateContact_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress := acctest.RandomEmailAddress(domain)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, account.EndpointsID),
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

func TestAccAccountAlternateContact_accountID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_account_alternate_contact.test"
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, account.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckAlternateContactDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAlternateContactConfig_organization(rName1, emailAddress1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAlternateContactExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
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
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "alternate_contact_type", "OPERATIONS"),
					resource.TestCheckResourceAttr(resourceName, "email_address", emailAddress2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "phone_number", "+17031235555"),
					resource.TestCheckResourceAttr(resourceName, "title", rName2),
				),
			},
		},
	})
}

func testAccCheckAlternateContactDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_account_alternate_contact" {
				continue
			}

			accountID, contactType, err := tfaccount.AlternateContactParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfaccount.FindAlternateContactByAccountIDAndContactType(ctx, conn, accountID, contactType)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountConn()

		_, err = tfaccount.FindAlternateContactByAccountIDAndContactType(ctx, conn, accountID, contactType)

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
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountConn()

	_, err := tfaccount.FindAlternateContactByAccountIDAndContactType(ctx, conn, "", account.AlternateContactTypeOperations)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !tfresource.NotFound(err) {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
