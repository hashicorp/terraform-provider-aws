// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSKeyLastUsageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_kms_key_last_usage.test"
	resourceName := "aws_kms_key.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyLastUsageDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKeyID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "key_creation_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tracking_start_date"),
				),
			},
		},
	})
}

func TestAccKMSKeyLastUsageDataSource_keyARN(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_kms_key_last_usage.test"
	resourceName := "aws_kms_key.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyLastUsageDataSourceConfig_keyARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKeyID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "key_creation_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tracking_start_date"),
				),
			},
		},
	})
}

func testAccKeyLastUsageDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

data "aws_kms_key_last_usage" "test" {
  key_id = aws_kms_key.test.id
}
`, rName)
}

func testAccKeyLastUsageDataSourceConfig_keyARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

data "aws_kms_key_last_usage" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}
