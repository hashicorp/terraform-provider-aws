// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsEventBusesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_bus.test"
	dataSourceName := "data.aws_cloudwatch_event_buses.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusesDataSourceConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "event_buses.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_buses.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "event_buses.0.creation_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "event_buses.0.last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_buses.0.name", resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccEventBusesDataSourceConfig_basic(busName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_cloudwatch_event_buses" "test" {
  depends_on = [aws_cloudwatch_event_bus.test]

  name_prefix = %[1]q
}
`, busName)
}
