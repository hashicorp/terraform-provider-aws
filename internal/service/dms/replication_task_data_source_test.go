package dms_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
//
// Remember to register this new data source in the provider
// (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// know about it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/dms/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"fmt"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
)

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the data source as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// data source name.
//
// Acceptance test access AWS and cost money to run.
func TestAccDMSReplicationTaskDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_replication_task.test"
	dataSourceName := "data.aws_dms_replication_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationTaskDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationTaskExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_task_id", resourceName, "replication_task_id"),
				),
			},
		},
	})
}

func testAccReplicationTaskDataSourceConfig_basic(rName string) string {
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

  target_endpoint_arn = aws_dms_endpoint.target.endpoint_arn
}

data "aws_dms_replication_task" "test" {
	replication_task_id = aws_dms_replication_task.test.replication_task_id
}
`, rName))
}
