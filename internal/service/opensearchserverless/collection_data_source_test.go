// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessCollectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection types.CollectionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_collection.test"
	resourceName := "aws_opensearchserverless_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionDataSourceConfig_basic(rName, "encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, dataSourceName, &collection),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_endpoint", resourceName, "collection_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dashboard_endpoint", resourceName, "dashboard_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyARN, resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "standby_replicas", resourceName, "standby_replicas"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessCollectionDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var collection types.CollectionDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_collection.test"
	resourceName := "aws_opensearchserverless_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionDataSourceConfig_name(rName, "encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(ctx, dataSourceName, &collection),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_endpoint", resourceName, "collection_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dashboard_endpoint", resourceName, "dashboard_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyARN, resourceName, names.AttrKMSKeyARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
				),
			},
		},
	})
}

func testAccCollectionDataSourceBaseConfig(rName, policyType string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name = %[1]q
  type = %[2]q
  policy = jsonencode({
    Rules = [
      {
        Resource = [
          "collection/%[1]s"
        ],
        ResourceType = "collection"
      }
    ],
    AWSOwnedKey = true
  })
}

resource "aws_opensearchserverless_collection" "test" {
  name       = %[1]q
  depends_on = [aws_opensearchserverless_security_policy.test]
}
`, rName, policyType)
}

func testAccCollectionDataSourceConfig_basic(rName, policyType string) string {
	return acctest.ConfigCompose(
		testAccCollectionDataSourceBaseConfig(rName, policyType),
		`
data "aws_opensearchserverless_collection" "test" {
  id = aws_opensearchserverless_collection.test.id
}
`)
}

func testAccCollectionDataSourceConfig_name(rName, policyType string) string {
	return acctest.ConfigCompose(
		testAccCollectionDataSourceBaseConfig(rName, policyType),
		`
data "aws_opensearchserverless_collection" "test" {
  name = aws_opensearchserverless_collection.test.name
}
`)
}
