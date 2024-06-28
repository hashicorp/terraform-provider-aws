// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfscheduler "github.com/hashicorp/terraform-provider-aws/internal/service/scheduler"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestResourceScheduleIDFromARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		ARN   string
		ID    string
		Fails bool
	}{
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test", //lintignore:AWSAT003,AWSAT005
			ID:    "default/test",
			Fails: false,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/test/test", //lintignore:AWSAT003,AWSAT005
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default/", //lintignore:AWSAT003,AWSAT005
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//test", //lintignore:AWSAT003,AWSAT005
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule//", //lintignore:AWSAT003,AWSAT005
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "arn:aws:scheduler:eu-west-1:735669964269:schedule/default", //lintignore:AWSAT003,AWSAT005
			ID:    "",
			Fails: true,
		},
		{
			ARN:   "",
			ID:    "",
			Fails: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ARN, func(t *testing.T) {
			t.Parallel()

			id, err := tfscheduler.ResourceScheduleIDFromARN(tc.ARN)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if id != tc.ID {
				t.Errorf("expected id %s, got: %s", tc.ID, id)
			}
		})
	}
}

func TestResourceScheduleParseID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		ID           string
		GroupName    string
		ScheduleName string
		Fails        bool
	}{
		{
			ID:           "default/test",
			GroupName:    "default",
			ScheduleName: "test",
			Fails:        false,
		},
		{
			ID:           "default/test/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/test",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "/",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "//",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "default",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
		{
			ID:           "",
			GroupName:    "",
			ScheduleName: "",
			Fails:        true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			t.Parallel()

			groupName, scheduleName, err := tfscheduler.ResourceScheduleParseID(tc.ID)

			if tc.Fails {
				if err == nil {
					t.Errorf("expected an error")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %s", err)
				}
			}

			if groupName != tc.GroupName {
				t.Errorf("expected group name %s, got: %s", tc.GroupName, groupName)
			}

			if scheduleName != tc.ScheduleName {
				t.Errorf("expected schedule name %s, got: %s", tc.ScheduleName, scheduleName)
			}
		})
	}
}

func TestAccSchedulerSchedule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "scheduler", regexache.MustCompile(regexp.QuoteMeta(`schedule/default/`+name))),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, names.AttrGroupName, "default"),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, fmt.Sprintf("default/%s", name)),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.dead_letter_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.input", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "185"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfscheduler.ResourceSchedule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_description(name, "test 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_description(name, "test 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test 2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_description(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_endDate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_endDate(name, "2100-01-01T01:02:03Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2100-01-01T01:02:03Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_endDate(name, "2099-01-01T01:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", "2099-01-01T01:00:00Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_flexibleTimeWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_flexibleTimeWindow(name, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "FLEXIBLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_flexibleTimeWindow(name, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", "20"),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "FLEXIBLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.maximum_window_in_minutes", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "flexible_time_window.0.mode", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_groupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_groupName(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, "aws_scheduler_schedule_group.test", names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_kmsKeyARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_kmsKeyARN(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_kmsKeyARN(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test.1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_scheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 hour)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 hour)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_scheduleExpression(name, "rate(1 day)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrScheduleExpression, "rate(1 day)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_scheduleExpressionTimezone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_scheduleExpressionTimezone(name, "Europe/Paris"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "Europe/Paris"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_scheduleExpressionTimezone(name, "Australia/Sydney"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "Australia/Sydney"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression_timezone", "UTC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_startDate(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_startDate(name, "2100-01-01T01:02:03Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2100-01-01T01:02:03Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_startDate(name, "2099-01-01T01:00:00Z"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", "2099-01-01T01:00:00Z"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_state(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_state(name, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_state(name, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetARN(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetARN(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.arn", "aws_sqs_queue.test.1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetDeadLetterConfig(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetDeadLetterConfig(name, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.dead_letter_config.0.arn", "aws_sqs_queue.dlq.0", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetDeadLetterConfig(name, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.dead_letter_config.0.arn", "aws_sqs_queue.dlq.1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.dead_letter_config.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetECSParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetECSParameters1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_ecs_managed_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_execute_command", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.group", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.launch_type", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.placement_constraints.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.placement_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.platform_version", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.propagate_tags", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.reference_id", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.task_count", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.ecs_parameters.0.task_definition_arn", "aws_ecs_task_definition.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetECSParameters2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct2,
						"capacity_provider": "test1",
						names.AttrWeight:    "50",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct0,
						"capacity_provider": "test2",
						names.AttrWeight:    "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_ecs_managed_tags", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_execute_command", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.group", "my-task-group"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.assign_public_ip", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.security_groups.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.security_groups.*", "sg-111111111"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.subnets.*", "subnet-11111111"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.placement_constraints.*", map[string]string{
						names.AttrType:       "memberOf",
						names.AttrExpression: "attribute:ecs.os-family in [LINUX]",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.placement_strategy.*", map[string]string{
						names.AttrType:  "binpack",
						names.AttrField: "cpu",
					}),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.platform_version", "LATEST"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.propagate_tags", "TASK_DEFINITION"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.reference_id", "test-ref-id"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.task_count", acctest.Ct3),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.ecs_parameters.0.task_definition_arn", "aws_ecs_task_definition.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetECSParameters3(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct3,
						"capacity_provider": "test3",
						names.AttrWeight:    "100",
					}),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_ecs_managed_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_execute_command", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.group", "my-task-group-2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.assign_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.security_groups.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.security_groups.*", "sg-111111112"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.security_groups.*", "sg-111111113"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.subnets.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.subnets.*", "subnet-11111112"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.0.subnets.*", "subnet-11111113"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.placement_constraints.*", map[string]string{
						names.AttrType: "distinctInstance",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target.0.ecs_parameters.0.placement_strategy.*", map[string]string{
						names.AttrType:  "spread",
						names.AttrField: "cpu",
					}),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.platform_version", "1.1.0"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.propagate_tags", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.reference_id", "test-ref-id-2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.Key1", "Value1updated"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.task_count", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.ecs_parameters.0.task_definition_arn", "aws_ecs_task_definition.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetECSParameters4(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.capacity_provider_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_ecs_managed_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.enable_execute_command", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.group", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.launch_type", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.network_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.placement_constraints.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.placement_strategy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.platform_version", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.propagate_tags", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.reference_id", ""),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.0.task_count", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.ecs_parameters.0.task_definition_arn", "aws_ecs_task_definition.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.ecs_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetEventBridgeParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	scheduleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	eventBusName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, "test-1", "tf.test.1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.detail_type", "test-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.source", "tf.test.1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, "test-2", "tf.test.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.detail_type", "test-2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.0.source", "tf.test.2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(scheduleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.eventbridge_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetInput(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"
	var queueUrl string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetInput(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrWith("aws_sqs_queue.test", names.AttrURL, func(value string) error {
						queueUrl = value
						return nil
					}),
					func(s *terraform.State) error {
						return acctest.CheckResourceAttrEquivalentJSON(
							resourceName,
							"target.0.input",
							fmt.Sprintf(`{"MessageBody": "test1", "QueueUrl": %q}`, queueUrl),
						)(s)
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetInput(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					func(s *terraform.State) error {
						return acctest.CheckResourceAttrEquivalentJSON(
							resourceName,
							"target.0.input",
							fmt.Sprintf(`{"MessageBody": "test2", "QueueUrl": %q}`, queueUrl),
						)(s)
					},
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetKinesisParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	scheduleName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	streamName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, "test-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.0.partition_key", "test-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, "test-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.0.partition_key", "test-2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(scheduleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.kinesis_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetRetryPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetRetryPolicy(name, 60, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetRetryPolicy(name, 61, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "61"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_event_age_in_seconds", "86400"),
					resource.TestCheckResourceAttr(resourceName, "target.0.retry_policy.0.maximum_retry_attempts", "185"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetRoleARN(name, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetRoleARN(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.role_arn", "aws_iam_role.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetSageMakerPipelineParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters1(name, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							names.AttrName:  acctest.CtKey1,
							names.AttrValue: acctest.CtValue1,
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters2(name, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							names.AttrName:  acctest.CtKey1,
							names.AttrValue: acctest.CtValue1Updated,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							names.AttrName:  acctest.CtKey2,
							names.AttrValue: acctest.CtValue2,
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSageMakerPipelineParameters1(name, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(
						resourceName, "target.0.sagemaker_pipeline_parameters.0.pipeline_parameter.*",
						map[string]string{
							names.AttrName:  acctest.CtKey2,
							names.AttrValue: acctest.CtValue2,
						}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sagemaker_pipeline_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSchedulerSchedule_targetSQSParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schedule scheduler.GetScheduleOutput
	name := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_scheduler_schedule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SchedulerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SchedulerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduleConfig_targetSQSParameters(name, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.0.message_group_id", "test1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_targetSQSParameters(name, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.0.message_group_id", "test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccScheduleConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScheduleExists(ctx, t, resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "target.0.sqs_parameters.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckScheduleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SchedulerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_scheduler_schedule" {
				continue
			}

			groupName, scheduleName, err := tfscheduler.ResourceScheduleParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfscheduler.FindScheduleByTwoPartKey(ctx, conn, groupName, scheduleName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("%s %s %s still exists", names.Scheduler, tfscheduler.ResNameSchedule, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScheduleExists(ctx context.Context, t *testing.T, name string, v *scheduler.GetScheduleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Scheduler, create.ErrActionCheckingExistence, tfscheduler.ResNameSchedule, name, errors.New("not set"))
		}

		groupName, scheduleName, err := tfscheduler.ResourceScheduleParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).SchedulerClient(ctx)

		output, err := tfscheduler.FindScheduleByTwoPartKey(ctx, conn, groupName, scheduleName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccScheduleConfig_base = `
data "aws_caller_identity" "main" {}
data "aws_partition" "main" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" : data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}
`

func testAccScheduleConfig_basic(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name),
	)
}

func testAccScheduleConfig_description(name, description string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  description = %[2]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, description),
	)
}

func testAccScheduleConfig_endDate(name, endDate string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  end_date = %[2]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, endDate),
	)
}

func testAccScheduleConfig_flexibleTimeWindow(name string, window int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    maximum_window_in_minutes = %[2]d
    mode                      = "FLEXIBLE"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, window),
	)
}

func testAccScheduleConfig_groupName(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule_group" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  group_name = aws_scheduler_schedule_group.test.name

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name),
	)
}

func testAccScheduleConfig_kmsKeyARN(name string, index int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_kms_key" "test" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  kms_key_arn = aws_kms_key.test[%[2]d].arn

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, index),
	)
}

func testAccScheduleConfig_nameGenerated() string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`,
	)
}

func testAccScheduleConfig_namePrefix(namePrefix string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name_prefix = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, namePrefix),
	)
}

func testAccScheduleConfig_scheduleExpression(name, expression string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, expression),
	)
}

func testAccScheduleConfig_scheduleExpressionTimezone(name, timezone string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  schedule_expression_timezone = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, timezone),
	)
}

func testAccScheduleConfig_startDate(name, startDate string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  start_date = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, startDate),
	)
}

func testAccScheduleConfig_state(name, state string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  state = %[2]q

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, state),
	)
}

func testAccScheduleConfig_targetARN(name string, i int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test[%[2]d].arn
    role_arn = aws_iam_role.test.arn
  }
}
`, name, i),
	)
}

func testAccScheduleConfig_targetDeadLetterConfig(name string, index int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "dlq" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn

    dead_letter_config {
      arn = aws_sqs_queue.dlq[%[2]d].arn
    }
  }
}
`, name, index),
	)
}

func testAccScheduleConfig_targetECSParameters1(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_ecs_cluster.test.arn
    role_arn = aws_iam_role.test.arn

    ecs_parameters {
      task_definition_arn = aws_ecs_task_definition.test.arn
    }
  }
}
`, name),
	)
}

func testAccScheduleConfig_targetECSParameters2(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_ecs_cluster.test.arn
    role_arn = aws_iam_role.test.arn

    ecs_parameters {
      capacity_provider_strategy {
        base              = 2
        capacity_provider = "test1"
        weight            = 50
      }

      capacity_provider_strategy {
        base              = 0
        capacity_provider = "test2"
        weight            = 50
      }

      enable_ecs_managed_tags = true

      enable_execute_command = false

      group = "my-task-group"

      launch_type = "FARGATE"

      network_configuration {
        assign_public_ip = true
        security_groups  = ["sg-111111111"]
        subnets          = ["subnet-11111111"]
      }

      placement_constraints {
        type       = "memberOf"
        expression = "attribute:ecs.os-family in [LINUX]"
      }

      placement_strategy {
        type  = "binpack"
        field = "cpu"
      }

      platform_version = "LATEST"

      propagate_tags = "TASK_DEFINITION"

      reference_id = "test-ref-id"

      tags = {
        Key1 = "Value1"
        Key2 = "Value2"
      }

      task_count = 3

      task_definition_arn = aws_ecs_task_definition.test.arn
    }
  }
}
`, name),
	)
}

func testAccScheduleConfig_targetECSParameters3(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

locals {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = "${local.name}-2"
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_ecs_cluster.test.arn
    role_arn = aws_iam_role.test.arn

    ecs_parameters {
      capacity_provider_strategy {
        base              = 3
        capacity_provider = "test3"
        weight            = 100
      }

      enable_ecs_managed_tags = false

      enable_execute_command = true

      group = "my-task-group-2"

      launch_type = "FARGATE"

      network_configuration {
        assign_public_ip = false
        security_groups  = ["sg-111111112", "sg-111111113"]
        subnets          = ["subnet-11111112", "subnet-11111113"]
      }

      placement_constraints {
        type = "distinctInstance"
      }

      placement_strategy {
        type  = "spread"
        field = "cpu"
      }

      platform_version = "1.1.0"

      reference_id = "test-ref-id-2"

      tags = {
        Key1 = "Value1updated"
      }

      task_count = 1

      task_definition_arn = aws_ecs_task_definition.test.arn
    }
  }
}
`, name),
	)
}

func testAccScheduleConfig_targetECSParameters4(name string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

locals {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = "${local.name}-2"
  cpu                      = 256
  memory                   = 512
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_ecs_cluster.test.arn
    role_arn = aws_iam_role.test.arn

    ecs_parameters {
      launch_type         = "EC2"
      task_definition_arn = aws_ecs_task_definition.test.arn
    }
  }
}
`, name),
	)
}

func testAccScheduleConfig_targetEventBridgeParameters(scheduleName, eventBusName, detailType, source string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[2]q
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_cloudwatch_event_bus.test.arn
    role_arn = aws_iam_role.test.arn

    eventbridge_parameters {
      detail_type = %[3]q
      source      = %[4]q
    }
  }
}
`, scheduleName, eventBusName, detailType, source),
	)
}

func testAccScheduleConfig_targetInput(name, messageBody string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:${data.aws_partition.main.partition}:scheduler:::aws-sdk:sqs:sendMessage"
    role_arn = aws_iam_role.test.arn

    input = jsonencode({
      MessageBody = %[2]q
      QueueUrl    = aws_sqs_queue.test.url
    })
  }
}
`, name, messageBody),
	)
}

func testAccScheduleConfig_targetKinesisParameters(scheduleName, streamName, partitionKey string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[2]q
  shard_count = 1
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_kinesis_stream.test.arn
    role_arn = aws_iam_role.test.arn

    kinesis_parameters {
      partition_key = %[3]q
    }
  }
}
`, scheduleName, streamName, partitionKey),
	)
}

func testAccScheduleConfig_targetRetryPolicy(name string, maxEventAge, maxRetryAttempts int) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "dlq" {
  count = 2
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.test.arn

    retry_policy {
      maximum_event_age_in_seconds = %[2]d
      maximum_retry_attempts       = %[3]d
    }
  }
}
`, name, maxEventAge, maxRetryAttempts),
	)
}

func testAccScheduleConfig_targetRoleARN(name, resourceName string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_iam_role" "test1" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "scheduler.${data.aws_partition.main.dns_suffix}"
      }
    }
  })
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.test.arn
    role_arn = aws_iam_role.%[2]s.arn
  }
}
`, name, resourceName),
	)
}

func testAccScheduleConfig_targetSageMakerPipelineParameters1(name, name1, value1 string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
data "aws_region" "main" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:${data.aws_partition.main.partition}:sagemaker:${data.aws_region.main.name}:${data.aws_caller_identity.main.account_id}:pipeline/test"
    role_arn = aws_iam_role.test.arn

    sagemaker_pipeline_parameters {
      pipeline_parameter {
        name  = %[2]q
        value = %[3]q
      }
    }
  }
}
`, name, name1, value1),
	)
}

func testAccScheduleConfig_targetSageMakerPipelineParameters2(name, name1, value1, name2, value2 string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
data "aws_region" "main" {}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = "arn:${data.aws_partition.main.partition}:sagemaker:${data.aws_region.main.name}:${data.aws_caller_identity.main.account_id}:pipeline/test"
    role_arn = aws_iam_role.test.arn

    sagemaker_pipeline_parameters {
      pipeline_parameter {
        name  = %[2]q
        value = %[3]q
      }

      pipeline_parameter {
        name  = %[4]q
        value = %[5]q
      }
    }
  }
}
`, name, name1, value1, name2, value2),
	)
}

func testAccScheduleConfig_targetSQSParameters(name, messageGroupId string) string {
	return acctest.ConfigCompose(
		testAccScheduleConfig_base,
		fmt.Sprintf(`
resource "aws_sqs_queue" "test" {}

resource "aws_sqs_queue" "fifo" {
  fifo_queue = true
}

resource "aws_scheduler_schedule" "test" {
  name = %[1]q

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(1 hour)"

  target {
    arn      = aws_sqs_queue.fifo.arn
    role_arn = aws_iam_role.test.arn

    sqs_parameters {
      message_group_id = %[2]q
    }
  }
}
`, name, messageGroupId),
	)
}
