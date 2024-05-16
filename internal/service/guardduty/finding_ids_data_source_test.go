// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFindingIDsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_guardduty_finding_ids.test"
	detectorDataSourceName := "data.aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFindingIDsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "detector_id", detectorDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "has_findings"),
					resource.TestCheckResourceAttrSet(dataSourceName, "finding_ids.#"),
				),
			},
		},
	})
}

func testAccFindingIDsDataSourceConfig_basic() string {
	return `
data "aws_guardduty_detector" "test" {}

data "aws_guardduty_finding_ids" "test" {
  detector_id = data.aws_guardduty_detector.test.id
}
`
}
