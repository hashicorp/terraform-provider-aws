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
	dataSource1Name := "data.aws_cloudwatch_event_buses.by_name_prefix"
	dataSource2Name := "data.aws_cloudwatch_event_buses.all"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusesDataSourceConfig_basic(busName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "event_buses.#", "1"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "event_buses.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSource1Name, "event_buses.0.creation_time"),
					resource.TestCheckResourceAttrSet(dataSource1Name, "event_buses.0.last_modified_time"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "event_buses.0.name", resourceName, names.AttrName),
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSource2Name, "event_buses.#", 1),
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

data "aws_cloudwatch_event_buses" "by_name_prefix" {
  name_prefix = aws_cloudwatch_event_bus.test.name
}

data "aws_cloudwatch_event_buses" "all" {
  depends_on = [aws_cloudwatch_event_bus.test]
}
`, busName)
}
