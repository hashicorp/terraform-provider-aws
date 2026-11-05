// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"bytes"
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestRemoveFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		fields   []string
		want     string
	}{
		{
			testName: "empty object",
			input:    "{}",
			want:     "{}",
		},
		{
			testName: "single field",
			input:    `{ "key": 42 }`,
			want:     `{"key":42}`,
		},
		{
			testName: "remove object field",
			input:    `{"unifiedAlerting":{"enabled":true},"plugins":{"pluginAdminEnabled":false}}`,
			fields:   []string{`"plugins"`},
			want:     `{"unifiedAlerting":{"enabled":true}}`,
		},
		{
			testName: "remove nested fields and their values",
			input:    `{"remove":{"nested":[1,2]},"keep":{"remove":true,"value":1}}`,
			fields:   []string{`"remove"`},
			want:     `{"keep":{"value":1}}`,
		},
		{
			testName: "remove duplicate fields",
			input:    `{"remove":1,"keep":2,"remove":3}`,
			fields:   []string{`"remove"`},
			want:     `{"keep":2}`,
		},
		{
			testName: "compare raw field names",
			input:    `{"pl\u0075gins":1,"plugins":2}`,
			fields:   []string{`"plugins"`},
			want:     `{"pl\u0075gins":1}`,
		},
		{
			testName: "preserve top-level scalar",
			input:    ` 1e+02 `,
			want:     `1e+02`,
		},
		{
			testName: "invalid JSON",
			input:    `{"key":`,
			fields:   []string{`"key"`},
			want:     "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got := tfjson.RemoveFields(testCase.input, testCase.fields...); got != testCase.want {
				t.Errorf("RemoveFields(%q) = %q, want %q", testCase.input, got, testCase.want)
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
			testName: "empty object",
			input:    `{}`,
			want:     `{}`,
		},
		{
			testName: "empty array",
			input:    `[]`,
			want:     `[]`,
		},
		{
			testName: "top-level null",
			input:    `null`,
			want:     `null`,
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
			testName: "single empty string field",
			input:    `{"key":""}`,
			want:     `{"key":""}`,
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
			testName: "empty fields deeply nested",
			input:    `{"key":{"a":[1,2],"b":{},"c":{"d":null},"e":{},"f":99}}`,
			want:     `{"key":{"a":[1,2],"f":99}}`,
		},
		{
			testName: "empty values in nested arrays",
			input:    `{"key":[null,[],[[]],[1,[],null],{},[{"empty":null}],2]}`,
			want:     `{"key":[[1],2]}`,
		},
		{
			testName: "preserve duplicate fields and raw values",
			input:    `{"key":1e+02,"key":"a b","empty":null}`,
			want:     `{"key":1e+02,"key":"a b"}`,
		},
		{
			testName: "preserve raw string escapes",
			input:    `{"key":"\u003cscript\u003e","empty":null}`,
			want:     `{"key":"\u003cscript\u003e"}`,
		},
		{
			testName: "invalid JSON",
			input:    `{"key":`,
			want:     "",
		},
		{
			testName: "multiple top-level values",
			input:    `{} {}`,
			want:     "",
		},
		{
			testName: "real life example",
			input:    `{"TargetMetadata":{"SupportLobs":true,"LimitedSizeLobMode":true,"LobMaxSize":32},"FullLoadSettings":{"TargetTablePrepMode":"DROP_AND_CREATE","MaxFullLoadSubTasks":8,"TransactionConsistencyTimeout":600,"CommitRate":10000},"TTSettings":{"TTS3Settings":{},"TTRecordSettings":{}},"Logging":{},"ControlTablesSettings":{"HistoryTimeslotInMinutes":5},"StreamBufferSettings":{"StreamBufferCount":3,"StreamBufferSizeInMB":8},"ChangeProcessingTuning":{"BatchApplyPreserveTransaction":true,"BatchApplyTimeoutMin":1,"BatchApplyTimeoutMax":30,"BatchApplyMemoryLimit":500,"MinTransactionSize":1000,"CommitTimeout":1,"MemoryLimitTotal":1024,"MemoryKeepTime":60,"StatementCacheSize":50},"ChangeProcessingDdlHandlingPolicy":{"HandleSourceTableDropped":true,"HandleSourceTableTruncated":true,"HandleSourceTableAltered":true},"LoopbackPreventionSettings":{},"CharacterSetSettings":{"CharacterSetSupport":{}},"BeforeImageSettings":{},"ErrorBehavior":{"DataErrorPolicy":"LOG_ERROR","DataTruncationErrorPolicy":"LOG_ERROR","DataErrorEscalationPolicy":"SUSPEND_TABLE","TableErrorPolicy":"SUSPEND_TABLE","TableErrorEscalationPolicy":"STOP_TASK","RecoverableErrorCount":-1,"RecoverableErrorInterval":5,"RecoverableErrorThrottling":true,"RecoverableErrorThrottlingMax":1800,"ApplyErrorDeletePolicy":"IGNORE_RECORD","ApplyErrorInsertPolicy":"LOG_ERROR","ApplyErrorUpdatePolicy":"LOG_ERROR","ApplyErrorEscalationPolicy":"LOG_ERROR","FullLoadIgnoreConflicts":true},"ValidationSettings":{"ValidationMode":"ROW_LEVEL","ThreadCount":5,"PartitionSize":10000,"FailureMaxCount":10000,"TableFailureMaxCount":1000}}`,
			want:     `{"TargetMetadata":{"SupportLobs":true,"LimitedSizeLobMode":true,"LobMaxSize":32},"FullLoadSettings":{"TargetTablePrepMode":"DROP_AND_CREATE","MaxFullLoadSubTasks":8,"TransactionConsistencyTimeout":600,"CommitRate":10000},"ControlTablesSettings":{"HistoryTimeslotInMinutes":5},"StreamBufferSettings":{"StreamBufferCount":3,"StreamBufferSizeInMB":8},"ChangeProcessingTuning":{"BatchApplyPreserveTransaction":true,"BatchApplyTimeoutMin":1,"BatchApplyTimeoutMax":30,"BatchApplyMemoryLimit":500,"MinTransactionSize":1000,"CommitTimeout":1,"MemoryLimitTotal":1024,"MemoryKeepTime":60,"StatementCacheSize":50},"ChangeProcessingDdlHandlingPolicy":{"HandleSourceTableDropped":true,"HandleSourceTableTruncated":true,"HandleSourceTableAltered":true},"ErrorBehavior":{"DataErrorPolicy":"LOG_ERROR","DataTruncationErrorPolicy":"LOG_ERROR","DataErrorEscalationPolicy":"SUSPEND_TABLE","TableErrorPolicy":"SUSPEND_TABLE","TableErrorEscalationPolicy":"STOP_TASK","RecoverableErrorCount":-1,"RecoverableErrorInterval":5,"RecoverableErrorThrottling":true,"RecoverableErrorThrottlingMax":1800,"ApplyErrorDeletePolicy":"IGNORE_RECORD","ApplyErrorInsertPolicy":"LOG_ERROR","ApplyErrorUpdatePolicy":"LOG_ERROR","ApplyErrorEscalationPolicy":"LOG_ERROR","FullLoadIgnoreConflicts":true},"ValidationSettings":{"ValidationMode":"ROW_LEVEL","ThreadCount":5,"PartitionSize":10000,"FailureMaxCount":10000,"TableFailureMaxCount":1000}}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := tfjson.RemoveEmptyFields([]byte(testCase.input))
			if string(got) != testCase.want {
				t.Errorf("RemoveEmptyFields(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestRemoveEmptyStringFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "empty object",
			input:    `{}`,
			want:     `{}`,
		},
		{
			testName: "top-level empty string",
			input:    `""`,
			want:     `""`,
		},
		{
			testName: "single null field",
			input:    `{"key":null}`,
			want:     `{"key":null}`,
		},
		{
			testName: "single empty array field",
			input:    `{"key":[]}`,
			want:     `{"key":[]}`,
		},
		{
			testName: "single empty object field",
			input:    `{"key":{}}`,
			want:     `{"key":{}}`,
		},
		{
			testName: "single empty string field",
			input:    `{"key":""}`,
			want:     `{}`,
		},
		{
			testName: "single non-empty string field",
			input:    `{"key":"a"}`,
			want:     `{"key":"a"}`,
		},
		{
			testName: "empty strings in arrays are retained",
			input:    `{"key":["",{"empty":"","keep":"value"},""]}`,
			want:     `{"key":["",{"keep":"value"},""]}`,
		},
		{
			testName: "nested empty string fields",
			input:    `{"key":{"a":[1,"two",""],"b":"","c":{"d":"dee","e":""}}}`,
			want:     `{"key":{"a":[1,"two",""],"c":{"d":"dee"}}}`,
		},
		{
			testName: "invalid JSON",
			input:    `{"key":`,
			want:     "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := tfjson.RemoveEmptyStringFields([]byte(testCase.input))
			if string(got) != testCase.want {
				t.Errorf("RemoveEmptyStringFields(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func TestRemoveNullAndEmptyArrayValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "top-level empty array",
			input:    `[]`,
			want:     `[]`,
		},
		{
			testName: "empty objects retained",
			input:    `{"empty":{},"becomesEmpty":{"null":null}}`,
			want:     `{"empty":{},"becomesEmpty":{}}`,
		},
		{
			testName: "nested empty arrays and nulls",
			input:    `{"values":[[[]],[],[1,[]],null],"object":{"empty":[],"null":null}}`,
			want:     `{"values":[[1]],"object":{}}`,
		},
		{
			testName: "invalid JSON",
			input:    `{"key":`,
			want:     "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := tfjson.RemoveNullAndEmptyArrayValues([]byte(testCase.input))
			if string(got) != testCase.want {
				t.Errorf("RemoveNullAndEmptyArrayValues(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}

func FuzzRemoveEmptyFields(f *testing.F) {
	for _, input := range []string{
		`{}`,
		`[]`,
		`null`,
		`{"key":[null,{},[],{"nested":null},1]}`,
		`{"duplicate":null,"duplicate":{"value":1}}`,
	} {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if !json.Valid([]byte(input)) {
			return
		}

		got := tfjson.RemoveEmptyFields([]byte(input))
		assertValidAndIdempotent(t, got, tfjson.RemoveEmptyFields)
	})
}

func FuzzRemoveEmptyStringFields(f *testing.F) {
	for _, input := range []string{
		`{}`,
		`[]`,
		`""`,
		`{"key":["",{"nested":""}]}`,
		`{"duplicate":"","duplicate":"value"}`,
	} {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if !json.Valid([]byte(input)) {
			return
		}

		got := tfjson.RemoveEmptyStringFields([]byte(input))
		assertValidAndIdempotent(t, got, tfjson.RemoveEmptyStringFields)
	})
}

func FuzzRemoveNullAndEmptyArrayValues(f *testing.F) {
	for _, input := range []string{
		`{}`,
		`[]`,
		`null`,
		`{"key":[null,[],[[]],[1,[]]]}`,
		`{"duplicate":null,"duplicate":[1]}`,
	} {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if !json.Valid([]byte(input)) {
			return
		}

		got := tfjson.RemoveNullAndEmptyArrayValues([]byte(input))
		assertValidAndIdempotent(t, got, tfjson.RemoveNullAndEmptyArrayValues)
	})
}

func assertValidAndIdempotent(t *testing.T, got []byte, transform func([]byte) []byte) {
	t.Helper()

	if !json.Valid(got) {
		t.Fatalf("result is invalid JSON: %q", got)
	}
	if again := transform(got); !bytes.Equal(again, got) {
		t.Errorf("transform is not idempotent: first result %q, second result %q", got, again)
	}
}
