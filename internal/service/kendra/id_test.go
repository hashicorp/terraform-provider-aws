package kendra_test

import (
	"testing"

	tfkendra "github.com/hashicorp/terraform-provider-aws/internal/service/kendra"
)

func TestExperienceParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName        string
		Input           string
		ExpectedId      string
		ExpectedIndexId string
		Error           bool
	}{
		{
			TestName:        "empty",
			Input:           "",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID",
			Input:           "abcdefg12345678/",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID separator",
			Input:           "abcdefg12345678:qwerty09876",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID with more than 1 separator",
			Input:           "abcdefg12345678/qwerty09876/zxcvbnm123456",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Valid ID",
			Input:           "abcdefg12345678/qwerty09876",
			ExpectedId:      "abcdefg12345678",
			ExpectedIndexId: "qwerty09876",
			Error:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotId, gotIndexId, err := tfkendra.ExperienceParseResourceID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (Id: %s, IndexId: %s) and no error, expected error", gotId, gotIndexId)
			}

			if gotId != testCase.ExpectedId {
				t.Errorf("got %s, expected %s", gotId, testCase.ExpectedIndexId)
			}

			if gotIndexId != testCase.ExpectedIndexId {
				t.Errorf("got %s, expected %s", gotIndexId, testCase.ExpectedIndexId)
			}
		})
	}
}

func TestQuerySuggestionsBlockListParseID(t *testing.T) {
	testCases := []struct {
		TestName        string
		Input           string
		ExpectedId      string
		ExpectedIndexId string
		Error           bool
	}{
		{
			TestName:        "empty",
			Input:           "",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID",
			Input:           "abcdefg12345678/",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID separator",
			Input:           "abcdefg12345678:qwerty09876",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID with more than 1 separator",
			Input:           "abcdefg12345678/qwerty09876/zxcvbnm123456",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Valid ID",
			Input:           "abcdefg12345678/qwerty09876",
			ExpectedId:      "abcdefg12345678",
			ExpectedIndexId: "qwerty09876",
			Error:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotId, gotIndexId, err := tfkendra.QuerySuggestionsBlockListParseResourceID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (Id: %s, IndexId: %s) and no error, expected error", gotId, gotIndexId)
			}

			if gotId != testCase.ExpectedId {
				t.Errorf("got %s, expected %s", gotId, testCase.ExpectedIndexId)
			}

			if gotIndexId != testCase.ExpectedIndexId {
				t.Errorf("got %s, expected %s", gotIndexId, testCase.ExpectedIndexId)
			}
		})
	}
}

func TestThesaurusParseResourceID(t *testing.T) {
	testCases := []struct {
		TestName        string
		Input           string
		ExpectedId      string
		ExpectedIndexId string
		Error           bool
	}{
		{
			TestName:        "empty",
			Input:           "",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID",
			Input:           "abcdefg12345678/",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID separator",
			Input:           "abcdefg12345678:qwerty09876",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Invalid ID with more than 1 separator",
			Input:           "abcdefg12345678/qwerty09876/zxcvbnm123456",
			ExpectedId:      "",
			ExpectedIndexId: "",
			Error:           true,
		},
		{
			TestName:        "Valid ID",
			Input:           "abcdefg12345678/qwerty09876",
			ExpectedId:      "abcdefg12345678",
			ExpectedIndexId: "qwerty09876",
			Error:           false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			gotId, gotIndexId, err := tfkendra.ThesaurusParseResourceID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (Id: %s, IndexId: %s) and no error, expected error", gotId, gotIndexId)
			}

			if gotId != testCase.ExpectedId {
				t.Errorf("got %s, expected %s", gotId, testCase.ExpectedIndexId)
			}

			if gotIndexId != testCase.ExpectedIndexId {
				t.Errorf("got %s, expected %s", gotIndexId, testCase.ExpectedIndexId)
			}
		})
	}
}
