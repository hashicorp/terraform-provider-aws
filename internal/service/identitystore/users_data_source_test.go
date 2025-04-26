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

func TestAccIdentityStoreUsersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rEmail := acctest.RandomEmailAddress(acctest.RandomDomainName())
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_identitystore_users.test"
	userResourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigUsersDataSource_basic(rName, rEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "users.#", 0),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.user_id", userResourceName, "user_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.display_name", userResourceName, names.AttrDisplayName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.user_name", userResourceName, names.AttrUserName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.name.0.family_name", userResourceName, "name.0.family_name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.name.0.given_name", userResourceName, "name.0.given_name"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "users.*.emails.0.value", userResourceName, "emails.0.value"),
				),
			},
		},
	})
}

func testAccConfigUsersDataSource_basic(name, email string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = data.aws_ssoadmin_instances.test.identity_store_ids[0]
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

data "aws_identitystore_users" "test" {
  identity_store_id = aws_identitystore_user.test.identity_store_id
}
`, name, email)
}
