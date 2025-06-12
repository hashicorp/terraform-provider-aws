// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"
	"time"
)

func TestSuppressEquivalentCloudWatchLogsLogGroupARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		old  string
		new  string
		want bool
	}{
		{
			old:  "arn:aws:s3:::tf-acc-test-3740243764086645346", //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:s3:::tf-acc-test-3740243764086645346", //lintignore:AWSAT003,AWSAT005
			want: true,
		},
		{
			old:  "arn:aws:s3:::tf-acc-test-3740243764086645346",                                                    //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			want: false,
		},
		{
			old:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			want: true,
		},
		{
			old:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346",   //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			want: true,
		},
		{
			old:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645347:*", //lintignore:AWSAT003,AWSAT005
			want: false,
		},
		{
			old:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645346:*", //lintignore:AWSAT003,AWSAT005
			new:  "arn:aws:logs:us-west-2:123456789012:log-group:/aws/vpclattice/tf-acc-test-3740243764086645347",   //lintignore:AWSAT003,AWSAT005
			want: false,
		},
	}
	for _, testCase := range testCases {
		if got, want := SuppressEquivalentCloudWatchLogsLogGroupARN("test_property", testCase.old, testCase.new, nil), testCase.want; got != want {
			t.Errorf("SuppressEquivalentCloudWatchLogsLogGroupARN(%q, %q) = %v, want %v", testCase.old, testCase.new, got, want)
		}
	}
}

func TestSuppressEquivalentRoundedTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		old        string
		new        string
		layout     string
		d          time.Duration
		equivalent bool
	}{
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:13.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:01:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: true,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Minute,
			equivalent: false,
		},
		{
			old:        "2024-04-19T23:00:00.000Z",
			new:        "2024-04-19T23:00:45.000Z",
			layout:     time.RFC3339,
			d:          time.Hour,
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		value := SuppressEquivalentRoundedTime(tc.layout, tc.d)("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}
