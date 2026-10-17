// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBGIMinorVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_odb_gi_minor_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGIMinorVersionsConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "gi_minor_versions.0.version"),
				),
			},
		},
	})
}

func TestAccODBGIMinorVersionsDataSource_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_odb_gi_minor_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGIMinorVersionsConfig_availabilityZone,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "gi_minor_versions.0.grid_image_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "gi_minor_versions.0.version"),
				),
			},
		},
	})
}

func TestAccODBGIMinorVersionsDataSource_availabilityZoneID(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_odb_gi_minor_versions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGIMinorVersionsConfig_availabilityZoneID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "gi_minor_versions.0.grid_image_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "gi_minor_versions.0.version"),
				),
			},
		},
	})
}

const testAccGIMinorVersionsConfig_basic = `
data "aws_odb_gi_minor_versions" "test" {
  gi_version = "19.0.0.0"
}
`

const testAccGIMinorVersionsConfig_availabilityZone = `
data "aws_availability_zone" "test" {
  zone_id = "use1-az6"
}

data "aws_odb_gi_minor_versions" "test" {
  availability_zone = data.aws_availability_zone.test.name
  gi_version        = "19.0.0.0"
  shape_family      = "EXADB_XS"
}
`

const testAccGIMinorVersionsConfig_availabilityZoneID = `
data "aws_odb_gi_minor_versions" "test" {
  availability_zone_id = "use1-az6"
  gi_version           = "19.0.0.0"
  shape_family         = "EXADB_XS"
}
`
