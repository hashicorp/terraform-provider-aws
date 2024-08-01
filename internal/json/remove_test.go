// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestRemoveFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "empty JSON",
			input:    "{}",
			want:     "{}",
		},
		{
			testName: "single field",
			input:    `{ "key": 42 }`,
			want:     `{"key":42}`,
		},
		{
			testName: "with read-only field",
			input:    "{\"unifiedAlerting\": {\"enabled\": true}, \"plugins\": {\"pluginAdminEnabled\" :false}}",
			want:     "{\"unifiedAlerting\":{\"enabled\":true}}",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := json.RemoveFields(testCase.input, `"plugins"`), testCase.want; got != want {
				t.Errorf("RemoveReadOnlyFields(%q) = %q, want %q", testCase.input, got, want)
			}
		})
	}
}

func TestRemoveEmptyFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "empty JSON",
			input:    "{}",
			want:     "{}",
		},
		{
			testName: "single non-empty simple field",
			input:    `{"key": 42}`,
			want:     `{"key":42}`,
		},
		{
			testName: "single non-empty array field",
			input:    `{"key": [1, true, "answer"]}`,
			want:     `{"key":[1,true,"answer"]}`,
		},
		{
			testName: "single non-empty object field",
			input:    `{"key": {"inner": true}}`,
			want:     `{"key":{"inner":true}}`,
		},
		{
			testName: "single null field",
			input:    `{"key": null}`,
			want:     `{}`,
		},
		{
			testName: "single empty array field",
			input:    `{"key": []}`,
			want:     `{}`,
		},
		{
			testName: "single empty object field",
			input:    `{"key": {}}`,
			want:     `{}`,
		},
		{
			testName: "empty fields deeply nested 1 pass",
			input:    `{"key": {"a": [1, 2], "b": [], "c": {"d": true, "e": null}}}`,
			want:     `{"key":{"a":[1,2],"c":{"d":true}}}`,
		},
		{
			testName: "empty fields deeply nested 2 passes",
			input:    `{"key": {"a": [1, 2], "b": {}, "c": {"d": null}}}`,
			want:     `{"key":{"a":[1,2]}}`,
		},
		{
			testName: "empty fields deeply nested many empty objects",
			input:    `{"key": {"a": [1, 2], "b": {}, "c": {"d": {}}, "e": {}, "f": 99}}`,
			want:     `{"key":{"a":[1,2],"f":99}}`,
		},
		{
			testName: "empty fields nested empty arrays",
			input:    `{"key": {"a": [1, [2], [], [[]], 3]}}`,
			want:     `{"key":{"a":[1,[2],3]}}`,
		},
		{
			testName: "real life example",
			input:    `{"TargetMetadata":{"SupportLobs":true,"LimitedSizeLobMode":true,"LobMaxSize":32},"FullLoadSettings":{"TargetTablePrepMode":"DROP_AND_CREATE","MaxFullLoadSubTasks":8,"TransactionConsistencyTimeout":600,"CommitRate":10000},"TTSettings":{"TTS3Settings":{},"TTRecordSettings":{}},"Logging":{},"ControlTablesSettings":{"HistoryTimeslotInMinutes":5},"StreamBufferSettings":{"StreamBufferCount":3,"StreamBufferSizeInMB":8},"ChangeProcessingTuning":{"BatchApplyPreserveTransaction":true,"BatchApplyTimeoutMin":1,"BatchApplyTimeoutMax":30,"BatchApplyMemoryLimit":500,"MinTransactionSize":1000,"CommitTimeout":1,"MemoryLimitTotal":1024,"MemoryKeepTime":60,"StatementCacheSize":50},"ChangeProcessingDdlHandlingPolicy":{"HandleSourceTableDropped":true,"HandleSourceTableTruncated":true,"HandleSourceTableAltered":true},"LoopbackPreventionSettings":{},"CharacterSetSettings":{"CharacterSetSupport":{}},"BeforeImageSettings":{},"ErrorBehavior":{"DataErrorPolicy":"LOG_ERROR","DataTruncationErrorPolicy":"LOG_ERROR","DataErrorEscalationPolicy":"SUSPEND_TABLE","TableErrorPolicy":"SUSPEND_TABLE","TableErrorEscalationPolicy":"STOP_TASK","RecoverableErrorCount":-1,"RecoverableErrorInterval":5,"RecoverableErrorThrottling":true,"RecoverableErrorThrottlingMax":1800,"ApplyErrorDeletePolicy":"IGNORE_RECORD","ApplyErrorInsertPolicy":"LOG_ERROR","ApplyErrorUpdatePolicy":"LOG_ERROR","ApplyErrorEscalationPolicy":"LOG_ERROR","FullLoadIgnoreConflicts":true},"ValidationSettings":{"ValidationMode":"ROW_LEVEL","ThreadCount":5,"PartitionSize":10000,"FailureMaxCount":10000,"TableFailureMaxCount":1000}}`,
			want:     `{"TargetMetadata":{"SupportLobs":true,"LimitedSizeLobMode":true,"LobMaxSize":32},"FullLoadSettings":{"TargetTablePrepMode":"DROP_AND_CREATE","MaxFullLoadSubTasks":8,"TransactionConsistencyTimeout":600,"CommitRate":10000},"ControlTablesSettings":{"HistoryTimeslotInMinutes":5},"StreamBufferSettings":{"StreamBufferCount":3,"StreamBufferSizeInMB":8},"ChangeProcessingTuning":{"BatchApplyPreserveTransaction":true,"BatchApplyTimeoutMin":1,"BatchApplyTimeoutMax":30,"BatchApplyMemoryLimit":500,"MinTransactionSize":1000,"CommitTimeout":1,"MemoryLimitTotal":1024,"MemoryKeepTime":60,"StatementCacheSize":50},"ChangeProcessingDdlHandlingPolicy":{"HandleSourceTableDropped":true,"HandleSourceTableTruncated":true,"HandleSourceTableAltered":true},"ErrorBehavior":{"DataErrorPolicy":"LOG_ERROR","DataTruncationErrorPolicy":"LOG_ERROR","DataErrorEscalationPolicy":"SUSPEND_TABLE","TableErrorPolicy":"SUSPEND_TABLE","TableErrorEscalationPolicy":"STOP_TASK","RecoverableErrorCount":-1,"RecoverableErrorInterval":5,"RecoverableErrorThrottling":true,"RecoverableErrorThrottlingMax":1800,"ApplyErrorDeletePolicy":"IGNORE_RECORD","ApplyErrorInsertPolicy":"LOG_ERROR","ApplyErrorUpdatePolicy":"LOG_ERROR","ApplyErrorEscalationPolicy":"LOG_ERROR","FullLoadIgnoreConflicts":true},"ValidationSettings":{"ValidationMode":"ROW_LEVEL","ThreadCount":5,"PartitionSize":10000,"FailureMaxCount":10000,"TableFailureMaxCount":1000}}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := json.RemoveEmptyFields([]byte(testCase.input)), testCase.want; string(got) != want {
				t.Errorf("RemoveEmptyFields(%q) = %q, want %q", testCase.input, got, want)
			}
		})
	}
}
