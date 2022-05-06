package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func TestAccGlueTrigger_basic(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_OnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", "0"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("trigger/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "workflow_name", ""),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_crawler(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test_trigger"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_Crawler(rName, "SUCCEEDED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.crawler_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawler_name", fmt.Sprintf("%scrawl2", rName)),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawl_state", "SUCCEEDED"),
					resource.TestCheckResourceAttr(resourceName, "type", "CONDITIONAL"),
				),
			},
			{
				Config: testAccTriggerConfig_Crawler(rName, "FAILED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.crawler_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawler_name", fmt.Sprintf("%scrawl2", rName)),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.crawl_state", "FAILED"),
					resource.TestCheckResourceAttr(resourceName, "type", "CONDITIONAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_description(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccTriggerConfig_Description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_enabled(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccTriggerConfig_Enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				Config: testAccTriggerConfig_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_predicate(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_Predicate(rName, "SUCCEEDED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.state", "SUCCEEDED"),
					resource.TestCheckResourceAttr(resourceName, "type", "CONDITIONAL"),
				),
			},
			{
				Config: testAccTriggerConfig_Predicate(rName, "FAILED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "predicate.0.conditions.0.state", "FAILED"),
					resource.TestCheckResourceAttr(resourceName, "type", "CONDITIONAL"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_schedule(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_Schedule(rName, "cron(1 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(1 2 * * ? *)"),
				),
			},
			{
				Config: testAccTriggerConfig_Schedule(rName, "cron(2 3 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(2 3 * * ? *)"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_startOnCreate(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_ScheduleStart(rName, "cron(1 2 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(1 2 * * ? *)"),
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
	var trigger1, trigger2, trigger3 glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
			{
				Config: testAccTriggerTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTriggerTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_workflowName(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_WorkflowName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "workflow_name", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_Actions_notify(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerActionsNotificationConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
			{
				Config: testAccTriggerActionsNotificationConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", "2"),
				),
			},
			{
				Config: testAccTriggerActionsNotificationConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.notification_property.0.notify_delay_after", "1"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_Actions_security(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerActionsSecurityConfigurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "actions.0.job_name", rName),
					resource.TestCheckResourceAttr(resourceName, "actions.0.security_configuration", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
		},
	})
}

func TestAccGlueTrigger_onDemandDisable(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_OnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", "ON_DEMAND"),
				),
			},
			{
				Config: testAccTriggerConfig_OnDemandEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "type", "ON_DEMAND"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled"},
			},
			{
				Config: testAccTriggerConfig_OnDemandEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "type", "ON_DEMAND"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_eventBatchingCondition(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfigEvent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_window", "900"),
					resource.TestCheckResourceAttr(resourceName, "type", "EVENT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "start_on_creation"},
			},
			{
				Config: testAccTriggerConfigEventUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_batching_condition.0.batch_window", "50"),
					resource.TestCheckResourceAttr(resourceName, "type", "EVENT"),
				),
			},
		},
	})
}

func TestAccGlueTrigger_disappears(t *testing.T) {
	var trigger glue.Trigger

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_trigger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTriggerConfig_OnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTriggerExists(resourceName, &trigger),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceTrigger(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTriggerExists(resourceName string, trigger *glue.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Trigger ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := tfglue.FindTriggerByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output.Trigger == nil {
			return fmt.Errorf("Glue Trigger (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Trigger.Name) == rs.Primary.ID {
			*trigger = *output.Trigger
			return nil
		}

		return fmt.Errorf("Glue Trigger (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckTriggerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_trigger" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := tfglue.FindTriggerByName(conn, rs.Primary.ID)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}

		}

		trigger := output.Trigger
		if trigger != nil && aws.StringValue(trigger.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Trigger %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccTriggerConfig_Description(rName, description string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_Enabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_OnDemand(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
resource "aws_glue_trigger" "test" {
  name = %[1]q
  type = "ON_DEMAND"

  actions {
    job_name = aws_glue_job.test.name
  }
}
`, rName))
}

func testAccTriggerConfig_OnDemandEnabled(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_Predicate(rName, state string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_Crawler(rName, state string) string {
	return acctest.ConfigCompose(testAccCrawlerConfig_s3Target(rName, "s3://test_bucket"), fmt.Sprintf(`
resource "aws_glue_crawler" "test2" {
  depends_on = [aws_iam_role_policy_attachment.test-AWSGlueServiceRole]

  database_name = aws_glue_catalog_database.test.name
  name          = "%[1]scrawl2"
  role          = aws_iam_role.test.name

  s3_target {
    path = "s3://test_bucket"
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

func testAccTriggerConfig_Schedule(rName, schedule string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_ScheduleStart(rName, schedule string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfig_WorkflowName(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerActionsNotificationConfig(rName string, delay int) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerActionsSecurityConfigurationConfig(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfigEvent(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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

func testAccTriggerConfigEventUpdated(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_Required(rName), fmt.Sprintf(`
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
