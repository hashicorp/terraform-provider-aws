// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"testing"
)

func TestGetResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Description        string
		ResourceIdentifier string
		ExpectedIdentifier string
		ExpectedKey        string
		ExpectedError      func(err error) bool
	}{
		{
			Description:        "empty resource identifier",
			ResourceIdentifier: "",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (), expected ID,KEY"
			},
		},
		{
			Description:        "missing identifier",
			ResourceIdentifier: ",testkey",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (,testkey), expected ID,KEY"
			},
		},
		{
			Description:        "missing key",
			ResourceIdentifier: "testidentifier,",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (testidentifier,), expected ID,KEY"
			},
		},
		{
			Description:        "incorrect separator",
			ResourceIdentifier: "testidentifier;testkey",
			ExpectedError: func(err error) bool {
				return err.Error() == "invalid resource identifier (testidentifier;testkey), expected ID,KEY"
			},
		},
		{
			Description:        "correct",
			ResourceIdentifier: "testidentifier,testkey",
			ExpectedIdentifier: "testidentifier",
			ExpectedKey:        "testkey",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			t.Parallel()

			gotIdentifier, gotKey, err := GetResourceID(testCase.ResourceIdentifier)

			if err != nil && !testCase.ExpectedError(err) {
				t.Fatalf("unexpected error: %s", err)
			}

			if testCase.ExpectedError == nil && err != nil {
				t.Fatalf("expected no error, got error: %s", err)
			}

			if testCase.ExpectedError != nil && err == nil {
				t.Fatalf("expected error, got no error")
			}

			if gotIdentifier != testCase.ExpectedIdentifier {
				t.Errorf("got identifier %s, expected identifier %s", gotIdentifier, testCase.ExpectedIdentifier)
			}

			if gotKey != testCase.ExpectedKey {
				t.Errorf("got key %s, expected key %s", gotKey, testCase.ExpectedKey)
			}
		})
	}
}

func TestSetResourceId(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Description                string
		Identifier                 string
		Key                        string
		ExpectedResourceIdentifier string
	}{
		{
			Description:                "correct",
			Identifier:                 "testidentifier",
			Key:                        "testkey",
			ExpectedResourceIdentifier: "testidentifier,testkey",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			t.Parallel()

			got := SetResourceID(testCase.Identifier, testCase.Key)

			if got != testCase.ExpectedResourceIdentifier {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedResourceIdentifier)
			}
		})
	}
}
