package s3_test

import (
	"testing"

	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestParseResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName            string
		InputID             string
		ExpectError         bool
		ExpectedBucket      string
		ExpectedBucketOwner string
	}{
		{
			TestName:    "empty ID",
			InputID:     "",
			ExpectError: true,
		},
		{
			TestName:    "incorrect format",
			InputID:     "test,example,123456789012",
			ExpectError: true,
		},
		{
			TestName:            "valid ID with bucket",
			InputID:             tfs3.CreateResourceID("example", ""),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "",
		},
		{
			TestName:            "valid ID with bucket and bucket owner",
			InputID:             tfs3.CreateResourceID("example", "123456789012"),
			ExpectedBucket:      "example",
			ExpectedBucketOwner: "123456789012",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotBucket, gotExpectedBucketOwner, err := tfs3.ParseResourceID(testCase.InputID)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("unexpected error")
			}

			if gotBucket != testCase.ExpectedBucket {
				t.Errorf("got bucket %s, expected %s", gotBucket, testCase.ExpectedBucket)
			}

			if gotExpectedBucketOwner != testCase.ExpectedBucketOwner {
				t.Errorf("got ExpectedBucketOwner %s, expected %s", gotExpectedBucketOwner, testCase.ExpectedBucketOwner)
			}
		})
	}
}
