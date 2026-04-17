// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessCollectionGroupDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_collection_group.test"
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupDataSourceConfig_byID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "standby_replicas", resourceName, "standby_replicas"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionGroupDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_collection_group.test"
	resourceName := "aws_opensearchserverless_collection_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckCollectionGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionGroupDataSourceConfig_byName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "standby_replicas", resourceName, "standby_replicas"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
				),
			},
		},
	})
}

func testAccCollectionGroupDataSourceConfig_byID(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  description      = "test description"
  standby_replicas = "ENABLED"

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }
}

data "aws_opensearchserverless_collection_group" "test" {
  id = aws_opensearchserverless_collection_group.test.id
}
`, rName)
}

func testAccCollectionGroupDataSourceConfig_byName(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_collection_group" "test" {
  name             = %[1]q
  description      = "test description"
  standby_replicas = "ENABLED"

  capacity_limits {
    max_indexing_capacity_in_ocu = 1
    max_search_capacity_in_ocu   = 1
  }
}

data "aws_opensearchserverless_collection_group" "test" {
  name = aws_opensearchserverless_collection_group.test.name
}
`, rName)
}
