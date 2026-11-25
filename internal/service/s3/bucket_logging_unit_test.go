// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestResourceBucketLoggingFlatten_targetObjectKeyFormat exercises how the
// TargetObjectKeyFormat returned by GetBucketLogging is written to state.
// Amazon S3 only returns it when it was explicitly set, but S3-compatible
// backends (e.g. Ceph RADOS Gateway, OVHcloud Object Storage) always return
// it, with its default SimplePrefix. That default must not reach state when
// target_object_key_format is not configured, otherwise every plan wants to
// remove it.
func TestResourceBucketLoggingFlatten_targetObjectKeyFormat(t *testing.T) {
	t.Parallel()

	simplePrefix := &awstypes.TargetObjectKeyFormat{
		SimplePrefix: &awstypes.SimplePrefix{},
	}
	partitionedPrefix := &awstypes.TargetObjectKeyFormat{
		PartitionedPrefix: &awstypes.PartitionedPrefix{
			PartitionDateSource: awstypes.PartitionDateSourceEventTime,
		},
	}

	testCases := []struct {
		name       string
		configured []any
		observed   *awstypes.TargetObjectKeyFormat
		want       int
	}{
		{
			name:     "not returned",
			observed: nil,
			want:     0,
		},
		{
			name:     "default SimplePrefix returned, not configured",
			observed: simplePrefix,
			want:     0,
		},
		{
			name:       "SimplePrefix returned, configured",
			configured: []any{map[string]any{"simple_prefix": []any{map[string]any{}}}},
			observed:   simplePrefix,
			want:       1,
		},
		{
			name:     "PartitionedPrefix returned, not configured",
			observed: partitionedPrefix,
			want:     1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			raw := map[string]any{
				names.AttrBucket: "example",
				"target_bucket":  "example-logs",
				"target_prefix":  "",
			}
			if testCase.configured != nil {
				raw["target_object_key_format"] = testCase.configured
			}
			d := schema.TestResourceDataRaw(t, tfs3.ResourceBucketLogging().SchemaMap(), raw)

			loggingEnabled := &awstypes.LoggingEnabled{
				TargetBucket:          aws.String("example-logs"),
				TargetPrefix:          aws.String(""),
				TargetObjectKeyFormat: testCase.observed,
			}
			if err := tfs3.ResourceBucketLoggingFlatten(loggingEnabled, d); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if got := len(d.Get("target_object_key_format").([]any)); got != testCase.want {
				t.Errorf("got %d target_object_key_format block(s), want %d", got, testCase.want)
			}
		})
	}
}
