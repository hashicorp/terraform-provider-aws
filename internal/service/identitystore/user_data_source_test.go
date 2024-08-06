// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_basic(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(dataSourceName, "addresses.0", resourceName, "addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "emails.0", resourceName, "emails.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "external_ids.#", resourceName, "external_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, dataSourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name.0", resourceName, "name.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "nickname", resourceName, "nickname"),
					resource.TestCheckResourceAttrPair(dataSourceName, "phone_numbers.0", resourceName, "phone_numbers.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "preferred_language", resourceName, "preferred_language"),
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_url", resourceName, "profile_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timezone", resourceName, "timezone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "title", resourceName, "title"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrUserName, name),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_type", resourceName, "user_type"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_filterUserName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_identitystore_user.test"
	resourceName := "aws_identitystore_user.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_filterUserName(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrUserName, name),
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
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_uniqueAttributeUserName(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrUserName, name),
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
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_email(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrUserName, name),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_userID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_identitystore_user.test"
	resourceName := "aws_identitystore_user.test"
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDataSourceConfig_id(name, email),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUserName, resourceName, names.AttrUserName),
				),
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

func testAccUserDataSourceConfig_filterUserName(name, email string) string {
	return acctest.ConfigCompose(testAccUserDataSourceConfig_base(name, email), `
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  filter {
    attribute_path  = "UserName"
    attribute_value = aws_identitystore_user.test.user_name
  }
}
`)
}

func testAccUserDataSourceConfig_uniqueAttributeUserName(name, email string) string {
	return acctest.ConfigCompose(testAccUserDataSourceConfig_base(name, email), `
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "UserName"
      attribute_value = aws_identitystore_user.test.user_name
    }
  }
}
`)
}

func testAccUserDataSourceConfig_email(name, email string) string {
	return acctest.ConfigCompose(testAccUserDataSourceConfig_base(name, email), `
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  alternate_identifier {
    unique_attribute {
      attribute_path  = "Emails.Value"
      attribute_value = aws_identitystore_user.test.emails[0].value
    }
  }
}
`)
}

func testAccUserDataSourceConfig_id(name, email string) string {
	return acctest.ConfigCompose(testAccUserDataSourceConfig_base(name, email), `
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  filter {
    attribute_path  = "UserName"
    attribute_value = aws_identitystore_user.test.user_name
  }

  user_id = aws_identitystore_user.test.user_id
}
`)
}
