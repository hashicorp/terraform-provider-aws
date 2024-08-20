// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEvidentlyLaunch_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "evidently", fmt.Sprintf("project/%s/launch/%s", rName, rName3)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					// not returned at create time
					// resource.TestCheckResourceAttr(resourceName, "execution.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "execution.0.started_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "groups.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrLastUpdatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName3),
					resource.TestCheckResourceAttrPair(resourceName, "project", "aws_evidently_project.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "randomization_salt", rName3), // set to name if not specified
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.LaunchStatusCreated)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.LaunchTypeScheduledSplitsLaunch)),
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

func TestAccEvidentlyLaunch_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_description(rName, rName2, rName3, startTime, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_description(rName, rName2, rName3, startTime, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_updateGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName5 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_twoGroups(rName, rName2, rName3, rName4, rName5, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "groups.0.description", "first-group-add-desc"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1UpdatedName"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.description", "second-group"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.1.feature", "aws_evidently_feature.test2", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.1.name", "Variation2OriginalName"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.variation", "Variation2"),
				),
			},
			{
				Config: testAccLaunchConfig_threeGroups(rName, rName2, rName3, rName4, rName5, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "groups.0.description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.description", "second-group-update-desc"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.1.feature", "aws_evidently_feature.test2", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.1.name", "Variation2UpdatedName"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.variation", "Variation2a"),
					resource.TestCheckResourceAttr(resourceName, "groups.2.description", "third-group"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.2.feature", "aws_evidently_feature.test3", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.2.name", "Variation3OriginalName"),
					resource.TestCheckResourceAttr(resourceName, "groups.2.variation", "Variation3"),
				),
			},
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_updateMetricMonitors(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_oneMetricMonitor(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_twoMetricMonitors(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",11,\"<=\",22]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.entity_id_key", "entity_id_key2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",9,\"<=\",19]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.unit_label", "unit_label2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.value_key", "value_key2"),
				),
			},
			{
				Config: testAccLaunchConfig_threeMetricMonitors(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",15,\"<=\",25]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.entity_id_key", "entity_id_key2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",8,\"<=\",18]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.name", "name2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.unit_label", "unit_label2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.value_key", "value_key2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.entity_id_key", "entity_id_key3"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.name", "name3"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.value_key", "value_key3"),
				),
			},
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_updateRandomizationSalt(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	originalRandomizationSalt := "original randomization salt"
	updatedRandomizationSalt := "updated randomization salt"
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_randomizationSalt(rName, rName2, rName3, startTime, originalRandomizationSalt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "randomization_salt", originalRandomizationSalt),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_randomizationSalt(rName, rName2, rName3, startTime, updatedRandomizationSalt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "randomization_salt", updatedRandomizationSalt),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_scheduledSplitsConfig_updateSteps(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime1 := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	startTime2 := time.Now().AddDate(0, 0, 3).Format("2006-01-02T15:04:05Z")
	startTime3 := time.Now().AddDate(0, 0, 4).Format("2006-01-02T15:04:05Z")
	startTime4 := time.Now().AddDate(0, 0, 5).Format("2006-01-02T15:04:05Z")
	startTime5 := time.Now().AddDate(0, 0, 6).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigTwoStepsConfig(rName, rName2, rName3, startTime2, startTime3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", "15"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation2", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.Variation1", "20"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.Variation2", "25"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.start_time", startTime3),
				),
			},
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigThreeStepsConfig(rName, rName2, rName3, startTime3, startTime4, startTime5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", "60"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation2", "65"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime3),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.Variation1", "11"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.group_weights.Variation2", "12"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.1.start_time", startTime4),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.2.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.2.group_weights.Variation1", "44"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.2.group_weights.Variation2", "40"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.2.start_time", startTime5),
				),
			},
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime1),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_scheduledSplitsConfig_steps_updateSegmentOverrides(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName5 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName6 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigStepsOneSegmentOverrideConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.evaluation_order", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.segment", "aws_evidently_segment.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.Variation1", "20000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigStepsTwoSegmentOverridesConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.evaluation_order", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.segment", "aws_evidently_segment.test3", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.Variation2", "10000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.evaluation_order", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.segment", "aws_evidently_segment.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.Variation1", "40000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.Variation2", "30000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
				),
			},
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigStepsThreeSegmentOverridesConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.evaluation_order", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.segment", "aws_evidently_segment.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.Variation2", "5000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.evaluation_order", acctest.Ct3),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.segment", "aws_evidently_segment.test2", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.Variation1", "60000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.1.weights.Variation2", "70000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.2.evaluation_order", acctest.Ct4),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.2.segment", "aws_evidently_segment.test3", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.2.weights.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.2.weights.Variation1", "10000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.2.weights.Variation2", "90000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
				),
			},
			{
				Config: testAccLaunchConfig_scheduledSplitsConfigStepsOneSegmentOverrideConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.evaluation_order", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.segment", "aws_evidently_segment.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.segment_overrides.0.weights.Variation1", "20000"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EvidentlyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_tags2(rName, rName2, rName3, startTime, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var launch awstypes.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EvidentlyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(ctx, resourceName, &launch),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudwatchevidently.ResourceLaunch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLaunchDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_evidently_launch" {
				continue
			}

			launchName, projectNameOrARN, err := tfcloudwatchevidently.LaunchParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfcloudwatchevidently.FindLaunchWithProjectNameorARN(ctx, conn, launchName, projectNameOrARN)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Evidently Launch %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLaunchExists(ctx context.Context, n string, v *awstypes.Launch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Launch ID is set")
		}

		launchName, projectNameOrARN, err := tfcloudwatchevidently.LaunchParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyClient(ctx)

		output, err := tfcloudwatchevidently.FindLaunchWithProjectNameorARN(ctx, conn, launchName, projectNameOrARN)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLaunchConfigBase(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name = %[1]q
}

resource "aws_evidently_feature" "test" {
  name    = %[2]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }

  variations {
    name = "Variation1b"
    value {
      string_value = "test1b"
    }
  }
}
`, rName, rName2)
}

func testAccLaunchConfig_basic(rName, rName2, rName3, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime))
}

func testAccLaunchConfig_description(rName, rName2, rName3, startTime, description string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name        = %[1]q
  project     = aws_evidently_project.test.name
  description = %[3]q

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime, description))
}

func testAccLaunchConfigGroupsBase(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_feature" "test2" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation2"
    value {
      string_value = "test2"
    }
  }

  variations {
    name = "Variation2a"
    value {
      string_value = "test2a"
    }
  }
}

resource "aws_evidently_feature" "test3" {
  name    = %[2]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation3"
    value {
      string_value = "test3"
    }
  }
}
`, rName, rName2)
}

func testAccLaunchConfig_twoGroups(rName, rName2, rName3, rName4, rName5, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigGroupsBase(rName3, rName4),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature     = aws_evidently_feature.test.name
    name        = "Variation1UpdatedName"
    variation   = "Variation1"
    description = "first-group-add-desc"
  }

  groups {
    feature     = aws_evidently_feature.test2.name
    name        = "Variation2OriginalName"
    variation   = "Variation2"
    description = "second-group"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1UpdatedName"  = 0
        "Variation2OriginalName" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName5, startTime))
}

func testAccLaunchConfig_threeGroups(rName, rName2, rName3, rName4, rName5, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigGroupsBase(rName3, rName4),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature     = aws_evidently_feature.test2.name
    name        = "Variation2UpdatedName"
    variation   = "Variation2a"
    description = "second-group-update-desc"
  }

  groups {
    feature     = aws_evidently_feature.test3.name
    name        = "Variation3OriginalName"
    variation   = "Variation3"
    description = "third-group"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1"             = 0
        "Variation2UpdatedName"  = 0
        "Variation3OriginalName" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName5, startTime))
}

func testAccLaunchConfig_oneMetricMonitor(rName, rName2, rName3, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
      name          = "name1"
      unit_label    = "unit_label1"
      value_key     = "value_key1"
    }
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime))
}

func testAccLaunchConfig_twoMetricMonitors(rName, rName2, rName3, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1a"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",11,\"<=\",22]}]}"
      name          = "name1a"
      unit_label    = "unit_label1a"
      value_key     = "value_key1a"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key2"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",9,\"<=\",19]}]}"
      name          = "name2"
      unit_label    = "unit_label2"
      value_key     = "value_key2"
    }
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime))
}

func testAccLaunchConfig_threeMetricMonitors(rName, rName2, rName3, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1b"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",15,\"<=\",25]}]}"
      name          = "name1b"
      unit_label    = "unit_label1b"
      value_key     = "value_key1b"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key2a"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",8,\"<=\",18]}]}"
      name          = "name2a"
      unit_label    = "unit_label2a"
      value_key     = "value_key2a"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key3"
      name          = "name3"
      value_key     = "value_key3"
    }
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime))
}

func testAccLaunchConfig_randomizationSalt(rName, rName2, rName3, startTime, randomizationSalt string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name               = %[1]q
  project            = aws_evidently_project.test.name
  randomization_salt = %[3]q

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime, randomizationSalt))
}

func testAccLaunchConfig_scheduledSplitsConfigTwoStepsConfig(rName, rName2, rName3, startTime, startTime2 string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation2"
    variation = "Variation2"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 15
        "Variation2" = 10
      }
      start_time = %[2]q
    }

    steps {
      group_weights = {
        "Variation1" = 20
        "Variation2" = 25
      }
      start_time = %[3]q
    }
  }
}
`, rName3, startTime, startTime2))
}

func testAccLaunchConfig_scheduledSplitsConfigThreeStepsConfig(rName, rName2, rName3, startTime, startTime2, startTime3 string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation2"
    variation = "Variation2"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 60
        "Variation2" = 65
      }
      start_time = %[2]q
    }

    steps {
      group_weights = {
        "Variation1" = 11
        "Variation2" = 12
      }
      start_time = %[3]q
    }

    steps {
      group_weights = {
        "Variation1" = 44
        "Variation2" = 40
      }
      start_time = %[4]q
    }
  }
}
`, rName3, startTime, startTime2, startTime3))
}

func testAccLaunchConfigSegmentOverridesBase(rName, rName2, rName3 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_segment" "test" {
  name    = %[1]q
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
}

resource "aws_evidently_segment" "test2" {
  name    = %[2]q
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
}

resource "aws_evidently_segment" "test3" {
  name    = %[3]q
  pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
}
`, rName, rName2, rName3)
}

func testAccLaunchConfig_scheduledSplitsConfigStepsOneSegmentOverrideConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigSegmentOverridesBase(rName3, rName4, rName5),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }

      segment_overrides {
        evaluation_order = 1
        segment          = aws_evidently_segment.test.name

        weights = {
          "Variation1" = 20000
        }
      }

      start_time = %[2]q
    }
  }
}
`, rName6, startTime))
}

func testAccLaunchConfig_scheduledSplitsConfigStepsTwoSegmentOverridesConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigSegmentOverridesBase(rName3, rName4, rName5),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation2"
    variation = "Variation2"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
        "Variation2" = 0
      }

      segment_overrides {
        evaluation_order = 1
        segment          = aws_evidently_segment.test3.name

        weights = {
          "Variation2" = 10000
        }
      }

      segment_overrides {
        evaluation_order = 2
        segment          = aws_evidently_segment.test2.name

        weights = {
          "Variation1" = 40000
          "Variation2" = 30000
        }
      }

      start_time = %[2]q
    }
  }
}
`, rName6, startTime))
}

func testAccLaunchConfig_scheduledSplitsConfigStepsThreeSegmentOverridesConfig(rName, rName2, rName3, rName4, rName5, rName6, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigSegmentOverridesBase(rName3, rName4, rName5),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation2"
    variation = "Variation2"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
        "Variation2" = 0
      }

      segment_overrides {
        evaluation_order = 1
        segment          = aws_evidently_segment.test.name

        weights = {
          "Variation2" = 5000
        }
      }

      segment_overrides {
        evaluation_order = 3
        segment          = aws_evidently_segment.test2.name

        weights = {
          "Variation1" = 60000
          "Variation2" = 70000
        }
      }

      segment_overrides {
        evaluation_order = 4
        segment          = aws_evidently_segment.test3.name

        weights = {
          "Variation1" = 10000
          "Variation2" = 90000
        }
      }

      start_time = %[2]q
    }
  }
}
`, rName6, startTime))
}

func testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, tag, value string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName3, startTime, tag, value))
}

func testAccLaunchConfig_tags2(rName, rName2, rName3, startTime, tag1, value1, tag2, value2 string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName3, startTime, tag1, value1, tag2, value2))
}
