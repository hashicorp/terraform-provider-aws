// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDataLakeSettingsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeSettingsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "catalog_id", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "admins.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", "data.aws_iam_session_context.current", "issuer_arn"),
					resource.TestCheckResourceAttr(resourceName, "allow_external_data_filtering", "false"),
					resource.TestCheckResourceAttr(resourceName, "external_data_filtering_allow_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "authorized_session_tag_value_list.#", "0"),
				),
			},
		},
	})
}

const testAccDataLakeSettingsDataSourceConfig_basic = `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = data.aws_caller_identity.current.account_id
  admins     = [data.aws_iam_session_context.current.issuer_arn]
}

data "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = aws_lakeformation_data_lake_settings.test.catalog_id
}
`
