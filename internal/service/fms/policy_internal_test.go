// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestExpandPolicyResourceTags(t *testing.T) {
	t.Parallel()

	securityServicePolicyData := []any{map[string]any{
		"type":                 "WAFV2",
		"managed_service_data": `{"type":"WAFV2"}`,
	}}

	testCases := map[string]struct {
		raw                 map[string]any
		wantTags            []awstypes.ResourceTag
		wantLogicalOperator awstypes.ResourceTagLogicalOperator
	}{
		"no tags": {
			raw: map[string]any{
				"resource_tag_logical_operator": "OR",
			},
		},
		"resource_tags": {
			raw: map[string]any{
				"resource_tags": map[string]any{
					"key1": "value1",
				},
				"resource_tag_logical_operator": "OR",
			},
			wantTags: []awstypes.ResourceTag{
				{Key: aws.String("key1"), Value: aws.String("value1")},
			},
			wantLogicalOperator: awstypes.ResourceTagLogicalOperatorOr,
		},
		"resource_tag repeated key": {
			raw: map[string]any{
				"resource_tag": []any{
					map[string]any{"key": "key1", "value": "value1"},
					map[string]any{"key": "key1", "value": "value2"},
					map[string]any{"key": "key2"},
				},
				"resource_tag_logical_operator": "OR",
			},
			wantTags: []awstypes.ResourceTag{
				{Key: aws.String("key1"), Value: aws.String("value1")},
				{Key: aws.String("key1"), Value: aws.String("value2")},
				{Key: aws.String("key2"), Value: aws.String("")},
			},
			wantLogicalOperator: awstypes.ResourceTagLogicalOperatorOr,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			raw := map[string]any{
				"exclude_resource_tags":        true,
				"name":                         "test",
				"security_service_policy_data": securityServicePolicyData,
			}
			for k, v := range testCase.raw {
				raw[k] = v
			}
			d := schema.TestResourceDataRaw(t, resourcePolicy().SchemaMap(), raw)

			got := expandPolicy(d)

			sortTags := cmpopts.SortSlices(func(a, b awstypes.ResourceTag) bool {
				return aws.ToString(a.Key)+"="+aws.ToString(a.Value) < aws.ToString(b.Key)+"="+aws.ToString(b.Value)
			})
			if diff := cmp.Diff(testCase.wantTags, got.ResourceTags, sortTags, cmpopts.IgnoreUnexported(awstypes.ResourceTag{}), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unexpected ResourceTags diff (-want +got):\n%s", diff)
			}
			if got, want := got.ResourceTagLogicalOperator, testCase.wantLogicalOperator; got != want {
				t.Errorf("ResourceTagLogicalOperator = %q, want %q", got, want)
			}
		})
	}
}

func TestHasDuplicateResourceTagKeys(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input []awstypes.ResourceTag
		want  bool
	}{
		"nil": {},
		"unique keys": {
			input: []awstypes.ResourceTag{
				{Key: aws.String("key1"), Value: aws.String("value1")},
				{Key: aws.String("key2"), Value: aws.String("value1")},
			},
		},
		"repeated key": {
			input: []awstypes.ResourceTag{
				{Key: aws.String("key1"), Value: aws.String("value1")},
				{Key: aws.String("key1"), Value: aws.String("value2")},
			},
			want: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := hasDuplicateResourceTagKeys(testCase.input); got != testCase.want {
				t.Errorf("hasDuplicateResourceTagKeys() = %t, want %t", got, testCase.want)
			}
		})
	}
}
