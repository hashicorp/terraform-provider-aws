// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventBridgeV2EventBusesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_eventbridgev2_event_buses.test"
	resourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusesDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("event_buses"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrAccountID), tfknownvalue.AccountID()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrName), resourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrState), resourceName, tfjsonpath.New(names.AttrState), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccEventBridgeV2EventBusesDataSource_sharedEventBus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_eventbridgev2_event_buses.test"
	resourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventBusesDataSourceConfig_sharedEventBus(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("event_buses"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrAccountID), "data.aws_caller_identity.owner", tfjsonpath.New(names.AttrAccountID), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_buses").AtSliceIndex(0).AtMapKey(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccEventBusesDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEventBusConfig_basic(rName), `
data "aws_eventbridgev2_event_buses" "test" {
  name_prefix = aws_eventbridgev2_event_bus.test.name
}
`)
}

func testAccEventBusesDataSourceConfig_sharedEventBus(rName string) string {
	return acctest.ConfigCompose(testAccEventBusConfig_shared(rName), `
data "aws_caller_identity" "owner" {
  provider = "awsalternate"
}

data "aws_eventbridgev2_event_buses" "test" {
  event_bus_account_id = data.aws_caller_identity.owner.account_id
  name_prefix          = aws_eventbridgev2_event_bus.test.name

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`)
}
