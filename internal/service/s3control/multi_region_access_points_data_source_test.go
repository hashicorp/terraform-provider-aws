// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlMultiRegionAccessPointsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3control_multi_region_access_points.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointsDataSourceConfig_basic(bucketName, rName),
				// Use traditional Check block here as the knownvalue.ListPartial check requires
				// the index at which the partially known object is stored. When run in parallel
				// this test configuration may return many multi-region access points in an unpredictable order.
				Check: resource.ComposeTestCheckFunc(
					// Check that we have at least one multi-region access point with the expected name
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "access_points.*", map[string]string{
						names.AttrName: rName,
					}),
				),
			},
		},
	})
}

func testAccMultiRegionAccessPointsDataSourceConfig_basic(bucketName, rName string) string {
	return acctest.ConfigCompose(
		testAccMultiRegionAccessPointConfig_basic(bucketName, rName),
		`
data "aws_s3control_multi_region_access_points" "test" {
  depends_on = [aws_s3control_multi_region_access_point.test]
}
`)
}
