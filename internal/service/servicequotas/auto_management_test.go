// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicequotas "github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceQuotasAutoManagement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var automanagement servicequotas.GetAutoManagementConfigurationOutput
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
				Config: testAccAutoManagementConfig_basic("NotifyOnly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttr(resourceName, "opt_in_type", "NotifyOnly"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), tfknownvalue.AccountID()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoManagementConfig_basic("NotifyAndAdjust"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttr(resourceName, "opt_in_type", "NotifyAndAdjust"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), tfknownvalue.AccountID()),
				},
			},
		},
	})
}

func TestAccServiceQuotasAutoManagement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var automanagement servicequotas.GetAutoManagementConfigurationOutput
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
				Config: testAccAutoManagementConfig_basic("NotifyOnly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicequotas.ResourceAutoManagement, resourceName),
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

func TestAccServiceQuotasAutoManagement_updateExclusionList(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var automanagement servicequotas.GetAutoManagementConfigurationOutput
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
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.dynamodb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.dynamodb.0", quotaCode),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoManagementConfig_exclusionList(quotaCodeUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.dynamodb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "exclusion_list.dynamodb.0", quotaCodeUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoManagementConfig_basic("NotifyOnly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckNoResourceAttr(resourceName, "exclusion_list.%"),
				),
			},
		},
	})
}

func TestAccServiceQuotasAutoManagement_updateNotificationARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var automanagement servicequotas.GetAutoManagementConfigurationOutput
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
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttrPair(resourceName, "notification_arn", notificationResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoManagementConfig_notificationARN(rName, notificationResourceName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckResourceAttrPair(resourceName, "notification_arn", notificationResourceName2, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoManagementConfig_basic("NotifyOnly"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAutoManagementExists(ctx, t, resourceName, &automanagement),
					resource.TestCheckNoResourceAttr(resourceName, "notification_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
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

			_, err := tfservicequotas.GetAutoManagementConfiguration(ctx, conn)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameAutoManagement, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameAutoManagement, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAutoManagementExists(ctx context.Context, t *testing.T, name string, automanagement *servicequotas.GetAutoManagementConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameAutoManagement, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameAutoManagement, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

		resp, err := tfservicequotas.GetAutoManagementConfiguration(ctx, conn)
		if err != nil {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameAutoManagement, "", err)
		}

		*automanagement = *resp

		return nil
	}
}

func testAccAutoManagementConfig_basic(optInType string) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_auto_management" "test" {
  opt_in_type = %[1]q
}
`, optInType)
}

func testAccAutoManagementConfig_exclusionList(quotaCode string) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_auto_management" "test" {
  opt_in_type = "NotifyOnly"

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
  opt_in_type = "NotifyOnly"

  notification_arn = %[2]s.arn
}
`, rName, notificationResourceName)
}
