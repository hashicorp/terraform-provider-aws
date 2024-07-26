// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationTask_basic(t *testing.T) {
	t.Parallel()

	for _, migrationType := range enum.Values[awstypes.MigrationTypeValue]() { //nolint:paralleltest // false positive
		t.Run(migrationType, func(t *testing.T) {
			ctx := acctest.Context(t)
			rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
			resourceName := "aws_dms_replication_task.test"
			var v awstypes.ReplicationTask

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:                 func() { acctest.PreCheck(ctx, t) },
				ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
				Steps: []resource.TestStep{
					{
						Config: testAccReplicationTaskConfig_basic(rName, migrationType),
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckReplicationTaskExists(ctx, resourceName, &v),
							resource.TestCheckResourceAttr(resourceName, "replication_task_id", rName),
							acctest.MatchResourceAttrRegionalARN(resourceName, "replication_task_arn", "dms", regexache.MustCompile(`task:[A-Z0-9]{26}`)),
							resource.TestCheckResourceAttr(resourceName, "cdc_start_position", ""),
							resource.TestCheckNoResourceAttr(resourceName, "cdc_start_time"),
							resource.TestCheckResourceAttr(resourceName, "migration_type", migrationType),
							resource.TestCheckResourceAttrPair(resourceName, "replication_instance_arn", "aws_dms_replication_instance.test", "replication_instance_arn"),
							acctest.CheckResourceAttrEquivalentJSON(resourceName, "replication_task_settings", defaultReplicationTaskSettings[migrationType]),
							resource.TestCheckResourceAttrPair(resourceName, "source_endpoint_arn", "aws_dms_endpoint.source", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, "start_replication_task", acctest.CtFalse),
							resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ready"),
							acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
							resource.TestCheckResourceAttrPair(resourceName, "target_endpoint_arn", "aws_dms_endpoint.target", "endpoint_arn"),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
							resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct0),
						),
					},
					{
						ResourceName:            resourceName,
						ImportState:             true,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: []string{"start_replication_task"},
					},
				},
			})
		})
	}
}

func TestAccDMSReplicationTask_updateSettingsAndMappings(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1024, "ZedsDead"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1024"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "ZedsDead"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1024, "EMBRZ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1024"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "EMBRZ"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1248, "ZedsDead"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1248"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "ZedsDead"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1024, "ZedsDead"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1024"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "ZedsDead"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1248, "ZedsDead"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1248"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "ZedsDead"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_updateSettingsAndMappings(rName, 1024, "EMBRZ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "ChangeProcessingTuning.MemoryLimitTotal", "1024"),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", "length(rules)", acctest.Ct1),
					acctest.CheckResourceAttrJMES(resourceName, "table_mappings", `rules[0]."rule-name"`, "EMBRZ"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_settings_EnableLogging(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_settings_EnableLogging(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogGroup", fmt.Sprintf("dms-tasks-%s", rName)),
					func(s *terraform.State) error {
						arn, err := arn.Parse(aws.ToString(v.ReplicationTaskArn))
						if err != nil {
							return err
						}
						l := strings.Split(arn.Resource, ":")
						if len(l) != 2 {
							return fmt.Errorf("expected 2 parts in %s", arn.Resource)
						}
						id := l[1]
						return acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogStream", fmt.Sprintf("dms-task-%s", id))(s)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_settings_EnableLogContext(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogContext", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogGroup", fmt.Sprintf("dms-tasks-%s", rName)),
					func(s *terraform.State) error {
						arn, err := arn.Parse(aws.ToString(v.ReplicationTaskArn))
						if err != nil {
							return err
						}
						l := strings.Split(arn.Resource, ":")
						if len(l) != 2 {
							return fmt.Errorf("expected 2 parts in %s", arn.Resource)
						}
						id := l[1]
						return acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogStream", fmt.Sprintf("dms-task-%s", id))(s)
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_settings_EnableLogging(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogging", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_DEFAULT"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogGroup", fmt.Sprintf("dms-tasks-%s", rName)),
					func(s *terraform.State) error {
						arn, err := arn.Parse(aws.ToString(v.ReplicationTaskArn))
						if err != nil {
							return err
						}
						l := strings.Split(arn.Resource, ":")
						if len(l) != 2 {
							return fmt.Errorf("expected 2 parts in %s", arn.Resource)
						}
						id := l[1]
						return acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.CloudWatchLogStream", fmt.Sprintf("dms-task-%s", id))(s)
					},
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_settings_LoggingValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationTaskConfig_settings_EnableLogContext(rName, false, true),
				ExpectError: regexache.MustCompile(`The parameter Logging.EnableLogContext is not allowed when\s+Logging.EnableLogging is not set to true.`),
			},
			{
				Config:      testAccReplicationTaskConfig_settings_LoggingReadOnly(rName, "CloudWatchLogGroup"),
				ExpectError: regexache.MustCompile(`The parameter Logging.CloudWatchLogGroup is read-only and cannot be set.`),
			},
			{
				Config:      testAccReplicationTaskConfig_settings_LoggingReadOnly(rName, "CloudWatchLogStream"),
				ExpectError: regexache.MustCompile(`The parameter Logging.CloudWatchLogStream is read-only and cannot be set.`),
			},
		},
	})
}

func TestAccDMSReplicationTask_settings_LogComponents(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_settings_LogComponents(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogging", acctest.CtTrue),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.EnableLogContext", acctest.CtFalse),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "Logging.LogComponents[?Id=='DATA_STRUCTURE'].Severity | [0]", "LOGGER_SEVERITY_WARNING"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "type(Logging.CloudWatchLogGroup)", "string"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "type(Logging.CloudWatchLogStream)", "string"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
		},
	})
}

func TestAccDMSReplicationTask_settings_StreamBuffer(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_settings_StreamBuffer(rName, 4, 16),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "StreamBufferSettings.StreamBufferCount", acctest.Ct4),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "StreamBufferSettings.StreamBufferSizeInMB", "16"),
					acctest.CheckResourceAttrJMES(resourceName, "replication_task_settings", "StreamBufferSettings.CtrlStreamBufferSizeInMB", "5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication"},
			},
		},
	})
}

func TestAccDMSReplicationTask_cdcStartPosition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_cdcStartPosition(rName, "mysql-bin-changelog.000024:373"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_position", "mysql-bin-changelog.000024:373"),
					resource.TestCheckNoResourceAttr(resourceName, "cdc_start_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_resourceIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_resourceIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "resource_identifier", names.AttrIdentifier),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"resource_identifier", "start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_startReplicationTask(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_start(rName, true, "testrule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "running"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_start(rName, true, "changedtestrule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "running"),
				),
			},
			{
				Config: testAccReplicationTaskConfig_start(rName, false, "changedtestrule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "stopped"),
				),
			},
		},
	})
}

func TestAccDMSReplicationTask_s3ToRDS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	//https://github.com/hashicorp/terraform-provider-aws/issues/28277

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_s3ToRDS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
				),
			},
			{
				Config:             testAccReplicationTaskConfig_s3ToRDS(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDMSReplicationTask_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_basic(rName, "full-load"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceReplicationTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSReplicationTask_cdcStartTime_rfc3339_date(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	currentTime := time.Now().UTC()
	rfc3339Time := currentTime.Format(time.RFC3339)
	awsDmsExpectedOutput := strings.TrimRight(rfc3339Time, "Z") // AWS API drop "Z" part.

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_cdcStartTime(rName, rfc3339Time),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_position", awsDmsExpectedOutput),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_time", rfc3339Time),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_cdcStartTime_unix_timestamp(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	var v awstypes.ReplicationTask

	currentTime := time.Now().UTC()
	rfc3339Time := currentTime.Format(time.RFC3339)
	awsDmsExpectedOutput := strings.TrimRight(rfc3339Time, "Z") // AWS API drop "Z" part.
	dateTime, _ := time.Parse(time.RFC3339, rfc3339Time)
	unixDateTime := strconv.Itoa(int(dateTime.Unix()))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_cdcStartTime(rName, unixDateTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_position", awsDmsExpectedOutput),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_time", unixDateTime),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
		},
	})
}

func TestAccDMSReplicationTask_move(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	instanceOne := "aws_dms_replication_instance.test"
	instanceTwo := "aws_dms_replication_instance.test2"
	var v awstypes.ReplicationTask

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_move(rName, "aws_dms_replication_instance.test.replication_instance_arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, resourceName),
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_instance_arn", instanceOne, "replication_instance_arn"),
				),
			},
			{
				Config: testAccReplicationTaskConfig_move(rName, "aws_dms_replication_instance.test2.replication_instance_arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationInstanceExists(ctx, resourceName),
					testAccCheckReplicationTaskExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "replication_instance_arn", instanceTwo, "replication_instance_arn"),
				),
			},
		},
	})
}

func testAccCheckReplicationTaskExists(ctx context.Context, n string, v *awstypes.ReplicationTask) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

		output, err := tfdms.FindReplicationTaskByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckReplicationTaskDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_task" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

			_, err := tfdms.FindReplicationTaskByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Replication Task %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReplicationEndpointConfig_dummyDatabase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_dms_endpoint" "source" {
  database_name = %[1]q
  endpoint_id   = "%[1]s-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_endpoint" "target" {
  database_name = %[1]q
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}
`, rName))
}

func testAccReplicationTaskConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationEndpointConfig_dummyDatabase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t3.medium"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`, rName))
}

func testAccReplicationTaskConfig_basic(rName, migrationType string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = %[2]q
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
}
`, rName, migrationType))
}

func testAccReplicationTaskConfig_resourceIdentifier(rName string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  resource_identifier      = "identifier"
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
}
`, rName))
}

func testAccReplicationTaskConfig_updateSettingsAndMappings(rName string, memLimitTotal int, ruleName string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  # terrafmt can't handle this using jsonencode or a heredoc
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":%[2]d,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  # terrafmt can't handle this using jsonencode or a heredoc
  table_mappings = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"%[3]s\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
}
`, rName, memLimitTotal, ruleName))
}

func testAccReplicationTaskConfig_settings_EnableLogging(rName string, enabled bool) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
  # terrafmt can't handle this using jsonencode or a heredoc
  replication_task_settings = "{\"Logging\":{\"EnableLogging\":%[2]t}}"
}
`, rName, enabled))
}

func testAccReplicationTaskConfig_settings_EnableLogContext(rName string, enableLogging, enableLogContext bool) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
  # terrafmt can't handle this using jsonencode or a heredoc
  replication_task_settings = "{\"Logging\":{\"EnableLogging\":%[2]t,\"EnableLogContext\":%[3]t}}"
}
`, rName, enableLogging, enableLogContext))
}

func testAccReplicationTaskConfig_settings_LoggingReadOnly(rName, field string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
  # terrafmt can't handle this using jsonencode or a heredoc
  replication_task_settings = "{\"Logging\":{\"EnableLogging\":true, \"%[2]s\":\"value\"}}"
}
`, rName, field))
}

func testAccReplicationTaskConfig_settings_LogComponents(rName string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  replication_task_settings = jsonencode(
    {
      Logging = {
        EnableLogging = true,
        LogComponents = [{
          Id       = "DATA_STRUCTURE",
          Severity = "LOGGER_SEVERITY_WARNING"
        }]
      }
    }
  )
}
`, rName))
}

func testAccReplicationTaskConfig_settings_StreamBuffer(rName string, bufferCount, bufferSize int) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  # terrafmt can't handle this using jsonencode or a heredoc
  replication_task_settings = "{\"StreamBufferSettings\":{\"StreamBufferCount\":%[2]d,\"StreamBufferSizeInMB\":%[3]d}}"
}
	`, rName, bufferCount, bufferSize))
}

func testAccReplicationTaskConfig_cdcStartPosition(rName, cdcStartPosition string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[2]q
  migration_type           = "cdc"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  cdc_start_position = %[1]q
}
`, cdcStartPosition, rName))
}

func testAccReplicationTaskConfig_start(rName string, startTask bool, ruleName string) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base_ValidDatabase(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load-and-cdc"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  start_replication_task = %[2]t

  depends_on = [aws_rds_cluster_instance.source, aws_rds_cluster_instance.target]
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t3.medium"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids       = [aws_security_group.test.id]
}
`, rName, startTask, ruleName))
}

func testAccReplicationTaskConfig_s3ToRDS(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), testAccS3EndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load-and-cdc"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_s3_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_rds_engine_version" "default" {
  engine = "aurora-mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine        = data.aws_rds_engine_version.default.engine
  license_model = "general-public-license"
  storage_type  = "aurora"

  preferred_engine_versions  = ["5.7.mysql_aurora.2.11.2", data.aws_rds_engine_version.default.version]
  preferred_instance_classes = ["db.t2.small"]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier     = "%[1]s-aurora-cluster-target"
  engine                 = "aurora-mysql"
  engine_version         = data.aws_rds_orderable_db_instance.test.engine_version
  database_name          = "tftest"
  master_username        = "tftest"
  master_password        = "mustbeeightcharaters"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "test" {
  identifier           = "%[1]s-test-primary"
  cluster_identifier   = aws_rds_cluster.test.id
  instance_class       = "db.t2.small"
  engine               = aws_rds_cluster.test.engine
  engine_version       = aws_rds_cluster.test.engine_version
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_dms_s3_endpoint" "source" {
  bucket_folder           = "folder"
  bucket_name             = aws_s3_bucket.test.id
  cdc_path                = "cdc-files"
  csv_delimiter           = ";"
  csv_row_delimiter       = "\\n"
  date_partition_enabled  = false
  endpoint_id             = "%[1]s-source"
  endpoint_type           = "source"
  expected_bucket_owner   = data.aws_caller_identity.current.account_id
  ignore_header_rows      = 1
  rfc_4180                = false
  service_access_role_arn = aws_iam_role.test.arn
  ssl_mode                = "none"

  external_table_definition = jsonencode({
    TableCount = 1
    Tables = [{
      TableName  = "employee"
      TablePath  = "hr/employee/"
      TableOwner = "hr"
      TableColumns = [{
        ColumnName     = "ID"
        ColumnType     = "INT8"
        ColumnNullable = "false"
        ColumnIsPk     = "true"
        }, {
        ColumnName   = "LastName"
        ColumnType   = "STRING"
        ColumnLength = "20"
      }]
      TableColumnsTotal = "2"
    }]
  })

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_dms_endpoint" "target" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.test.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t3.medium"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids       = [aws_security_group.test.id]
}
`, rName))
}

func testAccReplicationTaskConfig_cdcStartTime(rName, cdcStartPosition string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[2]q
  migration_type           = "cdc"
  replication_instance_arn = aws_dms_replication_instance.test.replication_instance_arn
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )

  cdc_start_time = %[1]q
}
`, cdcStartPosition, rName))
}

func testAccReplicationTaskConfig_move(rName, arn string) string {
	return acctest.ConfigCompose(testAccReplicationTaskConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  replication_task_id      = %[1]q
  migration_type           = "full-load"
  replication_instance_arn = %[2]s
  source_endpoint_arn      = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn      = aws_dms_endpoint.target.endpoint_arn
  table_mappings = jsonencode(
    {
      "rules" = [
        {
          "rule-type" = "selection",
          "rule-id"   = "1",
          "rule-name" = "1",
          "object-locator" = {
            "schema-name" = "%%",
            "table-name"  = "%%"
          },
          "rule-action" = "include"
        }
      ]
    }
  )
}

resource "aws_dms_replication_instance" "test2" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.t3.medium"
  replication_instance_id      = "%[1]s-2"
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`, rName, arn))
}

var (
	defaultReplicationTaskSettings = map[string]string{
		"cdc":               defaultReplicationTaskCdcSettings,
		"full-load":         defaultReplicationTaskFullLoadSettings,
		"full-load-and-cdc": defaultReplicationTaskFullLoadAndCdcSettings,
	}

	//go:embed testdata/replication_task/defaults/cdc.json
	defaultReplicationTaskCdcSettings string

	//go:embed testdata/replication_task/defaults/full-load.json
	defaultReplicationTaskFullLoadSettings string

	//go:embed testdata/replication_task/defaults/full-load-and-cdc.json
	defaultReplicationTaskFullLoadAndCdcSettings string
)
