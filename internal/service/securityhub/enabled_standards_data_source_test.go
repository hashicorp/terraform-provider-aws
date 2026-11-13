// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnabledStandardsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_enabled_standards.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnabledStandardsDataSourceConfig_basic,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("standards_subscriptions"), tfknownvalue.ListNotEmpty()),
				},
			},
		},
	})
}

func testAccEnabledStandardsDataSource_standardsSubscriptionARN(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_enabled_standards.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnabledStandardsDataSourceConfig_standardsSubscriptionARN,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("standards_subscriptions"), knownvalue.ListSizeExact(1)),
				},
			},
		},
	})
}

const testAccEnabledStandardsDataSourceConfig_basic = `
resource "aws_securityhub_account" "test" {}

data "aws_securityhub_enabled_standards" "test" {
  depends_on = [aws_securityhub_account.test]
}
`

const testAccEnabledStandardsDataSourceConfig_standardsSubscriptionARN = `
resource "aws_securityhub_account" "test" {}

data "aws_securityhub_enabled_standards" "all" {
  depends_on = [aws_securityhub_account.test]
}

data "aws_securityhub_enabled_standards" "test" {
  standards_subscription_arns = [data.aws_securityhub_enabled_standards.all.standards_subscriptions[0].standards_subscription_arn]
}
`
