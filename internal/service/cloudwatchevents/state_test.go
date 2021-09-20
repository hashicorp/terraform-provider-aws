package cloudwatchevents_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfcloudwatchevents "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchevents"
)

func TestRuleEnabledFromState(t *testing.T) {
	testCases := []struct {
		TestName        string
		State           string
		ExpectedError   bool
		ExpectedEnabled bool
	}{
		{
			TestName:      "empty state",
			ExpectedError: true,
		},
		{
			TestName:      "invalid state",
			State:         "UNKNOWN",
			ExpectedError: true,
		},
		{
			TestName:        "enabled",
			State:           "ENABLED",
			ExpectedEnabled: true,
		},
		{
			TestName:        "disabled",
			State:           "DISABLED",
			ExpectedEnabled: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotEnabled, err := tfcloudwatchevents.RuleEnabledFromState(testCase.State)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotEnabled != testCase.ExpectedEnabled {
				t.Errorf("got enabled %t, expected %t", gotEnabled, testCase.ExpectedEnabled)
			}
		})
	}
}

func RuleStateFromEnabled(t *testing.T) {
	testCases := []struct {
		TestName      string
		Enabled       bool
		ExpectedState string
	}{
		{
			TestName:      "enabled",
			Enabled:       true,
			ExpectedState: "ENABLED",
		},
		{
			TestName:      "disabled",
			Enabled:       false,
			ExpectedState: "DISABLED",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotState := tfcloudwatchevents.RuleStateFromEnabled(testCase.Enabled)

			if gotState != testCase.ExpectedState {
				t.Errorf("got enabled %s, expected %s", gotState, testCase.ExpectedState)
			}
		})
	}
}
