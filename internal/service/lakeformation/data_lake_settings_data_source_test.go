// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDataLakeSettingsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeSettingsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCatalogID, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "admins.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "admins.0", "data.aws_iam_session_context.current", "issuer_arn"),
					resource.TestCheckResourceAttr(resourceName, "allow_external_data_filtering", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "external_data_filtering_allow_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "authorized_session_tag_value_list.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "allow_full_table_external_data_access", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccDataLakeSettingsDataSource_readOnlyAdmins(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "data.aws_lakeformation_data_lake_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.LakeFormation) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataLakeSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataLakeSettingsDataSourceConfig_readOnlyAdmins,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCatalogID, "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "read_only_admins.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "read_only_admins.0", "data.aws_iam_session_context.current", "issuer_arn"),
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

const testAccDataLakeSettingsDataSourceConfig_readOnlyAdmins = `
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  catalog_id       = data.aws_caller_identity.current.account_id
  read_only_admins = [data.aws_iam_session_context.current.issuer_arn]
}

data "aws_lakeformation_data_lake_settings" "test" {
  catalog_id = aws_lakeformation_data_lake_settings.test.catalog_id
}
`
