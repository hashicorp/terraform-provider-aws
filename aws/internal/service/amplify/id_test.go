package amplify_test

import (
	"testing"

	tfamplify "github.com/hashicorp/terraform-provider-aws/aws/internal/service/amplify"
)

func TestBranchParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName           string
		InputID            string
		ExpectError        bool
		ExpectedAppID      string
		ExpectedBranchName string
	}{
		{
			TestName:    "empty ID",
			InputID:     "",
			ExpectError: true,
		},
		{
			TestName:    "incorrect format",
			InputID:     "test",
			ExpectError: true,
		},
		{
			TestName:           "valid ID",
			InputID:            tfamplify.BranchCreateResourceID("appID", "branchName"),
			ExpectedAppID:      "appID",
			ExpectedBranchName: "branchName",
		},
		{
			TestName:           "valid ID one slash",
			InputID:            tfamplify.BranchCreateResourceID("appID", "part1/part_2"),
			ExpectedAppID:      "appID",
			ExpectedBranchName: "part1/part_2",
		},
		{
			TestName:           "valid ID three slashes",
			InputID:            tfamplify.BranchCreateResourceID("appID", "part1/part_2/part-3/part4"),
			ExpectedAppID:      "appID",
			ExpectedBranchName: "part1/part_2/part-3/part4",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotAppID, gotBranchName, err := tfamplify.BranchParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("unexpected error")
			}

			if gotAppID != testCase.ExpectedAppID {
				t.Errorf("got AppID %s, expected %s", gotAppID, testCase.ExpectedAppID)
			}

			if gotBranchName != testCase.ExpectedBranchName {
				t.Errorf("got BranchName %s, expected %s", gotBranchName, testCase.ExpectedBranchName)
			}
		})
	}
}
