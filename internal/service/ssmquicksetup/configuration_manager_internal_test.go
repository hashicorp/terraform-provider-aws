// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup

import "testing"

func TestConfigurationManagerParametersEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		oldParameters map[string]string
		newParameters map[string]string
		want          bool
	}{
		"exact equality": {
			oldParameters: map[string]string{"OutputLogEnableS3": "false"},
			newParameters: map[string]string{"OutputLogEnableS3": "false"},
			want:          true,
		},
		"API default empty parameters added": {
			oldParameters: map[string]string{"OutputLogEnableS3": "false"},
			newParameters: map[string]string{
				"OutputLogEnableS3":  "false",
				"OutputBucketRegion": "",
				"OutputS3BucketName": "",
				"OutputS3KeyPrefix":  "",
			},
			want: true,
		},
		"API default empty parameters removed": {
			oldParameters: map[string]string{
				"OutputLogEnableS3":  "false",
				"OutputBucketRegion": "",
				"OutputS3BucketName": "",
				"OutputS3KeyPrefix":  "",
			},
			newParameters: map[string]string{"OutputLogEnableS3": "false"},
			want:          true,
		},
		"non-empty default parameter added": {
			oldParameters: map[string]string{"OutputLogEnableS3": "false"},
			newParameters: map[string]string{
				"OutputLogEnableS3":  "false",
				"OutputS3BucketName": "example-bucket",
			},
			want: false,
		},
		"non-default empty parameter added": {
			oldParameters: map[string]string{"OutputLogEnableS3": "false"},
			newParameters: map[string]string{
				"OutputLogEnableS3":  "false",
				"UnrelatedParameter": "",
			},
			want: false,
		},
		"configured parameter changed": {
			oldParameters: map[string]string{"RateControlConcurrency": "10%"},
			newParameters: map[string]string{"RateControlConcurrency": "20%"},
			want:          false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := configurationManagerParametersEqual(testCase.oldParameters, testCase.newParameters); got != testCase.want {
				t.Errorf("configurationManagerParametersEqual() = %t, want %t", got, testCase.want)
			}
		})
	}
}
