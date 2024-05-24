// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationautoscaling/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppAutoScalingScheduledAction_dynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	var sa1, sa2 awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedule1 := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	schedule2 := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05")
	updatedTimezone := "Pacific/Tahiti"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_dynamoDB(rName, schedule1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", schedule1)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccScheduledActionConfig_dynamoDBUpdated(rName, schedule2, updatedTimezone),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa2),
					testAccCheckScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", schedule2)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "9"),
					resource.TestCheckResourceAttr(resourceName, "timezone", updatedTimezone),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedule := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")

	resourceName := "aws_appautoscaling_scheduled_action.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_dynamoDB(rName, schedule),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappautoscaling.ResourceScheduledAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ecs(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_ecs(rName, ts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ecsUpdateScheduleRetainStartAndEndTime(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	tsUpdate := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, ts, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
			{
				Config: testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, tsUpdate, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", tsUpdate)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ecsUpdateStartAndEndTime(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	tsUpdate := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	startTimeUpdate := time.Now().AddDate(0, 0, 4).Format("2006-01-02T15:04:05Z")
	endTimeUpdate := time.Now().AddDate(0, 0, 10).Format("2006-01-02T15:04:05Z")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, ts, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
			{
				Config: testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, tsUpdate, startTimeUpdate, endTimeUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", tsUpdate)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTimeUpdate),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTimeUpdate),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ecsAddStartTimeAndEndTimeAfterResourceCreated(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	tsUpdate := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_ecs(rName, ts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
			{
				Config: testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, tsUpdate, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", tsUpdate)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_emr(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_emr(rName, ts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_Name_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	var sa1, sa2 awstypes.ScheduledAction
	resourceName := "aws_appautoscaling_scheduled_action.test"
	resourceName2 := "aws_appautoscaling_scheduled_action.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_duplicateName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa1),
					testAccCheckScheduledActionExists(ctx, resourceName2, &sa2),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_spotFleet(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_spotFleet(rName, ts, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleAtExpression_timezone(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	at := fmt.Sprintf("at(%s)", ts)
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_timezone(rName, at, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, at),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleCronExpression_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cron := "cron(0 17 * * ? *)"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_schedule(rName, cron),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleCronExpression_timezone(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cron := "cron(0 17 * * ? *)"
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_timezone(rName, cron, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleCronExpression_startEndTimeTimezone(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	cron := "cron(0 17 * * ? *)"
	scheduleTimezone := "Etc/GMT+9"                                    // Z-09:00 (IANA and RFC3339 have inverted signs)
	startTimezone, _ := time.LoadLocation("Antarctica/DumontDUrville") // Z+10:00
	endTimezone, _ := time.LoadLocation("America/Vancouver")           // Z-08:00
	startTime := time.Now().AddDate(0, 0, 2).In(startTimezone)
	startTimeUtc := startTime.UTC()
	endTime := time.Now().AddDate(0, 0, 8).In(endTimezone)
	endTimeUtc := endTime.UTC()
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_timezone(rName, cron, scheduleTimezone, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", scheduleTimezone),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTimeUtc.Format(time.RFC3339)),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTimeUtc.Format(time.RFC3339)),
				),
			},
			{
				Config: testAccScheduledActionConfig_schedule(rName, cron),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, ""),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleRateExpression_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rate := "rate(1 day)"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_schedule(rName, rate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, rate),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_ScheduleRateExpression_timezone(t *testing.T) {
	ctx := acctest.Context(t)
	var sa awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rate := "rate(1 day)"
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_timezone(rName, rate, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, rate),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, names.AttrStartTime, startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_minCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var sa1, sa2 awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedule := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_minCapacity(rName, schedule, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", schedule)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccScheduledActionConfig_minCapacity(rName, schedule, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa2),
					testAccCheckScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
				),
			},
			{
				Config: testAccScheduledActionConfig_maxCapacity(rName, schedule, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa2),
					testAccCheckScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
				),
			},
		},
	})
}

func TestAccAppAutoScalingScheduledAction_maxCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var sa1, sa2 awstypes.ScheduledAction
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	schedule := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppAutoScalingServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScheduledActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScheduledActionConfig_maxCapacity(rName, schedule, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, autoscalingTargetResourceName, names.AttrResourceID),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSchedule, fmt.Sprintf("at(%s)", schedule)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "autoscaling", regexache.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrStartTime),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccScheduledActionConfig_maxCapacity(rName, schedule, 8),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa2),
					testAccCheckScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "8"),
				),
			},
			{
				Config: testAccScheduledActionConfig_minCapacity(rName, schedule, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScheduledActionExists(ctx, resourceName, &sa2),
					testAccCheckScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
				),
			},
		},
	})
}

func testAccCheckScheduledActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appautoscaling_scheduled_action" {
				continue
			}

			_, err := tfappautoscaling.FindScheduledActionByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["scalable_dimension"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Application Auto Scaling Scheduled Action %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckScheduledActionExists(ctx context.Context, n string, v *awstypes.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppAutoScalingClient(ctx)

		output, err := tfappautoscaling.FindScheduledActionByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes[names.AttrResourceID], rs.Primary.Attributes["scalable_dimension"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckScheduledActionNotRecreated(i, j *awstypes.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreationTime).Equal(aws.ToTime(j.CreationTime)) {
			return fmt.Errorf("Application Auto Scaling Scheduled Action recreated")
		}

		return nil
	}
}

func testAccScheduledActionConfig_dynamoDB(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts)
}

func testAccScheduledActionConfig_dynamoDBUpdated(rName, ts, timezone string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"
  timezone = %[3]q

  scalable_target_action {
    min_capacity = 2
    max_capacity = 9
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, timezone)
}

func testAccScheduledActionConfig_ecs(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}
`, rName, ts)
}

func testAccScheduledActionConfig_ecsWithStartAndEndTime(rName, ts, startTime, endTime string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  start_time = %[3]q
  end_time   = %[4]q

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}
`, rName, ts, startTime, endTime)
}

func testAccScheduledActionConfig_emr(rName, ts string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "elasticmapreduce"
  resource_id        = "instancegroup/${aws_emr_cluster.test.id}/${aws_emr_instance_group.test.id}"
  scalable_dimension = "elasticmapreduce:instancegroup:InstanceCount"
  role_arn           = aws_iam_role.autoscale_role.arn
  min_capacity       = 1
  max_capacity       = 5
}

data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.4.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 2
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [aws_route_table_association.test]

  service_role     = aws_iam_role.emr_role.arn
  autoscaling_role = aws_iam_role.autoscale_role.arn
}

resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 1
  instance_type  = "c4.large"
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_iam_role" "emr_role" {
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_role" {
  role       = aws_iam_role.emr_role.id
  policy_arn = aws_iam_policy.emr_policy.arn
}

resource "aws_iam_policy" "emr_policy" {
  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Resource": "*",
    "Action": [
      "ec2:AuthorizeSecurityGroupEgress",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:CancelSpotInstanceRequests",
      "ec2:CreateNetworkInterface",
      "ec2:CreateSecurityGroup",
      "ec2:CreateTags",
      "ec2:DeleteNetworkInterface",
      "ec2:DeleteSecurityGroup",
      "ec2:DeleteTags",
      "ec2:DescribeAvailabilityZones",
      "ec2:DescribeAccountAttributes",
      "ec2:DescribeDhcpOptions",
      "ec2:DescribeInstanceStatus",
      "ec2:DescribeInstances",
      "ec2:DescribeKeyPairs",
      "ec2:DescribeNetworkAcls",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribePrefixLists",
      "ec2:DescribeRouteTables",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSpotInstanceRequests",
      "ec2:DescribeSpotPriceHistory",
      "ec2:DescribeSubnets",
      "ec2:DescribeVpcAttribute",
      "ec2:DescribeVpcEndpoints",
      "ec2:DescribeVpcEndpointServices",
      "ec2:DescribeVpcs",
      "ec2:DetachNetworkInterface",
      "ec2:ModifyImageAttribute",
      "ec2:ModifyInstanceAttribute",
      "ec2:RequestSpotInstances",
      "ec2:RevokeSecurityGroupEgress",
      "ec2:RunInstances",
      "ec2:TerminateInstances",
      "ec2:DeleteVolume",
      "ec2:DescribeVolumeStatus",
      "ec2:DescribeVolumes",
      "ec2:DetachVolume",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListInstanceProfiles",
      "iam:ListRolePolicies",
      "iam:PassRole",
      "s3:CreateBucket",
      "s3:Get*",
      "s3:List*",
      "sdb:BatchPutAttributes",
      "sdb:Select",
      "sqs:CreateQueue",
      "sqs:Delete*",
      "sqs:GetQueue*",
      "sqs:PurgeQueue",
      "sqs:ReceiveMessage"
    ]
  }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "instance_role" {
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "instance_profile" {
  name = %[1]q
  role = aws_iam_role.instance_role.name
}

resource "aws_iam_role_policy_attachment" "instance_role" {
  role       = aws_iam_role.instance_role.id
  policy_arn = aws_iam_policy.instance_policy.arn
}

resource "aws_iam_policy" "instance_policy" {
  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Resource": "*",
    "Action": [
      "cloudwatch:*",
      "dynamodb:*",
      "ec2:Describe*",
      "elasticmapreduce:Describe*",
      "elasticmapreduce:ListBootstrapActions",
      "elasticmapreduce:ListClusters",
      "elasticmapreduce:ListInstanceGroups",
      "elasticmapreduce:ListInstances",
      "elasticmapreduce:ListSteps",
      "kinesis:CreateStream",
      "kinesis:DeleteStream",
      "kinesis:DescribeStream",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:MergeShards",
      "kinesis:PutRecord",
      "kinesis:SplitShard",
      "rds:Describe*",
      "s3:*",
      "sdb:*",
      "sns:*",
      "sqs:*"
    ]
  }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "autoscale_role" {
  assume_role_policy = data.aws_iam_policy_document.autoscale_role.json
}

data "aws_iam_policy_document" "autoscale_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "autoscale_role" {
  role       = aws_iam_role.autoscale_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, rName, ts))
}

func testAccScheduledActionConfig_duplicateName(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test2" {
  name           = "%[1]s-2"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test2" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test2.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test2" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test2.service_namespace
  resource_id        = aws_appautoscaling_target.test2.resource_id
  scalable_dimension = aws_appautoscaling_target.test2.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
`, rName)
}

func testAccScheduledActionConfig_spotFleet(rName, ts, validUntil string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 3
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[3]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "t3.micro"
    ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  }
}
`, rName, ts, validUntil))
}

func testAccScheduledActionConfig_schedule(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = %[2]q

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, schedule)
}

func testAccScheduledActionConfig_timezone(rName, schedule, timezone, startTime, endTime string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  timezone   = %[2]q
  schedule   = %[3]q
  start_time = %[4]q
  end_time   = %[5]q

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, timezone, schedule, startTime, endTime)
}

func testAccScheduledActionConfig_minCapacity(rName, ts string, minCapacity int) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    min_capacity = %[3]d
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, minCapacity)
}

func testAccScheduledActionConfig_maxCapacity(rName, ts string, maxCapacity int) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    max_capacity = %[3]d
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, maxCapacity)
}
