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

func TestAccIdentityStoreGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Group Name
	dataSourceName := "data.aws_identitystore_groups.test"
	groupResourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// All destroy checks are already covered in sibling tests
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroups_basic(rName),
				// IdentityStore instances are account wide and there's no reasonable
				// way to pick a specific resource with the built-in helpers so, custom
				// test function time.
				Check: testAccCheckGroupsAttributes(dataSourceName, groupResourceName),
			},
		},
	})
}

func testAccCheckGroupsAttributes(dsName, rsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.ResNameGroup, rsName, errors.New("not found"))
		}

		ds, ok := s.RootModule().Resources[dsName]
		if !ok {
			return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.DSNameGroups, dsName, errors.New("not found"))
		}

		for i, _ := strconv.Atoi(ds.Primary.Attributes["groups.#"]); i > 0; i-- {
			if ds.Primary.Attributes[fmt.Sprintf("groups.%d.group_id", i-1)] == rs.Primary.Attributes["group_id"] &&
				ds.Primary.Attributes[fmt.Sprintf("groups.%d.display_name", i-1)] == rs.Primary.Attributes["display_name"] &&
				ds.Primary.Attributes[fmt.Sprintf("groups.%d.description", i-1)] == rs.Primary.Attributes["description"] {
				return nil
			}
		}

		return create.Error(names.IdentityStore, create.ErrActionCheckingExistence, tfidentitystore.DSNameGroups, dsName, errors.New("couldn't find rgroup in output groups or attributes were mismatched"))
	}
}

func testAccConfigGroups_basic(groupName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
  description       = "Acceptance Test"
}

data "aws_identitystore_groups" "test" {
  depends_on = [aws_identitystore_group.test]

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, groupName)
}
