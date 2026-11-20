// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2ConfigurationSetDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_pinpointsmsvoicev2_configuration_set.test"
	resourceName := "aws_pinpointsmsvoicev2_configuration_set.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrARN), resourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("created_timestamp"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("default_message_type"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("default_sender_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("event_destinations"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrName), resourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetDataSource_eventDestinations(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_pinpointsmsvoicev2_configuration_set.test"
	topicResourceName := "aws_sns_topic.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationSetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationSetDataSourceConfig_eventDestinations(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("event_destinations"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"cloudwatch_logs_destination":  knownvalue.Null(),
							names.AttrEnabled:              knownvalue.Bool(true),
							"event_destination_name":       knownvalue.StringExact(rName),
							"kinesis_firehose_destination": knownvalue.Null(),
							"matching_event_types": knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("TEXT_DELIVERED"),
							}),
						}),
					})),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("event_destinations").AtSliceIndex(0).AtMapKey("sns_destination").AtSliceIndex(0).AtMapKey(names.AttrTopicARN), topicResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2ConfigurationSetDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckConfigurationSet(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccConfigurationSetDataSourceConfig_nonExistent(rName),
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccConfigurationSetDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = aws_pinpointsmsvoicev2_configuration_set.test.name
}

resource "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, acctest.CtKey1, acctest.CtValue1)
}

func testAccConfigurationSetDataSourceConfig_eventDestinations(rName string) string {
	return acctest.ConfigCompose(testAccEventDestinationConfig_basic(rName), `
data "aws_pinpointsmsvoicev2_configuration_set" "test" {
  # Reference the event destination so the data source is read after it exists.
  name = aws_pinpointsmsvoicev2_event_destination.test.configuration_set_name
}
`)
}

func testAccConfigurationSetDataSourceConfig_nonExistent(rName string) string {
	return fmt.Sprintf(`
data "aws_pinpointsmsvoicev2_configuration_set" "test" {
  name = %[1]q
}
`, rName)
}
