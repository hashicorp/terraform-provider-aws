// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventBridgeV2EventBusDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_eventbridgev2_event_bus.test"
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
				Config: testAccEventBusDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrCreationTime), resourceName, tfjsonpath.New(names.AttrCreationTime), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrDescription), resourceName, tfjsonpath.New(names.AttrDescription), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrName), resourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrState), resourceName, tfjsonpath.New(names.AttrState), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("storage_configuration"), resourceName, tfjsonpath.New("storage_configuration"), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccEventBridgeV2EventBusDataSource_sharedEventBus(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_eventbridgev2_event_bus.test"
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
				Config: testAccEventBusDataSourceConfig_sharedEventBus(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrName), resourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrState), resourceName, tfjsonpath.New(names.AttrState), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccEventBusDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEventBusConfig_basic(rName), `
data "aws_eventbridgev2_event_bus" "test" {
  arn = aws_eventbridgev2_event_bus.test.arn
}
`)
}

func testAccEventBusDataSourceConfig_sharedEventBus(rName string) string {
	return acctest.ConfigCompose(testAccEventBusConfig_shared(rName), `
data "aws_caller_identity" "owner" {
  provider = "awsalternate"
}

# The consumer account finds the shared event bus by its owner account and name, then describes it by ARN.
data "aws_eventbridgev2_event_buses" "test" {
  event_bus_account_id = data.aws_caller_identity.owner.account_id
  name_prefix          = aws_eventbridgev2_event_bus.test.name

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}

data "aws_eventbridgev2_event_bus" "test" {
  arn = data.aws_eventbridgev2_event_buses.test.event_buses[0].arn
}
`)
}

// testAccEventBusConfig_shared creates an event bus in the alternate account and shares it with the primary account
// through AWS RAM. Both accounts must be in the same AWS Organization with RAM sharing with AWS Organizations enabled.
func testAccEventBusConfig_shared(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

# The event bus is owned by the alternate account and shared with the primary account within the AWS Organization,
# so no resource share invitation needs to be accepted.
resource "aws_eventbridgev2_event_bus" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name                      = %[1]q
  allow_external_principals = false
  permission_arns           = ["arn:${data.aws_partition.current.partition}:ram::aws:permission/AWSRAMEventBridgeEventBusV2FullAccess"]
}

resource "aws_ram_resource_association" "test" {
  provider = "awsalternate"

  resource_arn       = aws_eventbridgev2_event_bus.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_caller_identity.current.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}
`, rName))
}
