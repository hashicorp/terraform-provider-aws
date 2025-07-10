// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
)

func TestEndpointGroupARNToListenerARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789", //lintignore:AWSAT005
			ExpectedError: regexache.MustCompile(`expected service globalaccelerator`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456", //lintignore:AWSAT005
			ExpectedError: regexache.MustCompile(`expected at least 6 resource parts`),
		},
		{
			TestName:    "valid ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789", //lintignore:AWSAT005
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456",                       //lintignore:AWSAT005
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfglobalaccelerator.EndpointGroupARNToListenerARN(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}

func TestListenerOrEndpointGroupARNToAcceleratorARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName      string
		InputARN      string
		ExpectedError *regexp.Regexp
		ExpectedARN   string
	}{
		{
			TestName:      "empty ARN",
			InputARN:      "",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "unparsable ARN",
			InputARN:      "test",
			ExpectedError: regexache.MustCompile(`parsing ARN`),
		},
		{
			TestName:      "invalid ARN service",
			InputARN:      "arn:aws:ec2::123456789012:accelerator/a-123/listener/l-456", //lintignore:AWSAT005
			ExpectedError: regexache.MustCompile(`expected service globalaccelerator`),
		},
		{
			TestName:      "invalid ARN resource parts",
			InputARN:      "arn:aws:globalaccelerator::123456789012:accelerator/a-123", //lintignore:AWSAT005
			ExpectedError: regexache.MustCompile(`expected at least 4 resource parts`),
		},
		{
			TestName:    "valid listener ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456", //lintignore:AWSAT005
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123",                //lintignore:AWSAT005
		},
		{
			TestName:    "valid endpoint group ARN",
			InputARN:    "arn:aws:globalaccelerator::123456789012:accelerator/a-123/listener/l-456/endpoint-group/eg-789", //lintignore:AWSAT005
			ExpectedARN: "arn:aws:globalaccelerator::123456789012:accelerator/a-123",                                      //lintignore:AWSAT005
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			got, err := tfglobalaccelerator.ListenerOrEndpointGroupARNToAcceleratorARN(testCase.InputARN)

			if err == nil && testCase.ExpectedError != nil {
				t.Fatalf("expected error %s, got no error", testCase.ExpectedError.String())
			}

			if err != nil && testCase.ExpectedError == nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if err != nil && !testCase.ExpectedError.MatchString(err.Error()) {
				t.Fatalf("expected error %s, got: %s", testCase.ExpectedError.String(), err)
			}

			if got != testCase.ExpectedARN {
				t.Errorf("got %s, expected %s", got, testCase.ExpectedARN)
			}
		})
	}
}
