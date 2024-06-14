// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMemoryDBACLDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_acl.test"
	dataSourceName := "data.aws_memorydb_acl.test"
	userName1 := "tf-" + sdkacctest.RandString(8)
	userName2 := "tf-" + sdkacctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccACLDataSourceConfig_basic(rName, userName1, userName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_engine_version", resourceName, "minimum_engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "user_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "user_names.*", resourceName, "user_names.0"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "user_names.*", resourceName, "user_names.1"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
				),
			},
		},
	})
}

func testAccACLDataSourceConfig_basic(rName, userName1, userName2 string) string {
	return acctest.ConfigCompose(
		testAccACLConfigUsers(userName1, userName2),
		fmt.Sprintf(`
resource "aws_memorydb_acl" "test" {
  depends_on = [aws_memorydb_user.test]

  name       = %[1]q
  user_names = [%[2]q, %[3]q]

  tags = {
    Test = "test"
  }
}

data "aws_memorydb_acl" "test" {
  name = aws_memorydb_acl.test.name
}
`, rName, userName1, userName2),
	)
}
