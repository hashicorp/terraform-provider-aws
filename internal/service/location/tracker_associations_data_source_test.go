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

func TestAccLocationTrackerAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_tracker_associations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrackerAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrackerAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "consumer_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "consumer_arns.0", "aws_location_tracker_association.test", "consumer_arn"),
				),
			},
		},
	})
}

func testAccTrackerAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_geofence_collection" "test" {
  collection_name = %[1]q
}

resource "aws_location_tracker" "test" {
  tracker_name = %[1]q
}

resource "aws_location_tracker_association" "test" {
  consumer_arn = aws_location_geofence_collection.test.collection_arn
  tracker_name = aws_location_tracker.test.tracker_name
}

data "aws_location_tracker_associations" "test" {
  tracker_name = aws_location_tracker_association.test.tracker_name
}
`, rName)
}
