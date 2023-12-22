// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Tasks.CustomizingTasks.TaskSettings.html#CHAP_Tasks.CustomizingTasks.TaskSettings.Example
// https://mholt.github.io/json-to-go/

type taskSettings struct {
	TargetMetadata struct {
		TargetSchema                 string `json:"TargetSchema,omitempty"`
		SupportLobs                  bool   `json:"SupportLobs,omitempty"`
		FullLobMode                  bool   `json:"FullLobMode,omitempty"`
		LobChunkSize                 int    `json:"LobChunkSize,omitempty"`
		LimitedSizeLobMode           bool   `json:"LimitedSizeLobMode,omitempty"`
		LobMaxSize                   int    `json:"LobMaxSize,omitempty"`
		InlineLobMaxSize             int    `json:"InlineLobMaxSize,omitempty"`
		LoadMaxFileSize              int    `json:"LoadMaxFileSize,omitempty"`
		ParallelLoadThreads          int    `json:"ParallelLoadThreads,omitempty"`
		ParallelLoadBufferSize       int    `json:"ParallelLoadBufferSize,omitempty"`
		ParallelLoadQueuesPerThread  int    `json:"ParallelLoadQueuesPerThread,omitempty"`
		ParallelApplyThreads         int    `json:"ParallelApplyThreads,omitempty"`
		ParallelApplyBufferSize      int    `json:"ParallelApplyBufferSize,omitempty"`
		ParallelApplyQueuesPerThread int    `json:"ParallelApplyQueuesPerThread,omitempty"`
		BatchApplyEnabled            bool   `json:"BatchApplyEnabled,omitempty"`
		TaskRecoveryTableEnabled     bool   `json:"TaskRecoveryTableEnabled,omitempty"`
	} `json:"TargetMetadata,omitempty"`
	FullLoadSettings struct {
		TargetTablePrepMode             string `json:"TargetTablePrepMode,omitempty"`
		CreatePkAfterFullLoad           bool   `json:"CreatePkAfterFullLoad,omitempty"`
		StopTaskCachedChangesApplied    bool   `json:"StopTaskCachedChangesApplied,omitempty"`
		StopTaskCachedChangesNotApplied bool   `json:"StopTaskCachedChangesNotApplied,omitempty"`
		MaxFullLoadSubTasks             int    `json:"MaxFullLoadSubTasks,omitempty"`
		TransactionConsistencyTimeout   int    `json:"TransactionConsistencyTimeout,omitempty"`
		CommitRate                      int    `json:"CommitRate,omitempty"`
	} `json:"FullLoadSettings,omitempty"`
	TTSettings struct {
		EnableTT     bool `json:"EnableTT,omitempty"`
		TTS3Settings struct {
			EncryptionMode                   string `json:"EncryptionMode,omitempty"`
			ServerSideEncryptionKmsKeyID     string `json:"ServerSideEncryptionKmsKeyId,omitempty"`
			ServiceAccessRoleArn             string `json:"ServiceAccessRoleArn,omitempty"`
			BucketName                       string `json:"BucketName,omitempty"`
			BucketFolder                     string `json:"BucketFolder,omitempty"`
			EnableDeletingFromS3OnTaskDelete bool   `json:"EnableDeletingFromS3OnTaskDelete,omitempty"`
		} `json:"TTS3Settings,omitempty"`
		TTRecordSettings struct {
			EnableRawData   bool   `json:"EnableRawData,omitempty"`
			OperationsToLog string `json:"OperationsToLog,omitempty"`
			MaxRecordSize   int    `json:"MaxRecordSize,omitempty"`
		} `json:"TTRecordSettings,omitempty"`
	} `json:"TTSettings,omitempty"`
	Logging struct {
		EnableLogging bool `json:"EnableLogging,omitempty"`
	} `json:"Logging,omitempty"`
	ControlTablesSettings struct {
		ControlSchema               string `json:"ControlSchema,omitempty"`
		HistoryTimeslotInMinutes    int    `json:"HistoryTimeslotInMinutes,omitempty"`
		HistoryTableEnabled         bool   `json:"HistoryTableEnabled,omitempty"`
		SuspendedTablesTableEnabled bool   `json:"SuspendedTablesTableEnabled,omitempty"`
		StatusTableEnabled          bool   `json:"StatusTableEnabled,omitempty"`
	} `json:"ControlTablesSettings,omitempty"`
	StreamBufferSettings struct {
		StreamBufferCount    int `json:"StreamBufferCount,omitempty"`
		StreamBufferSizeInMB int `json:"StreamBufferSizeInMB,omitempty"`
	} `json:"StreamBufferSettings,omitempty"`
	ChangeProcessingTuning struct {
		BatchApplyPreserveTransaction bool `json:"BatchApplyPreserveTransaction,omitempty"`
		BatchApplyTimeoutMin          int  `json:"BatchApplyTimeoutMin,omitempty"`
		BatchApplyTimeoutMax          int  `json:"BatchApplyTimeoutMax,omitempty"`
		BatchApplyMemoryLimit         int  `json:"BatchApplyMemoryLimit,omitempty"`
		BatchSplitSize                int  `json:"BatchSplitSize,omitempty"`
		MinTransactionSize            int  `json:"MinTransactionSize,omitempty"`
		CommitTimeout                 int  `json:"CommitTimeout,omitempty"`
		MemoryLimitTotal              int  `json:"MemoryLimitTotal,omitempty"`
		MemoryKeepTime                int  `json:"MemoryKeepTime,omitempty"`
		StatementCacheSize            int  `json:"StatementCacheSize,omitempty"`
	} `json:"ChangeProcessingTuning,omitempty"`
	ChangeProcessingDdlHandlingPolicy struct {
		HandleSourceTableDropped   bool `json:"HandleSourceTableDropped,omitempty"`
		HandleSourceTableTruncated bool `json:"HandleSourceTableTruncated,omitempty"`
		HandleSourceTableAltered   bool `json:"HandleSourceTableAltered,omitempty"`
	} `json:"ChangeProcessingDdlHandlingPolicy,omitempty"`
	LoopbackPreventionSettings struct {
		EnableLoopbackPrevention bool   `json:"EnableLoopbackPrevention,omitempty"`
		SourceSchema             string `json:"SourceSchema,omitempty"`
		TargetSchema             string `json:"TargetSchema,omitempty"`
	} `json:"LoopbackPreventionSettings,omitempty"`
	CharacterSetSettings struct {
		CharacterReplacements []struct {
			SourceCharacterCodePoint int `json:"SourceCharacterCodePoint,omitempty"`
			TargetCharacterCodePoint int `json:"TargetCharacterCodePoint,omitempty"`
		} `json:"CharacterReplacements,omitempty"`
		CharacterSetSupport struct {
			CharacterSet                  string `json:"CharacterSet,omitempty"`
			ReplaceWithCharacterCodePoint int    `json:"ReplaceWithCharacterCodePoint,omitempty"`
		} `json:"CharacterSetSupport,omitempty"`
	} `json:"CharacterSetSettings,omitempty"`
	BeforeImageSettings struct {
		EnableBeforeImage bool   `json:"EnableBeforeImage,omitempty"`
		FieldName         string `json:"FieldName,omitempty"`
		ColumnFilter      string `json:"ColumnFilter,omitempty"`
	} `json:"BeforeImageSettings,omitempty"`
	ErrorBehavior struct {
		DataErrorPolicy               string `json:"DataErrorPolicy,omitempty"`
		DataTruncationErrorPolicy     string `json:"DataTruncationErrorPolicy,omitempty"`
		DataErrorEscalationPolicy     string `json:"DataErrorEscalationPolicy,omitempty"`
		DataErrorEscalationCount      int    `json:"DataErrorEscalationCount,omitempty"`
		TableErrorPolicy              string `json:"TableErrorPolicy,omitempty"`
		TableErrorEscalationPolicy    string `json:"TableErrorEscalationPolicy,omitempty"`
		TableErrorEscalationCount     int    `json:"TableErrorEscalationCount,omitempty"`
		RecoverableErrorCount         int    `json:"RecoverableErrorCount,omitempty"`
		RecoverableErrorInterval      int    `json:"RecoverableErrorInterval,omitempty"`
		RecoverableErrorThrottling    bool   `json:"RecoverableErrorThrottling,omitempty"`
		RecoverableErrorThrottlingMax int    `json:"RecoverableErrorThrottlingMax,omitempty"`
		ApplyErrorDeletePolicy        string `json:"ApplyErrorDeletePolicy,omitempty"`
		ApplyErrorInsertPolicy        string `json:"ApplyErrorInsertPolicy,omitempty"`
		ApplyErrorUpdatePolicy        string `json:"ApplyErrorUpdatePolicy,omitempty"`
		ApplyErrorEscalationPolicy    string `json:"ApplyErrorEscalationPolicy,omitempty"`
		ApplyErrorEscalationCount     int    `json:"ApplyErrorEscalationCount,omitempty"`
		FullLoadIgnoreConflicts       bool   `json:"FullLoadIgnoreConflicts,omitempty"`
	} `json:"ErrorBehavior,omitempty"`
	ValidationSettings struct {
		EnableValidation                 bool   `json:"EnableValidation,omitempty"`
		ValidationMode                   string `json:"ValidationMode,omitempty"`
		ThreadCount                      int    `json:"ThreadCount,omitempty"`
		PartitionSize                    int    `json:"PartitionSize,omitempty"`
		FailureMaxCount                  int    `json:"FailureMaxCount,omitempty"`
		RecordFailureDelayInMinutes      int    `json:"RecordFailureDelayInMinutes,omitempty"`
		RecordSuspendDelayInMinutes      int    `json:"RecordSuspendDelayInMinutes,omitempty"`
		MaxKeyColumnSize                 int    `json:"MaxKeyColumnSize,omitempty"`
		TableFailureMaxCount             int    `json:"TableFailureMaxCount,omitempty"`
		ValidationOnly                   bool   `json:"ValidationOnly,omitempty"`
		HandleCollationDiff              bool   `json:"HandleCollationDiff,omitempty"`
		RecordFailureDelayLimitInMinutes int    `json:"RecordFailureDelayLimitInMinutes,omitempty"`
		SkipLobColumns                   bool   `json:"SkipLobColumns,omitempty"`
		ValidationPartialLobSize         int    `json:"ValidationPartialLobSize,omitempty"`
		ValidationQueryCdcDelaySeconds   int    `json:"ValidationQueryCdcDelaySeconds,omitempty"`
	} `json:"ValidationSettings,omitempty"`
}

// normalizeTaskSettings returns a normalized DMS task settings JSON string.
// Read-only (non-configurable) fields are removed by using the published "schema".
// Empty fields are then removed.
func normalizeTaskSettings(apiObject string) string {
	var taskSettings taskSettings

	if err := json.Unmarshal([]byte(apiObject), &taskSettings); err != nil {
		return apiObject
	}

	if b, err := json.Marshal(&taskSettings); err != nil {
		return apiObject
	} else {
		return string(tfjson.RemoveEmptyFields(b))
	}
}

// suppressEquivalentTaskSettings provides custom difference suppression for task settings.
func suppressEquivalentTaskSettings(k, old, new string, d *schema.ResourceData) bool {
	if !json.Valid([]byte(old)) || !json.Valid([]byte(new)) {
		return old == new
	}

	old, new = normalizeTaskSettings(old), normalizeTaskSettings(new)

	return verify.JSONStringsEqual(old, new)
}
