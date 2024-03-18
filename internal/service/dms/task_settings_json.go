// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"cmp"
	"encoding/json"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"golang.org/x/exp/slices"
)

// https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Tasks.CustomizingTasks.TaskSettings.html#CHAP_Tasks.CustomizingTasks.TaskSettings.Example
// https://mholt.github.io/json-to-go/

type taskSettings struct {
	BeforeImageSettings struct {
		EnableBeforeImage bool   `json:"EnableBeforeImage,omitempty"`
		FieldName         string `json:"FieldName,omitempty"`
		ColumnFilter      string `json:"ColumnFilter,omitempty"`
	} `json:"BeforeImageSettings,omitempty"`
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
	ControlTablesSettings struct {
		ControlSchema                 string `json:"ControlSchema,omitempty"`
		CommitPositionTableEnabled    bool   `json:"CommitPositionTableEnabled,omitempty"`
		FullLoadExceptionTableEnabled bool   `json:"FullLoadExceptionTableEnabled,omitempty"`
		HistoryTimeslotInMinutes      int    `json:"HistoryTimeslotInMinutes,omitempty"`
		HistoryTableEnabled           bool   `json:"HistoryTableEnabled,omitempty"`
		SuspendedTablesTableEnabled   bool   `json:"SuspendedTablesTableEnabled,omitempty"`
		StatusTableEnabled            bool   `json:"StatusTableEnabled,omitempty"`
		// historyTimeslotInMinutes      int    `json:"historyTimeslotInMinutes,omitempty"`
	} `json:"ControlTablesSettings,omitempty"`
	ErrorBehavior struct {
		DataErrorPolicy                             string `json:"DataErrorPolicy,omitempty"`
		DataTruncationErrorPolicy                   string `json:"DataTruncationErrorPolicy,omitempty"`
		DataErrorEscalationPolicy                   string `json:"DataErrorEscalationPolicy,omitempty"`
		DataErrorEscalationCount                    int    `json:"DataErrorEscalationCount,omitempty"`
		EventErrorPolicy                            string `json:"EventErrorPolicy,omitempty"`
		FailOnNoTablesCaptured                      bool   `json:"FailOnNoTablesCaptured,omitempty"`
		FailOnTransactionConsistencyBreached        bool   `json:"FailOnTransactionConsistencyBreached,omitempty"`
		TableErrorPolicy                            string `json:"TableErrorPolicy,omitempty"`
		TableErrorEscalationPolicy                  string `json:"TableErrorEscalationPolicy,omitempty"`
		TableErrorEscalationCount                   int    `json:"TableErrorEscalationCount,omitempty"`
		RecoverableErrorCount                       int    `json:"RecoverableErrorCount,omitempty"`
		RecoverableErrorInterval                    int    `json:"RecoverableErrorInterval,omitempty"`
		RecoverableErrorThrottling                  bool   `json:"RecoverableErrorThrottling,omitempty"`
		RecoverableErrorThrottlingMax               int    `json:"RecoverableErrorThrottlingMax,omitempty"`
		RecoverableErrorStopRetryAfterThrottlingMax bool   `json:"RecoverableErrorStopRetryAfterThrottlingMax,omitempty"`
		ApplyErrorDeletePolicy                      string `json:"ApplyErrorDeletePolicy,omitempty"`
		ApplyErrorInsertPolicy                      string `json:"ApplyErrorInsertPolicy,omitempty"`
		ApplyErrorUpdatePolicy                      string `json:"ApplyErrorUpdatePolicy,omitempty"`
		ApplyErrorEscalationPolicy                  string `json:"ApplyErrorEscalationPolicy,omitempty"`
		ApplyErrorEscalationCount                   int    `json:"ApplyErrorEscalationCount,omitempty"`
		ApplyErrorFailOnTruncationDdl               bool   `json:"ApplyErrorFailOnTruncationDdl,omitempty"`
		FullLoadIgnoreConflicts                     bool   `json:"FullLoadIgnoreConflicts,omitempty"`
	} `json:"ErrorBehavior,omitempty"`
	FailTaskWhenCleanTaskResourceFailed bool `json:"FailTaskWhenCleanTaskResourceFailed,omitempty"`
	FullLoadSettings                    struct {
		TargetTablePrepMode             string `json:"TargetTablePrepMode,omitempty"`
		CreatePkAfterFullLoad           bool   `json:"CreatePkAfterFullLoad,omitempty"`
		StopTaskCachedChangesApplied    bool   `json:"StopTaskCachedChangesApplied,omitempty"`
		StopTaskCachedChangesNotApplied bool   `json:"StopTaskCachedChangesNotApplied,omitempty"`
		MaxFullLoadSubTasks             int    `json:"MaxFullLoadSubTasks,omitempty"`
		TransactionConsistencyTimeout   int    `json:"TransactionConsistencyTimeout,omitempty"`
		CommitRate                      int    `json:"CommitRate,omitempty"`
	} `json:"FullLoadSettings,omitempty"`
	Logging struct {
		EnableLogging    bool `json:"EnableLogging,omitempty"`
		EnableLogContext bool `json:"EnableLogContext,omitempty"`
		// CloudWatchLogGroup  struct{} `json:"CloudWatchLogGroup,omitempty"`
		// CloudWatchLogStream struct{} `json:"CloudWatchLogStream,omitempty"`
	} `json:"Logging,omitempty"`
	LoopbackPreventionSettings struct {
		EnableLoopbackPrevention bool   `json:"EnableLoopbackPrevention,omitempty"`
		SourceSchema             string `json:"SourceSchema,omitempty"`
		TargetSchema             string `json:"TargetSchema,omitempty"`
	} `json:"LoopbackPreventionSettings,omitempty"`
	PostProcessingRules  struct{} `json:"PostProcessingRules,omitempty"`
	StreamBufferSettings struct {
		CtrlStreamBufferSizeInMB int `json:"CtrlStreamBufferSizeInMB,omitempty"`
		StreamBufferCount        int `json:"StreamBufferCount,omitempty"`
		StreamBufferSizeInMB     int `json:"StreamBufferSizeInMB,omitempty"`
	} `json:"StreamBufferSettings,omitempty"`
	TTSettings struct {
		EnableTT            bool `json:"EnableTT,omitempty"`
		FailTaskOnTTFailure bool `json:"FailTaskOnTTFailure,omitempty"`
		TTS3Settings        struct {
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
	ChangeProcessingDdlHandlingPolicy struct {
		HandleSourceTableDropped   bool `json:"HandleSourceTableDropped,omitempty"`
		HandleSourceTableTruncated bool `json:"HandleSourceTableTruncated,omitempty"`
		HandleSourceTableAltered   bool `json:"HandleSourceTableAltered,omitempty"`
	} `json:"ChangeProcessingDdlHandlingPolicy,omitempty"`
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
	// var taskSettings taskSettings

	defaultValues := map[string]interface{}{
		"ChangeProcessingTuning": map[string]interface{}{
			"BatchApplyMemoryLimit":         500,
			"BatchApplyTimeoutMax":          30,
			"BatchApplyTimeoutMin":          1,
			"BatchSplitSize":                0,
			"CommitTimeout":                 1,
			"MemoryKeepTime":                60,
			"MemoryLimitTotal":              1024,
			"MinTransactionSize":            1000,
			"StatementCacheSize":            50,
			"BatchApplyPreserveTransaction": true,
		},
		"ControlTablesSettings": map[string]interface{}{
			"historyTimeslotInMinutes":      5,
			"CommitPositionTableEnabled":    false,
			"HistoryTimeslotInMinutes":      5,
			"StatusTableEnabled":            false,
			"SuspendedTablesTableEnabled":   false,
			"HistoryTableEnabled":           false,
			"ControlSchema":                 "",
			"FullLoadExceptionTableEnabled": false,
		},
		"BeforeImageSettings":                 nil,
		"FailTaskWhenCleanTaskResourceFailed": false,
		"ErrorBehavior": map[string]interface{}{
			"DataErrorPolicy":                             "LOG_ERROR",
			"DataTruncationErrorPolicy":                   "LOG_ERROR",
			"DataErrorEscalationPolicy":                   "SUSPEND_TABLE",
			"EventErrorPolicy":                            "IGNORE",
			"FailOnNoTablesCaptured":                      true,
			"TableErrorPolicy":                            "SUSPEND_TABLE",
			"TableErrorEscalationPolicy":                  "STOP_TASK",
			"RecoverableErrorCount":                       -1,
			"RecoverableErrorInterval":                    5,
			"RecoverableErrorThrottling":                  true,
			"RecoverableErrorThrottlingMax":               1800,
			"RecoverableErrorStopRetryAfterThrottlingMax": true,
			"ApplyErrorDeletePolicy":                      "IGNORE_RECORD",
			"ApplyErrorInsertPolicy":                      "LOG_ERROR",
			"ApplyErrorUpdatePolicy":                      "LOG_ERROR",
			"ApplyErrorEscalationPolicy":                  "LOG_ERROR",
			"FullLoadIgnoreConflicts":                     true,
			"ApplyErrorEscalationCount":                   0,
			"ApplyErrorFailOnTruncationDdl":               false,
			"DataErrorEscalationCount":                    0,
			"FailOnTransactionConsistencyBreached":        false,
			"TableErrorEscalationCount":                   0,
		},
		"TTSettings": map[string]interface{}{
			"TTS3Settings":        nil,
			"TTRecordSettings":    nil,
			"FailTaskOnTTFailure": false,
			"EnableTT":            false,
		},
		"FullLoadSettings": map[string]interface{}{
			"CommitRate":                      10000,
			"StopTaskCachedChangesApplied":    false,
			"StopTaskCachedChangesNotApplied": false,
			"MaxFullLoadSubTasks":             8,
			"TransactionConsistencyTimeout":   600,
			"CreatePkAfterFullLoad":           false,
			"TargetTablePrepMode":             "DO_NOTHING",
		},
		"Logging": map[string]interface{}{
			"EnableLogging":       true,
			"CloudWatchLogGroup":  nil,
			"CloudWatchLogStream": nil,
			"EnableLogContext":    false,
			"LogComponents": []map[string]string{
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "TRANSFORMATION",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "SOURCE_UNLOAD",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "IO",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "TARGET_LOAD",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "PERFORMANCE",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "SOURCE_CAPTURE",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "SORTER",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "REST_SERVER",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "VALIDATOR_EXT",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "TARGET_APPLY",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "TASK_MANAGER",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "TABLES_MANAGER",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "METADATA_MANAGER",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "FILE_FACTORY",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "COMMON",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "ADDONS",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "DATA_STRUCTURE",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "COMMUNICATION",
				},
				{
					"Severity": "LOGGER_SEVERITY_DEFAULT",
					"Id":       "FILE_TRANSFER",
				},
			},
		},
		"StreamBufferSettings": map[string]interface{}{
			"CtrlStreamBufferSizeInMB": 5,
			"StreamBufferCount":        3,
			"StreamBufferSizeInMB":     8,
		},
		"TargetMetadata": map[string]interface{}{
			"ParallelApplyBufferSize":      0,
			"ParallelApplyQueuesPerThread": 0,
			"ParallelApplyThreads":         0,
			"TargetSchema":                 "",
			"InlineLobMaxSize":             0,
			"ParallelLoadQueuesPerThread":  0,
			"SupportLobs":                  true,
			"LobChunkSize":                 64,
			"TaskRecoveryTableEnabled":     false,
			"ParallelLoadThreads":          0,
			"LobMaxSize":                   32,
			"BatchApplyEnabled":            false,
			"FullLobMode":                  false,
			"LimitedSizeLobMode":           true,
			"LoadMaxFileSize":              0,
			"ParallelLoadBufferSize":       0,
		},
		"ChangeProcessingDdlHandlingPolicy": map[string]interface{}{
			"HandleSourceTableDropped":   true,
			"HandleSourceTableTruncated": true,
			"HandleSourceTableAltered":   true,
		},
	}

	var jsonMap map[string]interface{}

	if err := json.Unmarshal([]byte(apiObject), &jsonMap); err != nil {
		log.Printf("[WARN] failed to unmarshal task settings JSON: %v", err)
		return apiObject
	}

	jsonMap = checkdefaultvalues(defaultValues, jsonMap)
	if b, err := json.Marshal(&jsonMap); err != nil {
		log.Printf("[WARN] failed to marshal task settings JSON: %v", err)
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
	log.Printf("[WARN] suppressEquivalentTaskSettings: old=%s, new=%s", old, new)

	return verify.JSONStringsEqual(old, new)
}

func checkdefaultvalues(defaultMap, oldMap map[string]interface{}) map[string]interface{} {
	for k, v := range oldMap {
		log.Printf("[WARN] checking key: %s, type: %T", k, v)
		if value, ok := defaultMap[k]; ok && v != nil {
			// Check the type of the value
			switch t := reflect.TypeOf(value); t.Kind() {
			// Check top level settings
			case reflect.Bool, reflect.String, reflect.Float64, reflect.Int:
				if reflect.DeepEqual(value, v) {
					delete(oldMap, k)
				}
			case reflect.Map:
				// Map of defaults
				kMap := value.(map[string]interface{})
				// Map of inner map (from user)
				vMap := v.(map[string]interface{})

				for kInner, vInner := range vMap {
					log.Printf("[WARN] default value type: %T, vinner type: %T", kMap[kInner], vInner)
					if kMap[kInner] != nil || vInner != nil {
						if reflect.TypeOf(vInner).Kind() == reflect.Float64 {
							if kMap[kInner] != nil {
								kMap[kInner] = float64(kMap[kInner].(int))
							}
						}
						if reflect.TypeOf(vInner).Kind() == reflect.Slice {
							temp := make([]map[string]string, 0)
							for _, v := range vInner.([]interface{}) {
								innerTemp := make(map[string]string)
								for k, v := range v.(map[string]interface{}) {
									innerTemp[k] = v.(string)
								}
								temp = append(temp, innerTemp)
							}
							// We are assuming the types; we know the type at the point of this code
							slices.SortFunc(temp, func(i, j map[string]string) int {
								return cmp.Compare(i["Id"], j["Id"])
							})
							vInner = temp

							slices.SortFunc(kMap[kInner].([]map[string]string), func(i, j map[string]string) int {
								return cmp.Compare(i["Id"], j["Id"])
							})
						}

						if reflect.DeepEqual(kMap[kInner], vInner) {
							delete(vMap, kInner)
						}
					}
				}
				if len(vMap) == 0 {
					delete(oldMap, k)
				}

			default:
				return oldMap
			}
		}
	}
	return oldMap
}
