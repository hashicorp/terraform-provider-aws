// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBatchListTags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tagsByIdentifier map[string][]awstypes.Tag
	}{
		"empty": {
			tagsByIdentifier: map[string][]awstypes.Tag{},
		},
		"single": {
			tagsByIdentifier: map[string][]awstypes.Tag{
				"id1": {
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Owner"),
						Value: aws.String("TeamA"),
					},
				},
			},
		},
		"multiple": {
			tagsByIdentifier: map[string][]awstypes.Tag{
				"id1": {
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Owner"),
						Value: aws.String("TeamA"),
					},
				},
				"id2": {
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("Owner"),
						Value: aws.String("TeamA"),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conn := &mockTagDescriber{
				tagsByIdentifier: testCase.tagsByIdentifier,
			}

			identifiers := slices.Collect(maps.Keys(testCase.tagsByIdentifier))

			got, err := batchListTags(t.Context(), conn, identifiers)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if diff := cmp.Diff(testCase.tagsByIdentifier, got, cmpopts.IgnoreUnexported(awstypes.Tag{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

type mockTagDescriber struct {
	tagsByIdentifier map[string][]awstypes.Tag
}

func (m *mockTagDescriber) DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error) {
	if len(params.ResourceArns) == 0 {
		return nil, &smithy.OperationError{
			ServiceID:     "Elastic Load Balancing v2",
			OperationName: "DescribeTags",
			Err: &smithy.GenericAPIError{
				Code:    "ValidationError",
				Message: "An ARN must be specified",
			},
		}
	}

	output := &elasticloadbalancingv2.DescribeTagsOutput{
		TagDescriptions: []awstypes.TagDescription{},
	}

	for _, identifier := range params.ResourceArns {
		if tags, ok := m.tagsByIdentifier[identifier]; ok {
			output.TagDescriptions = append(output.TagDescriptions, awstypes.TagDescription{
				ResourceArn: aws.String(identifier),
				Tags:        tags,
			})
		}
	}

	return output, nil
}
