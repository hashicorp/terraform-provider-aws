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
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v organizations.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsEnabled(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_CloseOnDeletion(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v organizations.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsEnabled(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "govcloud_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_ParentID(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v organizations.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	parentIdResourceName1 := "aws_organizations_organizational_unit.test1"
	parentIdResourceName2 := "aws_organizations_organizational_unit.test2"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_parentId1(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName1, "id"),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_parentId2(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
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

	var v organizations.Account
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_tags1(name, email, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_tags2(name, email, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAccountConfig_tags1(name, email, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAccount_govCloud(t *testing.T) {
	key := "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN"
	orgsEmailDomain := os.Getenv(key)
	if orgsEmailDomain == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var v organizations.Account
	resourceName := "aws_organizations_account.test"
	rInt := sdkacctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsEnabled(t) },
		ErrorCheck:        acctest.ErrorCheck(t, organizations.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_govCloud(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "govcloud_id"),
				),
			},
			testAccAccountImportStep(resourceName),
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

func testAccAccountConfig_parentId2(name, email string) string {
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

func testAccAccountConfig_tags1(name, email, tagKey1, tagValue1 string) string {
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

func testAccAccountConfig_tags2(name, email, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccAccountConfig_govCloud(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name            = %[1]q
  email           = %[2]q
  create_govcloud = true
}
`, name, email)
}
