// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
)

func TestValidEventSubscriptionName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := tfneptune.ValidEventSubscriptionName(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidEventSubscriptionNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "testing 123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 254, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := tfneptune.ValidEventSubscriptionNamePrefix(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name Prefix to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidParamGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfneptune.ValidParamGroupName(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidParamGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfneptune.ValidParamGroupNamePrefix(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidSubnetGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    "default",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 300, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfneptune.ValidSubnetGroupName(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfneptune.ValidSubnetGroupNamePrefix(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name prefix to trigger a validation error")
		}
	}
}
