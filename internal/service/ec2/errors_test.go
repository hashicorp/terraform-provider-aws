// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestUnsuccessfulItemError(t *testing.T) {
	t.Parallel()

	unsuccessfulItemError := &awstypes.UnsuccessfulItemError{
		Code:    aws.String("test code"),
		Message: aws.String("test message"),
	}

	err := tfec2.UnsuccessfulItemError(unsuccessfulItemError)

	if !tfawserr.ErrCodeEquals(err, "test code") {
		t.Errorf("tfawserr.ErrCodeEquals failed: %s", err)
	}

	if !tfawserr.ErrMessageContains(err, "test code", "est mess") {
		t.Errorf("tfawserr.ErrMessageContains failed: %s", err)
	}
}

func TestUnsuccessfulItemsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Items    []awstypes.UnsuccessfulItem
		Expected bool
	}{
		{
			Name: "no items",
		},
		{
			Name: "one item no error",
			Items: []awstypes.UnsuccessfulItem{
				{
					ResourceId: aws.String("test resource"),
				},
			},
		},
		{
			Name: "one item",
			Items: []awstypes.UnsuccessfulItem{
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource"),
				},
			},
			Expected: true,
		},
		{
			Name: "two items, first no error",
			Items: []awstypes.UnsuccessfulItem{
				{
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
			Expected: true,
		},
		{
			Name: "two items, first not as expected",
			Items: []awstypes.UnsuccessfulItem{
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("not what is required"),
						Message: aws.String("not what is wanted"),
					},
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
		},
		{
			Name: "two items, first as expected",
			Items: []awstypes.UnsuccessfulItem{
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("test code"),
						Message: aws.String("test message"),
					},
					ResourceId: aws.String("test resource 1"),
				},
				{
					Error: &awstypes.UnsuccessfulItemError{
						Code:    aws.String("not what is required"),
						Message: aws.String("not what is wanted"),
					},
					ResourceId: aws.String("test resource 2"),
				},
			},
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			err := tfec2.UnsuccessfulItemsError(testCase.Items)

			got := tfawserr.ErrCodeEquals(err, "test code")

			if got != testCase.Expected {
				t.Errorf("ErrCodeEquals got %t, expected %t", got, testCase.Expected)
			}

			got = tfawserr.ErrMessageContains(err, "test code", "est mess")

			if got != testCase.Expected {
				t.Errorf("ErrMessageContains got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
