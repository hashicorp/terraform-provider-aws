// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccStandardsControlAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_standards_control_associations.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlAssociationsDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "standards_control_associations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "standards_control_associations.*", map[string]string{
						"association_status":  "ENABLED",
						"security_control_id": "IAM.1",
					}),
				),
			},
		},
	})
}

const testAccStandardsControlAssociationsDataSourceConfig_basic = `
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
  enable_default_standards = false
}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"

  depends_on = [aws_securityhub_account.test]
}

data "aws_securityhub_standards_control_associations" "test" {
  security_control_id = "IAM.1"

  depends_on = [aws_securityhub_standards_subscription.test]
}
`
