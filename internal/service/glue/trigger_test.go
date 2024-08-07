// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueTrigger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_onDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", acctest.Ct0),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("trigger/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "workflow_name", ""),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_crawler(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test_trigger"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_crawler(rName, "SUCCEEDED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.crawler_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawler_name", fmt.Sprintf("%scrawl2", rName)),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawl_state", "SUCCEEDED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CONDITIONAL"),
				),
			},
			{
				Config: testAccTriggerConfig_crawler(rName, "FAILED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.crawler_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawler_name", fmt.Sprintf("%scrawl2", rName)),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawl_state", "FAILED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CONDITIONAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_description(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				Config: testAccTriggerConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccTriggerConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				Config: testAccTriggerConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled, names.AttrState}, // adding state to igonre list because trigger state changes faster before test can verify what is in TF state
			},
		},
	})
}

func TestAccGlueTrigger_predicate(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_predicate(rName, "SUCCEEDED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.state", "SUCCEEDED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CONDITIONAL"),
				),
			},
			{
				Config: testAccTriggerConfig_predicate(rName, "FAILED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.state", "FAILED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CONDITIONAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_schedule(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_schedule(rName, "cron(1 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(1 2 * * ? *)"),
				),
			},
			{
				Config: testAccTriggerConfig_schedule(rName, "cron(2 3 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(2 3 * * ? *)"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_startOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_scheduleStart(rName, "cron(1 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, "cron(1 2 * * ? *)"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_on_creation"},
			},
		},
	})
}

func TestAccGlueTrigger_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger1, trigger2, trigger3 awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
			{
				Config: testAccTriggerConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTriggerConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueTrigger_workflowName(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_workflowName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "workflow_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_Actions_notify(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_actionsNotification(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
			{
				Config: testAccTriggerConfig_actionsNotification(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", acctest.Ct2),
				),
			},
			{
				Config: testAccTriggerConfig_actionsNotification(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccGlueTrigger_Actions_security(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_actionsSecurityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.security_configuration", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
		},
	})
}

func TestAccGlueTrigger_onDemandDisable(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_onDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ON_DEMAND"),
				),
			},
			{
				Config: testAccTriggerConfig_onDemandEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ON_DEMAND"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled},
			},
			{
				Config: testAccTriggerConfig_onDemandEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ON_DEMAND"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_eventBatchingCondition(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_event(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_window", "900"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EVENT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEnabled, "start_on_creation"},
			},
			{
				Config: testAccTriggerConfig_eventUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_size", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_window", "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EVENT"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var trigger awstypes.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTriggerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_onDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(ctx, resourceName, &trigger),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceTrigger(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTriggerExists(ctx context.Context, resourceName string, trigger *awstypes.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Trigger ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := tfglue.FindTriggerByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output.Trigger == nil {
			return fmt.Errorf("Glue Trigger (%s) not found", rs.Primary.ID)
		}

		if aws.ToString(output.Trigger.Name) == rs.Primary.ID {
			*trigger = *output.Trigger
			return nil
		}

		return fmt.Errorf("Glue Trigger (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckTriggerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_trigger" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			output, err := tfglue.FindTriggerByName(ctx, conn, rs.Primary.ID)

			if err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					return nil
				}
			}

			trigger := output.Trigger
			if trigger != nil && aws.ToString(trigger.Name) == rs.Primary.ID {
				return fmt.Errorf("Glue Trigger %s still exists", rs.Primary.ID)
			}

			return err
		}

		return nil
	}
}

func testAccTriggerConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  description = %[1]q
  name        = %[2]q
  type        = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, description, rName))
}

func testAccTriggerConfig_enabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  enabled  = %[1]t
  name     = %[2]q
  schedule = "cron(15 12 * * ? *)"
  type     = "SCHEDULED"

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, enabled, rName))
}

func testAccTriggerConfig_onDemand(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName))
}

func testAccTriggerConfig_onDemandEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name    = %[1]q
  type    = "ON_DEMAND"
  enabled = %[2]t

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName, enabled))
}

func testAccTriggerConfig_predicate(rName, state string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_job" "test2" {
  name     = "%[1]s2"
  role_arn = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "CONDITIONAL"

  actions {
    job_name = aws_glue_job.test2.name
  }

  predicate {
    conditions {
      job_name = aws_glue_job.test.name
      state    = %[2]q
    }
  }
}
`, rName, state))
}

func testAccTriggerConfig_crawler(rName, state string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_s3Target(rName, "bucket1"), fmt.Sprintf(`
resource "aws_s3_bucket" "test2" {
  bucket = %[1]q
}

resource "aws_glue_crawler" "test2" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = "%[1]scrawl2"
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://${aws_s3_bucket.test2.bucket}"
  }
}

resource "aws_glue_trigger" "test_trigger" {
  name = %[1]q
  type = "CONDITIONAL"

  actions {
    crawler_name = aws_glue_crawler.test.name
  }

  predicate {
    conditions {
      crawler_name = aws_glue_crawler.test2.name
      crawl_state  = %[2]q
    }
  }
}
`, rName, state))
}

func testAccTriggerConfig_schedule(rName, schedule string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name     = %[1]q
  schedule = %[2]q
  type     = "SCHEDULED"

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName, schedule))
}

func testAccTriggerConfig_scheduleStart(rName, schedule string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name              = %[1]q
  schedule          = %[2]q
  type              = "SCHEDULED"
  start_on_creation = true

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName, schedule))
}

func testAccTriggerConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccTriggerConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTriggerConfig_workflowName(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_workflow" test {
  name = %[1]q
}

resource "aws_glue_trigger" "test" {
  name          = %[1]q
  type          = "ON_DEMAND"
  workflow_name = aws_glue_workflow.test.name

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName))
}

func testAccTriggerConfig_actionsNotification(rName string, delay int) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name

    notification_property {
      notify_delay_after = %[2]d
    }
  }
}
`, rName, delay))
}

func testAccTriggerConfig_actionsSecurityConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_security_configuration" "test" {
  name = %[1]q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}

resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name               = aws_glue_job.test.name
    security_configuration = aws_glue_security_configuration.test.name
  }
}
`, rName))
}

func testAccTriggerConfig_event(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_workflow" test {
  name = %[1]q
}

resource "aws_glue_trigger" "test" {
  name              = %[1]q
  type              = "EVENT"
  workflow_name     = aws_glue_workflow.test.name
  start_on_creation = false

  actions {
    job_name = aws_glue_job.test.name
  }

  event_batching_condition {
    batch_size = 1
  }
}
`, rName))
}

func testAccTriggerConfig_eventUpdated(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_required(rName), fmt.Sprintf(`
resource "aws_glue_workflow" test {
  name = %[1]q
}

resource "aws_glue_trigger" "test" {
  name              = %[1]q
  type              = "EVENT"
  workflow_name     = aws_glue_workflow.test.name
  start_on_creation = false

  actions {
    job_name = aws_glue_job.test.name
  }

  event_batching_condition {
    batch_size   = 1
    batch_window = 50
  }
}
`, rName))
}
