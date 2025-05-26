// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	"github.com/aws/aws-sdk-go-v2/service/notifications/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsNotificationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v notifications.GetNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_basic(rName, rDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("aggregation_duration"), tfknownvalue.StringExact(types.AggregationDurationNone)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("notifications", regexache.MustCompile(`configuration/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rDescription)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNotificationsNotificationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v notifications.GetNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_basic(rName, rDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotifications.ResourceNotificationConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccNotificationsNotificationConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v notifications.GetNotificationConfigurationOutput
	rNameOne := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameTwo := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescriptionOne := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescriptionTwo := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_all(rNameOne, rDescriptionOne, string(types.AggregationDurationShort)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("aggregation_duration"), tfknownvalue.StringExact(types.AggregationDurationShort)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("notifications", regexache.MustCompile(`configuration/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rDescriptionOne)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameOne)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				Config: testAccNotificationConfigurationConfig_all(rNameTwo, rDescriptionTwo, string(types.AggregationDurationLong)),
				Check:  resource.ComposeAggregateTestCheckFunc(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("aggregation_duration"), tfknownvalue.StringExact(types.AggregationDurationLong)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNRegexp("notifications", regexache.MustCompile(`configuration/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rDescriptionTwo)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameTwo)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccNotificationsNotificationConfiguration_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v notifications.GetNotificationConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rDescription := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_notifications_notification_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationConfigurationConfig_tags1(rName, rDescription, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccNotificationConfigurationConfig_tags2(rName, rDescription, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccNotificationConfigurationConfig_tags1(rName, rDescription, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationConfigurationExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckNotificationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_notification_configuration" {
				continue
			}

			_, err := tfnotifications.FindNotificationConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("User Notifications Notification Configuration %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckNotificationConfigurationExists(ctx context.Context, n string, v *notifications.GetNotificationConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		output, err := tfnotifications.FindNotificationConfigurationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccNotificationConfigurationConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccNotificationConfigurationConfig_all(rName, description, aggregation_duration string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name                 = %[1]q
  description          = %[2]q
  aggregation_duration = %[3]q
}
`, rName, description, aggregation_duration)
}

func testAccNotificationConfigurationConfig_tags1(rName, description, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, description, tag1Key, tag1Value)
}

func testAccNotificationConfigurationConfig_tags2(rName, description, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, description, tag1Key, tag1Value, tag2Key, tag2Value)
}
