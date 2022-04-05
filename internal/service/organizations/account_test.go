package organizations_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccAccount_basic(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var account organizations.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists("aws_organizations_account.test", &account),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "joined_method"),
					acctest.CheckResourceAttrRFC3339("aws_organizations_account.test", "joined_timestamp"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "parent_id"),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "name", name),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "email", email),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "status"),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "tags.%", "0"),
				),
			},
			{
				ResourceName:      "aws_organizations_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccount_ParentID(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var account organizations.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	parentIdResourceName1 := "aws_organizations_organizational_unit.test1"
	parentIdResourceName2 := "aws_organizations_organizational_unit.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountParentId1Config(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &account),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName1, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountParentId2Config(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &account),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName2, "id"),
				),
			},
		},
	})
}

func testAccAccount_Tags(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var account organizations.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, organizations.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountTags1Config(name, email, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &account),
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
				Config: testAccAccountTags2Config(name, email, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &account),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAccountConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists("aws_organizations_account.test", &account),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAccountDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_account" {
			continue
		}

		_, err := tforganizations.FindAccountByID(conn, rs.Primary.ID)

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

func testAccCheckAccountExists(n string, v *organizations.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AWS Organizations Account ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn

		output, err := tforganizations.FindAccountByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccountConfig(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[1]q
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountParentId1Config(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test1.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountParentId2Config(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test2.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountTags1Config(name, email, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

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

func testAccAccountTags2Config(name, email, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

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
