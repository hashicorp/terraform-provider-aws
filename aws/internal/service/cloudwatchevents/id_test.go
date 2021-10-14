package cloudwatchevents_test

import (
	"testing"

	tfevents "github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudwatchevents"
)

func TestPermissionParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputID       string
		ExpectedError bool
		ExpectedPart0 string
		ExpectedPart1 string
	}{
		{
			TestName:      "empty ID",
			InputID:       "",
			ExpectedError: true,
		},
		{
			TestName:      "single part",
			InputID:       "TestStatement",
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestStatement",
		},
		{
			TestName:      "two parts",
			InputID:       tfevents.PermissionCreateResourceID("TestEventBus", "TestStatement"),
			ExpectedPart0: "TestEventBus",
			ExpectedPart1: "TestStatement",
		},
		{
			TestName:      "two parts with default event bus",
			InputID:       tfevents.PermissionCreateResourceID(tfevents.DefaultEventBusName, "TestStatement"),
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestStatement",
		},
		{
			TestName:      "partner event bus",
			InputID:       "aws.partner/example.com/Test/TestStatement",
			ExpectedError: true,
		},
		{
			TestName:      "empty both parts",
			InputID:       "/",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part",
			InputID:       "/TestStatement",
			ExpectedError: true,
		},
		{
			TestName:      "empty second part",
			InputID:       "TestEventBus/",
			ExpectedError: true,
		},
		{
			TestName:      "three parts",
			InputID:       "TestEventBus/TestStatement/Suffix",
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotPart0, gotPart1, err := tfevents.PermissionParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotPart0 != testCase.ExpectedPart0 {
				t.Errorf("got part 0 %s, expected %s", gotPart0, testCase.ExpectedPart0)
			}

			if gotPart1 != testCase.ExpectedPart1 {
				t.Errorf("got part 1 %s, expected %s", gotPart1, testCase.ExpectedPart1)
			}
		})
	}
}

func TestRuleParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputID       string
		ExpectedError bool
		ExpectedPart0 string
		ExpectedPart1 string
	}{
		{
			TestName:      "empty ID",
			InputID:       "",
			ExpectedError: true,
		},
		{
			TestName:      "single part",
			InputID:       "TestRule",
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "two parts",
			InputID:       tfevents.RuleCreateResourceID("TestEventBus", "TestRule"),
			ExpectedPart0: "TestEventBus",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "two parts with default event bus",
			InputID:       tfevents.RuleCreateResourceID(tfevents.DefaultEventBusName, "TestRule"),
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "partner event bus 1",
			InputID:       "aws.partner/example.com/Test/TestRule",
			ExpectedPart0: "aws.partner/example.com/Test",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "partner event bus 2",
			InputID:       "aws.partner/foo.com/foo/18554d09-58ff-aa42-ba9c-c4c33899006f/test",
			ExpectedPart0: "aws.partner/foo.com/foo/18554d09-58ff-aa42-ba9c-c4c33899006f",
			ExpectedPart1: "test",
		},
		{
			TestName:      "ARN event bus",
			InputID:       tfevents.RuleCreateResourceID("arn:aws:events:us-east-2:123456789012:event-bus/default", "TestRule"),
			ExpectedPart0: "arn:aws:events:us-east-2:123456789012:event-bus/default",
			ExpectedPart1: "TestRule",
		},
		{
			TestName:      "empty both parts",
			InputID:       "/",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part",
			InputID:       "/TestRule",
			ExpectedError: true,
		},
		{
			TestName:      "empty second part",
			InputID:       "TestEventBus/",
			ExpectedError: true,
		},
		{
			TestName:      "empty partner event rule",
			InputID:       "aws.partner/example.com/Test/",
			ExpectedError: true,
		},
		{
			TestName:      "three parts",
			InputID:       "TestEventBus/TestRule/Suffix",
			ExpectedError: true,
		},
		{
			TestName:      "four parts",
			InputID:       "abc.partner/TestEventBus/TestRule/Suffix",
			ExpectedError: true,
		},
		{
			TestName:      "five parts",
			InputID:       "test/aws.partner/example.com/Test/TestRule",
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotPart0, gotPart1, err := tfevents.RuleParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotPart0 != testCase.ExpectedPart0 {
				t.Errorf("got part 0 %s, expected %s", gotPart0, testCase.ExpectedPart0)
			}

			if gotPart1 != testCase.ExpectedPart1 {
				t.Errorf("got part 1 %s, expected %s", gotPart1, testCase.ExpectedPart1)
			}
		})
	}
}

func TestTargetParseImportID(t *testing.T) {
	testCases := []struct {
		TestName      string
		InputID       string
		ExpectedError bool
		ExpectedPart0 string
		ExpectedPart1 string
		ExpectedPart2 string
	}{
		{
			TestName:      "empty ID",
			InputID:       "",
			ExpectedError: true,
		},
		{
			TestName:      "single part",
			InputID:       "TestRule",
			ExpectedError: true,
		},
		{
			TestName:      "two parts",
			InputID:       "TestTarget/TestRule",
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestTarget",
			ExpectedPart2: "TestRule",
		},
		{
			TestName:      "three parts",
			InputID:       "TestEventBus/TestRule/TestTarget",
			ExpectedPart0: "TestEventBus",
			ExpectedPart1: "TestRule",
			ExpectedPart2: "TestTarget",
		},
		{
			TestName:      "three parts with default event bus",
			InputID:       tfevents.DefaultEventBusName + "/TestRule/TestTarget",
			ExpectedPart0: tfevents.DefaultEventBusName,
			ExpectedPart1: "TestRule",
			ExpectedPart2: "TestTarget",
		},
		{
			TestName:      "empty two parts",
			InputID:       "/",
			ExpectedError: true,
		},
		{
			TestName:      "empty three parts",
			InputID:       "//",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part of two",
			InputID:       "/TestTarget",
			ExpectedError: true,
		},
		{
			TestName:      "empty second part of two",
			InputID:       "TestRule/",
			ExpectedError: true,
		},
		{
			TestName:      "empty first part of three",
			InputID:       "/TestRule/TestTarget",
			ExpectedError: true,
		},
		{
			TestName:      "empty second part of three",
			InputID:       "TestEventBus//TestTarget",
			ExpectedError: true,
		},
		{
			TestName:      "empty third part of three",
			InputID:       "TestEventBus/TestRule/",
			ExpectedError: true,
		},
		{
			TestName:      "empty first two of three parts",
			InputID:       "//TestTarget",
			ExpectedError: true,
		},
		{
			TestName:      "empty first and third of three parts",
			InputID:       "/TestRule/",
			ExpectedError: true,
		},
		{
			TestName:      "empty final two of three parts",
			InputID:       "TestEventBus//",
			ExpectedError: true,
		},
		{
			TestName:      "partner event bus",
			InputID:       "aws.partner/example.com/Test/TestRule/TestTarget",
			ExpectedPart0: "aws.partner/example.com/Test",
			ExpectedPart1: "TestRule",
			ExpectedPart2: "TestTarget",
		},
		{
			TestName:      "ARN event bus",
			InputID:       "arn:aws:events:us-east-2:123456789012:event-bus/default/TestRule/TestTarget",
			ExpectedPart0: "arn:aws:events:us-east-2:123456789012:event-bus/default",
			ExpectedPart1: "TestRule",
			ExpectedPart2: "TestTarget",
		},
		{
			TestName:      "empty partner event rule and target",
			InputID:       "aws.partner/example.com/Test//",
			ExpectedError: true,
		},
		{
			TestName:      "four parts",
			InputID:       "aws.partner/example.com/Test/TestRule",
			ExpectedError: true,
		},
		{
			TestName:      "five parts",
			InputID:       "abc.partner/example.com/Test/TestRule/TestTarget",
			ExpectedError: true,
		},
		{
			TestName:      "six parts",
			InputID:       "test/aws.partner/example.com/Test/TestRule/TestTarget",
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotPart0, gotPart1, gotPart2, err := tfevents.TargetParseImportID(testCase.InputID)

			if err == nil && testCase.ExpectedError {
				t.Fatalf("expected error, got no error")
			}

			if err != nil && !testCase.ExpectedError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if gotPart0 != testCase.ExpectedPart0 {
				t.Errorf("got part 0 %s, expected %s", gotPart0, testCase.ExpectedPart0)
			}

			if gotPart1 != testCase.ExpectedPart1 {
				t.Errorf("got part 1 %s, expected %s", gotPart1, testCase.ExpectedPart1)
			}

			if gotPart2 != testCase.ExpectedPart2 {
				t.Errorf("got part 2 %s, expected %s", gotPart2, testCase.ExpectedPart2)
			}
		})
	}
}
