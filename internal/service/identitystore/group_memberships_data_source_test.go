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

func TestAccIdentityStoreGroupMembershipsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // Group Name
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) // User Name
	dataSourceName := "data.aws_identitystore_group_memberships.test"
	groupResourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroupMemberships_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "group_memberships.#", 0),
					resource.TestCheckResourceAttrPair(dataSourceName, "group_memberships.0.group_id", groupResourceName, "group_id"),
				),
			},
		},
	})
}

func testAccConfigGroupMemberships_basic(groupName, userName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = "Acceptance Test"
  user_name         = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[2]q
  description       = "Acceptance Test"
}

resource "aws_identitystore_group_membership" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  group_id          = aws_identitystore_group.test.group_id
  member_id         = aws_identitystore_user.test.user_id
}

data "aws_identitystore_group_memberships" "test" {
  depends_on = [aws_identitystore_group_membership.test]

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  group_id          = aws_identitystore_group.test.group_id
}
`, userName, groupName)
}
