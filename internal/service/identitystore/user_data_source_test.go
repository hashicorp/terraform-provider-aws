// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/identitystore"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIdentityStoreUserDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_identitystore_user.test"
	resourceName := "aws_identitystore_user.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "display_name", resourceName, "display_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "addresses.0", resourceName, "addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "emails.0", resourceName, "emails.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "external_ids.#", resourceName, "external_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", dataSourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name.0", resourceName, "name.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "nickname", resourceName, "nickname"),
					resource.TestCheckResourceAttrPair(dataSourceName, "phone_numbers.0", resourceName, "phone_numbers.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "preferred_language", resourceName, "preferred_language"),
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_url", resourceName, "profile_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timezone", resourceName, "timezone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "title", resourceName, "title"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttr(dataSourceName, "user_name", name),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_type", resourceName, "user_type"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_uniqueAttributeUserName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_identitystore_user.test"
	resourceName := "aws_identitystore_user.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_uniqueAttributeUserName(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, "user_name", name),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_email(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_identitystore_user.test"
	resourceName := "aws_identitystore_user.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_email(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, "user_name", name),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`no Identity Store User found matching criteria`),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_externalIDConflictsWithUniqueAttribute(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig_externalIDConflictsWithUniqueAttribute,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_userIDConflictsWithExternalID(t *testing.T) {
	ctx := acctest.Context(t)
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, identitystore.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserDataSourceConfig_userIDConflictsWithUniqueAttribute(name, email),
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func testAccUserDataSourceConfig_base(name, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "Acceptance Test"
  user_name         = %[1]q

  name {
    family_name = "Acceptance"
    given_name  = "Test"
  }

  emails {
    value = %[2]q
  }
}
`, name, email)
}

func testAccUserDataSourceConfig_basic(name, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  user_name         = %[1]q

  addresses {
    country        = "US"
    formatted      = "Formatted Address 1"
    locality       = "The Locality 1"
    postal_code    = "AAA BBB 1"
    primary        = true
    region         = "The Region 1"
    street_address = "The Street Address 1"
    type           = "The Type 1"
  }

  emails {
    primary = true
    type    = "The Type 1"
    value   = %[2]q
  }

  locale = "The Locale"

  name {
    family_name      = "Acceptance"
    formatted        = "Acceptance Test"
    given_name       = "Test"
    honorific_prefix = "Dr"
    honorific_suffix = "PhD"
    middle_name      = "John"
  }

  nickname = "The Nickname"

  phone_numbers {
    primary = false
    type    = "The Type 2"
    value   = "2222222"
  }

  preferred_language = "en-US"
  profile_url        = "http://example.com"
  timezone           = "UTC"
  title              = "Mr"
  user_type          = "Member"
}

data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  user_id           = aws_identitystore_user.test.user_id
}
`, name, email)
}

func testAccUserDataSourceConfig_uniqueAttributeUserName(name, email string) string {
	return acctest.ConfigCompose(
		testAccUserDataSourceConfig_base(name, email),
		`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = aws_identitystore_user.test.user_name
    }
  }
}
`,
	)
}

func testAccUserDataSourceConfig_email(name, email string) string {
	return acctest.ConfigCompose(
		testAccUserDataSourceConfig_base(name, email),
		`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "Emails.Value"
      attribute_value = aws_identitystore_user.test.emails[0].value
    }
  }
}
`,
	)
}

const testAccUserDataSourceConfig_nonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = "does-not-exist"
    }
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`

const testAccUserDataSourceConfig_externalIDConflictsWithUniqueAttribute = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  alternate_identifier {
    external_id {
      id     = "test"
      issuer = "test"
    }

    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = "does-not-exist"
    }
  }

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`

func testAccUserDataSourceConfig_userIDConflictsWithUniqueAttribute(name, email string) string {
	return acctest.ConfigCompose(
		testAccUserDataSourceConfig_base(name, email),
		`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = aws_identitystore_user.test.user_name
    }
  }

  user_id = aws_identitystore_user.test.user_id
}
`)
}
