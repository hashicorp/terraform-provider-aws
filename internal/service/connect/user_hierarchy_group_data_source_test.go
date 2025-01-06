// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccUserHierarchyGroupDataSource_hierarchyGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"
	resourceName2 := "aws_connect_user_hierarchy_group.parent"
	datasourceName := "data.aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupDataSourceConfig_groupID(rName, rName2, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_group_id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.#", resourceName, "hierarchy_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.arn", resourceName2, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.id", resourceName2, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.name", resourceName2, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, "level_id", resourceName, "level_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccUserHierarchyGroupDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"
	resourceName2 := "aws_connect_user_hierarchy_group.parent"
	datasourceName := "data.aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupDataSourceConfig_name(rName, rName2, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_group_id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.#", resourceName, "hierarchy_path.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.arn", resourceName2, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.id", resourceName2, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_one.0.name", resourceName2, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "hierarchy_path.0.level_two.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, "level_id", resourceName, "level_id"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccUserHierarchyGroupBaseDataSourceConfig(rName, rName2, rName3 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = "levelone"
    }

    level_two {
      name = "leveltwo"
    }

    level_three {
      name = "levelthree"
    }

    level_four {
      name = "levelfour"
    }

    level_five {
      name = "levelfive"
    }
  }
}

resource "aws_connect_user_hierarchy_group" "parent" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q

  tags = {
    "Name" = "Test User Hierarchy Group Parent"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}

resource "aws_connect_user_hierarchy_group" "test" {
  instance_id     = aws_connect_instance.test.id
  name            = %[3]q
  parent_group_id = aws_connect_user_hierarchy_group.parent.hierarchy_group_id

  tags = {
    "Name" = "Test User Hierarchy Group Child"
  }
}
`, rName, rName2, rName3)
}

func testAccUserHierarchyGroupDataSourceConfig_groupID(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupBaseDataSourceConfig(rName, rName2, rName3),
		`
data "aws_connect_user_hierarchy_group" "test" {
  instance_id        = aws_connect_instance.test.id
  hierarchy_group_id = aws_connect_user_hierarchy_group.test.hierarchy_group_id
}
`)
}

func testAccUserHierarchyGroupDataSourceConfig_name(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupBaseDataSourceConfig(rName, rName2, rName3),
		`
data "aws_connect_user_hierarchy_group" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_user_hierarchy_group.test.name
}
`)
}
