package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDmsReplicationTask_basic(t *testing.T) {
	resourceName := "aws_dms_replication_task.dms_replication_task"
	randId := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: dmsReplicationTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsReplicationTaskConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsReplicationTaskExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "replication_task_arn"),
				),
			},
			{
				Config:             dmsReplicationTaskConfig(randId),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: dmsReplicationTaskConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsReplicationTaskExists(resourceName),
				),
			},
		},
	})
}

func checkDmsReplicationTaskExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeReplicationTasks(&dms.DescribeReplicationTasksInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-task-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return err
		}

		if resp.ReplicationTasks == nil {
			return fmt.Errorf("DMS replication task error: %v", err)
		}
		return nil
	}
}

func dmsReplicationTaskDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_replication_task" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeReplicationTasks(&dms.DescribeReplicationTasksInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("replication-task-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return nil
		}

		if resp != nil && len(resp.ReplicationTasks) > 0 {
			return fmt.Errorf("DMS replication task still exists: %v", err)
		}
	}

	return nil
}

func dmsReplicationTaskConfig(randId string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "dms_assume_role_policy_document" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["dms.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "dms_vpc_role" {
  assume_role_policy = data.aws_iam_policy_document.dms_assume_role_policy_document.json
  name = "dms-vpc-role"
}

resource "aws_iam_role" "dms_cloudwatch_logs_role" {
  assume_role_policy = data.aws_iam_policy_document.dms_assume_role_policy_document.json
  name = "dms-cloudwatch-logs-role"
}

resource "aws_iam_role_policy_attachment" "dms_vpc_access_policy" {
  role       = aws_iam_role.dms_vpc_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDMSVPCManagementRole"
}

resource "aws_iam_role_policy_attachment" "dms_cloudwatch_logs_access_policy" {
  role       = aws_iam_role.dms_cloudwatch_logs_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDMSCloudWatchLogsRole"
}

resource "aws_vpc" "dms_vpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-dms-replication-task-%[1]s"
  }

  depends_on = [
    aws_iam_role_policy_attachment.dms_vpc_access_policy,
    aws_iam_role_policy_attachment.dms_cloudwatch_logs_access_policy,
  ]
}

resource "aws_subnet" "dms_subnet_1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-task-1-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_subnet" "dms_subnet_2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-task-2-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_dms_endpoint" "dms_endpoint_source" {
  database_name = "tf-test-dms-db"
  endpoint_id   = "tf-test-dms-endpoint-source-%[1]s"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_endpoint" "dms_endpoint_target" {
  database_name = "tf-test-dms-db"
  endpoint_id   = "tf-test-dms-endpoint-target-%[1]s"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_replication_subnet_group" "dms_replication_subnet_group" {
  replication_subnet_group_id          = "tf-test-dms-replication-subnet-group-%[1]s"
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.dms_subnet_1.id, aws_subnet.dms_subnet_2.id]
}

resource "aws_dms_replication_instance" "dms_replication_instance" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.c4.large"
  replication_instance_id      = "tf-test-dms-replication-instance-%[1]s"
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.dms_replication_subnet_group.replication_subnet_group_id
}

resource "aws_dms_replication_task" "dms_replication_task" {
  migration_type            = "full-load"
  replication_instance_arn  = aws_dms_replication_instance.dms_replication_instance.replication_instance_arn
  replication_task_id       = "tf-test-dms-replication-task-%[1]s"
  replication_task_settings = "{\"BeforeImageSettings\":null,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false}}"
  source_endpoint_arn       = aws_dms_endpoint.dms_endpoint_source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  tags = {
    Name   = "tf-test-dms-replication-task-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  target_endpoint_arn = aws_dms_endpoint.dms_endpoint_target.endpoint_arn
}
`, randId))
}

func dmsReplicationTaskConfigUpdate(randId string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "dms_assume_role_policy_document" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      identifiers = ["dms.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "dms_vpc_role" {
  assume_role_policy = data.aws_iam_policy_document.dms_assume_role_policy_document.json
  name = "dms-vpc-role"
}

resource "aws_iam_role" "dms_cloudwatch_logs_role" {
  assume_role_policy = data.aws_iam_policy_document.dms_assume_role_policy_document.json
  name = "dms-cloudwatch-logs-role"
}

resource "aws_iam_role_policy_attachment" "dms_vpc_access_policy" {
  role       = aws_iam_role.dms_vpc_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDMSVPCManagementRole"
}

resource "aws_iam_role_policy_attachment" "dms_cloudwatch_logs_access_policy" {
  role       = aws_iam_role.dms_cloudwatch_logs_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonDMSCloudWatchLogsRole"
}

resource "aws_vpc" "dms_vpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-dms-replication-task-%[1]s"
  }

  depends_on = [
    aws_iam_role_policy_attachment.dms_vpc_access_policy,
    aws_iam_role_policy_attachment.dms_cloudwatch_logs_access_policy,
  ]
}

resource "aws_subnet" "dms_subnet_1" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-task-1-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_subnet" "dms_subnet_2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.dms_vpc.id

  tags = {
    Name = "tf-acc-dms-replication-task-2-%[1]s"
  }

  depends_on = [aws_vpc.dms_vpc]
}

resource "aws_dms_endpoint" "dms_endpoint_source" {
  database_name = "tf-test-dms-db"
  endpoint_id   = "tf-test-dms-endpoint-source-%[1]s"
  endpoint_type = "source"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_endpoint" "dms_endpoint_target" {
  database_name = "tf-test-dms-db"
  endpoint_id   = "tf-test-dms-endpoint-target-%[1]s"
  endpoint_type = "target"
  engine_name   = "aurora"
  server_name   = "tf-test-cluster.cluster-xxxxxxx.${data.aws_region.current.name}.rds.${data.aws_partition.current.dns_suffix}"
  port          = 3306
  username      = "tftest"
  password      = "tftest"
}

resource "aws_dms_replication_subnet_group" "dms_replication_subnet_group" {
  replication_subnet_group_id          = "tf-test-dms-replication-subnet-group-%[1]s"
  replication_subnet_group_description = "terraform test for replication subnet group"
  subnet_ids                           = [aws_subnet.dms_subnet_1.id, aws_subnet.dms_subnet_2.id]
}

resource "aws_dms_replication_instance" "dms_replication_instance" {
  allocated_storage            = 5
  auto_minor_version_upgrade   = true
  replication_instance_class   = "dms.c4.large"
  replication_instance_id      = "tf-test-dms-replication-instance-%[1]s"
  preferred_maintenance_window = "sun:00:30-sun:02:30"
  publicly_accessible          = false
  replication_subnet_group_id  = aws_dms_replication_subnet_group.dms_replication_subnet_group.replication_subnet_group_id
}

resource "aws_dms_replication_task" "dms_replication_task" {
  migration_type            = "full-load"
  replication_instance_arn  = aws_dms_replication_instance.dms_replication_instance.replication_instance_arn
  replication_task_id       = "tf-test-dms-replication-task-%[1]s"
  replication_task_settings = "{\"BeforeImageSettings\":null,\"ChangeProcessingDdlHandlingPolicy\":{\"HandleSourceTableAltered\":true,\"HandleSourceTableDropped\":true,\"HandleSourceTableTruncated\":true},\"ChangeProcessingTuning\":{\"BatchApplyMemoryLimit\":500,\"BatchApplyPreserveTransaction\":true,\"BatchApplyTimeoutMax\":30,\"BatchApplyTimeoutMin\":1,\"BatchSplitSize\":0,\"CommitTimeout\":1,\"MemoryKeepTime\":60,\"MemoryLimitTotal\":1024,\"MinTransactionSize\":1000,\"StatementCacheSize\":50},\"CharacterSetSettings\":null,\"ControlTablesSettings\":{\"ControlSchema\":\"\",\"HistoryTableEnabled\":false,\"HistoryTimeslotInMinutes\":5,\"StatusTableEnabled\":false,\"SuspendedTablesTableEnabled\":false},\"ErrorBehavior\":{\"ApplyErrorDeletePolicy\":\"IGNORE_RECORD\",\"ApplyErrorEscalationCount\":0,\"ApplyErrorEscalationPolicy\":\"LOG_ERROR\",\"ApplyErrorFailOnTruncationDdl\":false,\"ApplyErrorInsertPolicy\":\"LOG_ERROR\",\"ApplyErrorUpdatePolicy\":\"LOG_ERROR\",\"DataErrorEscalationCount\":0,\"DataErrorEscalationPolicy\":\"SUSPEND_TABLE\",\"DataErrorPolicy\":\"LOG_ERROR\",\"DataTruncationErrorPolicy\":\"LOG_ERROR\",\"FailOnNoTablesCaptured\":false,\"FailOnTransactionConsistencyBreached\":false,\"FullLoadIgnoreConflicts\":true,\"RecoverableErrorCount\":-1,\"RecoverableErrorInterval\":5,\"RecoverableErrorThrottling\":true,\"RecoverableErrorThrottlingMax\":1800,\"TableErrorEscalationCount\":0,\"TableErrorEscalationPolicy\":\"STOP_TASK\",\"TableErrorPolicy\":\"SUSPEND_TABLE\"},\"FullLoadSettings\":{\"CommitRate\":10000,\"CreatePkAfterFullLoad\":false,\"MaxFullLoadSubTasks\":8,\"StopTaskCachedChangesApplied\":false,\"StopTaskCachedChangesNotApplied\":false,\"TargetTablePrepMode\":\"DROP_AND_CREATE\",\"TransactionConsistencyTimeout\":600},\"Logging\":{\"EnableLogging\":false,\"LogComponents\":[{\"Id\":\"TRANSFORMATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_UNLOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"IO\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_LOAD\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"PERFORMANCE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SOURCE_CAPTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"SORTER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"REST_SERVER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"VALIDATOR_EXT\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TARGET_APPLY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TASK_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"TABLES_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"METADATA_MANAGER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_FACTORY\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMON\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"ADDONS\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"DATA_STRUCTURE\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"COMMUNICATION\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"},{\"Id\":\"FILE_TRANSFER\",\"Severity\":\"LOGGER_SEVERITY_DEFAULT\"}]},\"LoopbackPreventionSettings\":null,\"PostProcessingRules\":null,\"StreamBufferSettings\":{\"CtrlStreamBufferSizeInMB\":5,\"StreamBufferCount\":3,\"StreamBufferSizeInMB\":8},\"TargetMetadata\":{\"BatchApplyEnabled\":false,\"FullLobMode\":false,\"InlineLobMaxSize\":0,\"LimitedSizeLobMode\":true,\"LoadMaxFileSize\":0,\"LobChunkSize\":0,\"LobMaxSize\":32,\"ParallelApplyBufferSize\":0,\"ParallelApplyQueuesPerThread\":0,\"ParallelApplyThreads\":0,\"ParallelLoadBufferSize\":0,\"ParallelLoadQueuesPerThread\":0,\"ParallelLoadThreads\":0,\"SupportLobs\":true,\"TargetSchema\":\"\",\"TaskRecoveryTableEnabled\":false}}"
  source_endpoint_arn       = aws_dms_endpoint.dms_endpoint_source.endpoint_arn
  table_mappings            = "{\"rules\":[{\"rule-type\":\"selection\",\"rule-id\":\"1\",\"rule-name\":\"1\",\"object-locator\":{\"schema-name\":\"%%\",\"table-name\":\"%%\"},\"rule-action\":\"include\"}]}"

  tags = {
    Name   = "tf-test-dms-replication-task-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  target_endpoint_arn = aws_dms_endpoint.dms_endpoint_target.endpoint_arn
}
`, randId))
}
