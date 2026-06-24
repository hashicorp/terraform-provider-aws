// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRuleDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_cloudwatch_event_rule.test"
	resourceName := "aws_cloudwatch_event_rule.test"

	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_bus_name", resourceName, "event_bus_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrScheduleExpression, resourceName, names.AttrScheduleExpression),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_pattern", resourceName, "event_pattern"),
				),
			},
		},
	})
}

func TestAccRuleDataSource_eventBus(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_cloudwatch_event_rule.test"
	resourceName := "aws_cloudwatch_event_rule.test"

	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	eventBusName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_eventBus(name, eventBusName, "test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_bus_name", resourceName, "event_bus_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrScheduleExpression, resourceName, names.AttrScheduleExpression),
					resource.TestCheckResourceAttrPair(dataSourceName, "event_pattern", resourceName, "event_pattern"),
				),
			},
		},
	})
}

func testAccRuleDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRuleConfig_basic(rName),
		`
data "aws_cloudwatch_event_rule" "test" {
  name = aws_cloudwatch_event_rule.test.name
}
`,
	)
}

func testAccRuleDataSourceConfig_eventBus(rName, eventBusName, description string) string {
	return acctest.ConfigCompose(
		testAccRuleConfig_busName(rName, eventBusName, description),
		`
data "aws_cloudwatch_event_rule" "test" {
  name = aws_cloudwatch_event_rule.test.name
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
`,
	)
}
