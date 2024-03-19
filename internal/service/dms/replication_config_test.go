// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSReplicationConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.availability_zone", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.dns_name_servers", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "128"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.multi_az", "false"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.preferred_maintenance_window", "sun:23:45-mon:00:30"),
					resource.TestCheckResourceAttrSet(resourceName, "compute_config.0.replication_subnet_group_id"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.vpc_security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "replication_config_identifier", rName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_settings"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckNoResourceAttr(resourceName, "resource_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "source_endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
					resource.TestCheckResourceAttr(resourceName, "supplemental_settings", ""),
					resource.TestCheckResourceAttrSet(resourceName, "table_mappings"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "target_endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
		},
	})
}

func TestAccDMSReplicationConfig_noChangeOnDefault(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_noChangeOnDefault(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.availability_zone", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.dns_name_servers", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "128"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.multi_az", "false"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.preferred_maintenance_window", "sun:23:45-mon:00:30"),
					resource.TestCheckResourceAttrSet(resourceName, "compute_config.0.replication_subnet_group_id"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.vpc_security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "replication_config_identifier", rName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_settings"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckNoResourceAttr(resourceName, "resource_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "source_endpoint_arn"),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
					resource.TestCheckResourceAttr(resourceName, "supplemental_settings", ""),
					resource.TestCheckResourceAttrSet(resourceName, "table_mappings"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "target_endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
		},
	})
}

func TestAccDMSReplicationConfig_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceReplicationConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDMSReplicationConfig_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 2, 16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "2"),
				),
			},
			{
				Config: testAccReplicationConfigConfig_update(rName, "cdc", 4, 32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "replication_type", "cdc"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.max_capacity_units", "32"),
					resource.TestCheckResourceAttr(resourceName, "compute_config.0.min_capacity_units", "4"),
				),
			},
		},
	})
}

func TestAccDMSReplicationConfig_startReplication(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigConfig_startReplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"start_replication", "resource_identifier"},
			},
			{
				Config: testAccReplicationConfigConfig_startReplication(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "start_replication", "false"),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

		_, err := tfdms.FindReplicationConfigByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_replication_config" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).DMSConn(ctx)

			_, err := tfdms.FindReplicationConfigByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Replication Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccRDSClustersConfig_base configures a pair of Aurora RDS clusters (and instances) ready for replication.
func testAccRDSClustersConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
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

data "aws_rds_engine_version" "default" {
  engine = "aurora-mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t3.medium", "db.t3.large"]
}

resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%[1]s-pg-cluster"
  family      = data.aws_rds_engine_version.default.parameter_group_family
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

resource "aws_rds_cluster" "source" {
  cluster_identifier              = "%[1]s-aurora-cluster-source"
  engine                          = data.aws_rds_orderable_db_instance.test.engine
  engine_version                  = data.aws_rds_orderable_db_instance.test.engine_version
  database_name                   = "tftest"
  master_username                 = "tftest"
  master_password                 = "mustbeeightcharaters"
  skip_final_snapshot             = true
  vpc_security_group_ids          = [aws_security_group.test.id]
  db_subnet_group_name            = aws_db_subnet_group.test.name
  db_cluster_parameter_group_name = aws_rds_cluster_parameter_group.test.name
}

resource "aws_rds_cluster_instance" "source" {
  identifier           = "%[1]s-source-primary"
  cluster_identifier   = aws_rds_cluster.source.id
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster" "target" {
  cluster_identifier     = "%[1]s-aurora-cluster-target"
  engine                 = data.aws_rds_orderable_db_instance.test.engine
  engine_version         = data.aws_rds_orderable_db_instance.test.engine_version
  database_name          = "tftest"
  master_username        = "tftest"
  master_password        = "mustbeeightcharaters"
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.test.id]
  db_subnet_group_name   = aws_db_subnet_group.test.name
}

resource "aws_rds_cluster_instance" "target" {
  identifier           = "%[1]s-target-primary"
  cluster_identifier   = aws_rds_cluster.target.id
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  db_subnet_group_name = aws_db_subnet_group.test.name
}
`, rName))
}

func testAccReplicationConfigConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccRDSClustersConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_id          = %[1]q
  replication_subnet_group_description = "terraform test"
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_endpoint" "target" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-target"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.target.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}

resource "aws_dms_endpoint" "source" {
  database_name = "tftest"
  endpoint_id   = "%[1]s-source"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = aws_rds_cluster.source.endpoint
  port          = 3306
  username      = "tftest"
  password      = "mustbeeightcharaters"
}
`, rName))
}

func testAccReplicationConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  replication_settings          = "{\"BeforeImageSettings\":null,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"CommitPositionTableEnabled\":false,\"ControlSchema\":\"\",\"FullLoadExceptionTableEnabled\":false,\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"EventErrorPolicy\":\"IGNORE\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorStopRetryAfterThrottlingMax\":false,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FailTaskWhenCleanTaskResourceFailed\":false,\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TTSettings\":{\"EnableTT\":false,\"FailTaskOnTTFailure\":false,\"TTRecordSettings\":null,\"TTS3Settings\":null},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false}}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName))
}

func testAccReplicationConfigConfig_noChangeOnDefault(rName string) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"
  replication_settings          = "{\"Logging\":{\"EnableLogging\":true}}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName))
}

func testAccReplicationConfigConfig_update(rName, replicationType string, minCapacity, maxCapacity int) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = %[2]q
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "%[3]d"
    min_capacity_units           = "%[4]d"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, replicationType, maxCapacity, minCapacity))
}

func testAccReplicationConfigConfig_startReplication(rName string, start bool) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  resource_identifier           = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  start_replication = %[2]t

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }
}
`, rName, start))
}

func testAccReplicationConfigConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccReplicationConfigConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccReplicationConfigConfig_base(rName), fmt.Sprintf(`
resource "aws_dms_replication_config" "test" {
  replication_config_identifier = %[1]q
  replication_type              = "cdc"
  source_endpoint_arn           = aws_dms_endpoint.source.endpoint_arn
  target_endpoint_arn           = aws_dms_endpoint.target.endpoint_arn
  table_mappings                = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  compute_config {
    replication_subnet_group_id  = aws_dms_replication_subnet_group.test.replication_subnet_group_id
    max_capacity_units           = "128"
    min_capacity_units           = "2"
    preferred_maintenance_window = "sun:23:45-mon:00:30"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
