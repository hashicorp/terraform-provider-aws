// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppIntegrationsEventIntegrationDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_appintegrations_event_integration.test"
	datasourceName := "data.aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	sourceName := os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, appintegrationsservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationDataSourceConfig_Name(rName, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "event_filter.#", resourceName, "event_filter.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "event_filter.0.source", resourceName, "event_filter.0.source"),
					resource.TestCheckResourceAttrPair(datasourceName, "eventbridge_bus", resourceName, "eventbridge_bus"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
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

  tags = {
    "Key1" = "Value1"
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
