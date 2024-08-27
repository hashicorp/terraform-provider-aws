// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDetectorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_guardduty_detector.test"
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:                acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories:  acctest.ProtoV5ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "features.#", 0),
					resource.TestCheckResourceAttrPair(datasourceName, "finding_publishing_frequency", resourceName, "finding_publishing_frequency"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(datasourceName, names.AttrServiceRoleARN, "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttr(datasourceName, names.AttrStatus, "ENABLED"),
				),
			},
		},
	})
}

func testAccDetectorDataSource_ID(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_guardduty_detector.test"
	resourceName := "aws_guardduty_detector.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDetectorNotExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GuardDutyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_id,
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(datasourceName, "features.#", 0),
					resource.TestCheckResourceAttrPair(datasourceName, "finding_publishing_frequency", resourceName, "finding_publishing_frequency"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrID, resourceName, names.AttrID),
					acctest.CheckResourceAttrGlobalARN(datasourceName, names.AttrServiceRoleARN, "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttr(datasourceName, names.AttrStatus, "ENABLED"),
				),
			},
		},
	})
}

const testAccDetectorDataSourceConfig_basic = `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {
  depends_on = [aws_guardduty_detector.test]
}
`

const testAccDetectorDataSourceConfig_id = `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {
  id = aws_guardduty_detector.test.id
}
`
