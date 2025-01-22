// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationGeofenceCollectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_geofence_collection.test"
	resourceName := "aws_location_geofence_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGeofenceCollectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGeofenceCollectionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGeofenceCollectionExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_arn", resourceName, "collection_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "collection_name", resourceName, "collection_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreateTime, resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccGeofenceCollectionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

data "aws_location_geofence_collection" "test" {
  collection_name = aws_location_geofence_collection.test.collection_name
}
`, rName)
}
