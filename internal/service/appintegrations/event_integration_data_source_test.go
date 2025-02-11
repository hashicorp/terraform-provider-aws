// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations_test

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppIntegrationsEventIntegrationDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_appintegrations_event_integration.test"
	dataSourceName := "data.aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	sourceName := os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppIntegrationsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppIntegrationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationDataSourceConfig_Name(rName, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_filter.#", resourceName, "event_filter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_filter.0.source", resourceName, "event_filter.0.source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "eventbridge_bus", resourceName, "eventbridge_bus"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccEventIntegrationDataSourceConfig_base(rName, sourceName string) string {
	return fmt.Sprintf(`
resource "aws_appintegrations_event_integration" "test" {
  name            = %[1]q
  description     = "example description"
  eventbridge_bus = "default"

  event_filter {
    source = %[2]q
  }
}
`, rName, sourceName)
}

func testAccEventIntegrationDataSourceConfig_Name(rName, sourceName string) string {
	return acctest.ConfigCompose(testAccEventIntegrationDataSourceConfig_base(rName, sourceName), `
data "aws_appintegrations_event_integration" "test" {
  name = aws_appintegrations_event_integration.test.name
}
`)
}
