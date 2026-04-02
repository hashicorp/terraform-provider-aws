// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicequotas "github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceQuotasAutoManagement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_auto_management.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoManagementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoManagementConfig_basic(awstypes.OptInTypeNotifyOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclusion_list"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_in_type"), tfknownvalue.StringExact(awstypes.OptInTypeNotifyOnly)),
				},
			},
			{
				Config: testAccAutoManagementConfig_basic(awstypes.OptInTypeNotifyAndAdjust),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclusion_list"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_in_type"), tfknownvalue.StringExact(awstypes.OptInTypeNotifyAndAdjust)),
				},
			},
		},
	})
}

func TestAccServiceQuotasAutoManagement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_auto_management.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoManagementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoManagementConfig_basic(awstypes.OptInTypeNotifyOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicequotas.ResourceAutoManagement, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccServiceQuotasAutoManagement_updateExclusionList(t *testing.T) {
	ctx := acctest.Context(t)
	quotaCode := "L-F7858A77"
	quotaCodeUpdated := "L-F98FE922"
	resourceName := "aws_servicequotas_auto_management.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoManagementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoManagementConfig_exclusionList(quotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclusion_list"), knownvalue.MapExact(map[string]knownvalue.Check{
						"dynamodb": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(quotaCode)}),
					})),
				},
			},
			{
				Config: testAccAutoManagementConfig_exclusionList(quotaCodeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclusion_list"), knownvalue.MapExact(map[string]knownvalue.Check{
						"dynamodb": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(quotaCodeUpdated)}),
					})),
				},
			},
			{
				Config: testAccAutoManagementConfig_basic("NotifyOnly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("exclusion_list"), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccServiceQuotasAutoManagement_updateNotificationARN(t *testing.T) {
	ctx := acctest.Context(t)
	notificationResourceName1 := "aws_notifications_notification_configuration.test_1"
	notificationResourceName2 := "aws_notifications_notification_configuration.test_2"
	resourceName := "aws_servicequotas_auto_management.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAutoManagementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAutoManagementConfig_notificationARN(rName, notificationResourceName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("notification_arn"), notificationResourceName1, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccAutoManagementConfig_notificationARN(rName, notificationResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("notification_arn"), notificationResourceName2, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccAutoManagementConfig_basic(awstypes.OptInTypeNotifyOnly),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("notification_arn"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckAutoManagementDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicequotas_auto_management" {
				continue
			}

			_, err := tfservicequotas.FindAutoManagement(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("Service Quotas Auto Management still exists")
		}

		return nil
	}
}

func testAccCheckAutoManagementExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

		_, err := tfservicequotas.FindAutoManagement(ctx, conn)

		return err
	}
}

func testAccAutoManagementConfig_basic(optInType awstypes.OptInType) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_auto_management" "test" {
  opt_in_level = "ACCOUNT"
  opt_in_type  = %[1]q
}
`, optInType)
}

func testAccAutoManagementConfig_exclusionList(quotaCode string) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_auto_management" "test" {
  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"

  exclusion_list = {
    "dynamodb" = [
      %[1]q
    ]
  }
}
`, quotaCode)
}

func testAccAutoManagementConfig_notificationARN(rName, notificationResourceName string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test_1" {
  name        = "%[1]s-1"
  description = "example"
}

resource "aws_notifications_notification_configuration" "test_2" {
  name        = "%[1]s-2"
  description = "example"
}

resource "aws_servicequotas_auto_management" "test" {
  opt_in_level = "ACCOUNT"
  opt_in_type  = "NotifyOnly"

  notification_arn = %[2]s.arn
}
`, rName, notificationResourceName)
}
