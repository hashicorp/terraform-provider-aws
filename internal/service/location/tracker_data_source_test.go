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

func TestAccLocationTrackerDataSource_indexName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_tracker.test"
	resourceName := "aws_location_tracker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerDataSourceConfig_indexName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreateTime, resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "position_filtering", resourceName, "position_filtering"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracker_arn", resourceName, "tracker_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tracker_name", resourceName, "tracker_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccTrackerDataSourceConfig_indexName(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}

data "aws_location_tracker" "test" {
  tracker_name = aws_location_tracker.test.tracker_name
}
`, rName)
}
