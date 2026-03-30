// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package identitystore_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIdentityStoreGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_identitystore_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IdentityStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigGroups_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "groups.#", 0),
				),
			},
		},
	})
}

func testAccConfigGroups_basic(groupName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_identitystore_group" "test" {
  identity_store_id = data.aws_ssoadmin_instances.test.identity_store_ids[0]
  display_name      = %[1]q
  description       = "Acceptance Test"
}

data "aws_identitystore_groups" "test" {
  depends_on = [aws_identitystore_group.test]

  identity_store_id = data.aws_ssoadmin_instances.test.identity_store_ids[0]
}
`, groupName)
}
