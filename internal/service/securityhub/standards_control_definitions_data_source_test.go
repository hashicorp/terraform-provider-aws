// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSecurityHubStandardsControlDefinitionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_standards_control_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityHubServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlDefinitionsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "control_definitions.#"),
				),
			},
		},
	})
}

func TestAccSecurityHubStandardsControlDefinitionsDataSource_standardsArn(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_standards_control_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityHubServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlDefinitionsDataSourceConfig_standardsArn(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "control_definitions.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "standards_arn"),
				),
			},
		},
	})
}

func TestAccSecurityHubStandardsControlDefinitionsDataSource_currentRegionAvailability(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_standards_control_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityHubServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlDefinitionsDataSourceConfig_currentRegionAvailability(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "control_definitions.#"),
					resource.TestCheckResourceAttr(dataSourceName, "current_region_availability", "AVAILABLE"),
				),
			},
		},
	})
}

func TestAccSecurityHubStandardsControlDefinitionsDataSource_severityRating(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_securityhub_standards_control_definitions.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityHubServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlDefinitionsDataSourceConfig_severityRating(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "control_definitions.#"),
					resource.TestCheckResourceAttr(dataSourceName, "severity_rating", "CRITICAL"),
				),
			},
		},
	})
}

func testAccStandardsControlDefinitionsDataSourceConfig_basic() string {
	return `
data "aws_securityhub_standards_control_definitions" "test" {}
`
}

func testAccStandardsControlDefinitionsDataSourceConfig_standardsArn() string {
	return `
data "aws_securityhub_standards_subscriptions" "example" {}

data "aws_securityhub_standards_control_definitions" "test" {
  standards_arn = data.aws_securityhub_standards_subscriptions.example.standards_subscriptions[0].standards_arn
}
`
}

func testAccStandardsControlDefinitionsDataSourceConfig_currentRegionAvailability() string {
	return `
data "aws_securityhub_standards_control_definitions" "test" {
  current_region_availability = "AVAILABLE"
}
`
}

func testAccStandardsControlDefinitionsDataSourceConfig_severityRating() string {
	return `
data "aws_securityhub_standards_control_definitions" "test" {
  severity_rating = "CRITICAL"
}
`
}
