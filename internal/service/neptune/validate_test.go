package neptune

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidEventSubscriptionName(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validEventSubscriptionName(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidEventSubscriptionNamePrefix(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(254, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validEventSubscriptionNamePrefix(tc.Value, "aws_neptune_event_subscription")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Event Subscription Name Prefix to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidParamGroupName(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validParamGroupName(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidParamGroupNamePrefix(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validParamGroupNamePrefix(tc.Value, "aws_neptune_cluster_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Parameter Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidSubnetGroupName(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(300, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validSubnetGroupName(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupNamePrefix(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(230, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validSubnetGroupNamePrefix(tc.Value, "aws_neptune_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Neptune Subnet Group name prefix to trigger a validation error")
		}
	}
}
