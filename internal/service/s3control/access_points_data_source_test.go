// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3ControlAccessPointsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3control_access_points.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointsDataSourceConfig_basic(bucketName, accessPointName),
				// Use traditional Check block here as the knownvalue.ListPartial check requires
				// the index at which the partially known object is stored. When run in parallel
				// this test configuration may return many access points in an unpredictable order.
				Check: resource.ComposeTestCheckFunc(
					// Check that we have at least one access point with the expected name
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "access_points.*", map[string]string{
						names.AttrBucket: bucketName,
						names.AttrName:   accessPointName,
					}),
				),
			},
		},
	})
}

func TestAccS3ControlAccessPointsDataSource_bucket(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	accessPointName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3control_access_points.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointsDataSourceConfig_bucket(bucketName, accessPointName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("access_points"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("access_points"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							names.AttrBucket: knownvalue.StringExact(bucketName),
							names.AttrName:   knownvalue.StringExact(accessPointName),
						}),
					})),
				},
			},
		},
	})
}

func TestAccS3ControlAccessPointsDataSource_bucketNoAccessPoints(t *testing.T) {
	ctx := acctest.Context(t)
	bucketName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_s3control_access_points.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointsDataSourceConfig_bucketNoAccessPoints(bucketName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("access_points"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccAccessPointsDataSourceConfig_basic(bucketName, accessPointName string) string {
	return acctest.ConfigCompose(
		testAccAccessPointConfig_basic(bucketName, accessPointName),
		`
data "aws_s3control_access_points" "test" {
  depends_on = ["aws_s3_access_point.test"]
}
`)
}

func testAccAccessPointsDataSourceConfig_bucket(bucketName, accessPointName string) string {
	return acctest.ConfigCompose(
		testAccAccessPointConfig_basic(bucketName, accessPointName),
		`
data "aws_s3control_access_points" "test" {
  bucket = aws_s3_access_point.test.bucket
}
`)
}

func testAccAccessPointsDataSourceConfig_bucketNoAccessPoints(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3control_access_points" "test" {
  bucket = aws_s3_bucket.test.bucket
}
`, bucketName)
}
