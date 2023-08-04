// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDMSReplicationTask_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"

	tags := `
    Update = "to-update"
    Remove = "to-remove"
`
	updatedTags := `
    Update = "updated"
    Add    = "added"
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_basic(rName, tags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
				),
			},
			{
				Config:             testAccReplicationTaskConfig_basic(rName, tags),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication_task"},
			},
			{
				Config: testAccReplicationTaskConfig_basic(rName, updatedTags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccDMSReplicationTask_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1024, "ZedsDead"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1024")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"ZedsDead")),
				),
			},
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1024, "EMBRZ"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1024")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"EMBRZ")),
				),
			},
			{
				Config:             testAccReplicationTaskConfig_update(rName, "full-load", 1024, "EMBRZ"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1248, "ZedsDead"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1248")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"ZedsDead")),
				),
			},
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1024, "ZedsDead"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1024")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"ZedsDead")),
				),
			},
			{
				Config:             testAccReplicationTaskConfig_update(rName, "full-load", 1024, "ZedsDead"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1248, "ZedsDead"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1248")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"ZedsDead")),
				),
			},
			{
				Config: testAccReplicationTaskConfig_update(rName, "full-load", 1024, "EMBRZ"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "migration_type", "full-load"),
					resource.TestMatchResourceAttr(resourceName, "replication_task_settings", regexp.MustCompile("MemoryLimitTotal\":1024")),
					resource.TestMatchResourceAttr(resourceName, "table_mappings", regexp.MustCompile("rule-name\":\"EMBRZ")),
				),
			},
			{
				Config:             testAccReplicationTaskConfig_update(rName, "full-load", 1024, "EMBRZ"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDMSReplicationTask_cdcStartPosition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_cdcStartPosition(rName, "mysql-bin-changelog.000024:373"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "cdc_start_position", "mysql-bin-changelog.000024:373"),
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

func TestAccDMSReplicationTask_startReplicationTask(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_start(rName, true, "testrule"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "running"),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "running"),
				),
			},
			{
				Config: testAccReplicationTaskConfig_start(rName, false, "changedtestrule"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "stopped"),
				),
			},
		},
	})
}

func TestAccDMSReplicationTask_s3ToRDS(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"

	//https://github.com/hashicorp/terraform-provider-aws/issues/28277

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_s3ToRDS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
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

	tags := `
    Test = "test"
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskConfig_basic(rName, tags),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceReplicationTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckReplicationTaskExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

		_, err := tfdms.FindReplicationTaskByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationTaskDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_task" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

			_, err := tfdms.FindReplicationTaskByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS replication task (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func replicationTaskConfigBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

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

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.test1.id, aws_subnet.test2.id]
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.c4.large"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}
`, rName))
}

func testAccReplicationTaskConfig_basic(rName, tags string) string {
	return acctest.ConfigCompose(
		replicationTaskConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  migration_type            = "full-load"
  replication_instance_arn  = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id       = %[1]q
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  source_endpoint_arn       = aws_dms_endpoint.source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  tags = {
    Name = %[1]q
%[2]s
  }

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn
}
`, rName, tags))
}

func testAccReplicationTaskConfig_update(rName, migType string, memLimitTotal int, ruleName string) string {
	return acctest.ConfigCompose(
		replicationTaskConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  migration_type            = %[2]q
  replication_instance_arn  = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id       = %[1]q
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":%[3]d,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  source_endpoint_arn       = aws_dms_endpoint.source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"%[4]s\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  tags = {
    Name = %[1]q
  }

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn
}
`, rName, migType, memLimitTotal, ruleName))
}

func testAccReplicationTaskConfig_cdcStartPosition(rName, cdcStartPosition string) string {
	return acctest.ConfigCompose(
		replicationTaskConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_replication_task" "test" {
  cdc_start_position        = %[1]q
  migration_type            = "cdc"
  replication_instance_arn  = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id       = %[2]q
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  source_endpoint_arn       = aws_dms_endpoint.source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  target_endpoint_arn       = aws_dms_endpoint.target.endpoint_arn
}
`, cdcStartPosition, rName))
}

func testAccReplicationTaskConfig_start(rName string, startTask bool, ruleName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_security_group" "test" {
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
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%[1]s-pg-cluster"
  family      = "aurora-mysql5.7"
  description = "DMS cluster parameter group"

  parameter {
    name         = "binlog_format"
    value        = "ROW"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_row_image"
    value        = "Full"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "binlog_checksum"
    value        = "NONE"
    apply_method = "pending-reboot"
  }
}

resource "aws_rds_cluster" "test1" {
  cluster_identifier              = "%[1]s-aurora-cluster-source"
  engine                          = "aurora-mysql"
  engine_version                  = "5.7.mysql_aurora.2.11.2"
  database_name                   = "tftest"
  master_username                 = "tftest"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
}

resource "aws_rds_cluster_instance" "test1" {
  identifier           = "%[1]s-test1-primary"
  cluster_identifier   = aws_rds_cluster.test1.id
  instance_class       = "db.t2.small"
  engine               = aws_rds_cluster.test1.engine
  engine_version       = aws_rds_cluster.test1.engine_version
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster" "test2" {
  cluster_identifier     = "%[1]s-aurora-cluster-target"
  engine                 = "aurora-mysql"
  engine_version         = "5.7.mysql_aurora.2.11.2"
  database_name          = "tftest"
  master_username        = "tftest"
  master_password        = "mustbeeightcharaters"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "test2" {
  identifier           = "%[1]s-test2-primary"
  cluster_identifier   = aws_rds_cluster.test2.id
  instance_class       = "db.t2.small"
  engine               = aws_rds_cluster.test2.engine
  engine_version       = aws_rds_cluster.test2.engine_version
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_dms_endpoint" "source" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.test1.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_endpoint" "target" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.test2.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.test1.id, aws_subnet.test2.id]
}

resource "aws_dms_replication_instance" "test" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.c4.large"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids       = [aws_security_group.test.id]
}

resource "aws_dms_replication_task" "test" {
  migration_type            = "full-load-and-cdc"
  replication_instance_arn  = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id       = %[1]q
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  source_endpoint_arn       = aws_dms_endpoint.source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"%[3]s\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  start_replication_task = %[2]t

  tags = {
    Name = %[1]q
  }

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn

  depends_on = [aws_rds_cluster_instance.test1, aws_rds_cluster_instance.test2]
}
`, rName, startTask, ruleName))
}

func testAccReplicationTaskConfig_s3ToRDS(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccS3EndpointConfig_base(rName),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_security_group" "test" {
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
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.test1.id, aws_subnet.test2.id]
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
  replication_instance_class   = "dms.c4.large"
  replication_instance_id      = %[1]q
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
  vpc_security_group_ids       = [aws_security_group.test.id]
}

resource "aws_dms_replication_task" "test" {
  migration_type            = "full-load-and-cdc"
  replication_instance_arn  = aws_dms_replication_instance.test.replication_instance_arn
  replication_task_id       = %[1]q
  replication_task_settings = "{\"BeforeImageSettings\":null,\"FailTaskWhenCleanTaskResourceFailed\":false,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false},\"TTSettings\":{\"EnableTT\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null}}"
  source_endpoint_arn       = aws_dms_s3_endpoint.source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  tags = {
    Name = %[1]q
  }

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn

  depends_on = [aws_rds_cluster_instance.test]
}
`, rName))
}
