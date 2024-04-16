// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"cmp"
	"encoding/json"
	"log"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Tasks.CustomizingTasks.TaskSettings.html#CHAP_Tasks.CustomizingTasks.TaskSettings.Example
// https://mholt.github.io/json-to-go/

// normalizeTaskSettings returns a normalized DMS task settings JSON string.
// Read-only (non-configurable) fields are removed by using the published "schema".
// Empty fields are then removed.
func normalizeTaskSettings(apiObject string) string {
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
		log.Printf("[DEBUG] failed to unmarshal task settings JSON: %v", err)
		return apiObject
	}

	jsonMap = checkdefaultvalues(defaultValues, jsonMap)
	if b, err := json.Marshal(&jsonMap); err != nil {
		log.Printf("[DEBUG] failed to marshal task settings JSON: %v", err)
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

func checkdefaultvalues(defaultMap, oldMap map[string]interface{}) map[string]interface{} {
	for k, v := range oldMap {
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
