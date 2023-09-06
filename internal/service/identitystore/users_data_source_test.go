// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfidentitystore "github.com/hashicorp/terraform-provider-aws/internal/service/identitystore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreUsersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	email := acctest.RandomEmailAddress(acctest.RandomDomainName())
	dataSourceName := "data.aws_identitystore_users.test"
	userResourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// All destroy checks are already covered in sibling tests
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigUsers_basic(name, email),
				// IdentityStore instances are account wide and there's no reasonable
				// way to pick a specific resource with the built-in helpers so, custom
				// test function time.
				Check: testAccCheckUsersAttributes(dataSourceName, userResourceName),
			},
		},
	})
}

func testAccCheckUsersAttributes(dsName, rsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameUser, rsName, errors.New("not found"))
		}

		ds, ok := s.RootModule().Resources[dsName]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.DSNameUsers, dsName, errors.New("not found"))
		}

		for i, _ := strconv.Atoi(ds.Primary.Attributes["users.#"]); i >= 0; i-- {
			if ds.Primary.Attributes[fmt.Sprintf("users.%d.user_id", i)] == rs.Primary.Attributes["user_id"] &&
				ds.Primary.Attributes[fmt.Sprintf("users.%d.display_name", i)] == rs.Primary.Attributes["display_name"] {
				return nil
			}
		}

		return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.DSNameUsers, dsName, errors.New("couldn't find ruser in output users or attributes were mismatched"))
	}
}

func testAccConfigUsers_basic(name, email string) string {
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

data "aws_identitystore_users" "test" {
  depends_on = [aws_identitystore_user.test]

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name, email)
}
