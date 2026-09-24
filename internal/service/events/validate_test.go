// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSuppressEquivalentBusNameOrARN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Old      string
		New      string
		Suppress bool
	}{
		{
			Name:     "same name",
			Old:      "my-bus",
			New:      "my-bus",
			Suppress: true,
		},
		{
			Name:     "same ARN",
			Old:      "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			New:      "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			Suppress: true,
		},
		{
			Name:     "name to equivalent ARN",
			Old:      "my-bus",
			New:      "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			Suppress: true,
		},
		{
			Name:     "ARN to equivalent name",
			Old:      "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			New:      "my-bus",
			Suppress: true,
		},
		{
			Name:     "different bus names",
			Old:      "my-bus",
			New:      "other-bus",
			Suppress: false,
		},
		{
			Name:     "name vs different ARN",
			Old:      "my-bus",
			New:      "arn:aws:events:us-east-1:123456789012:event-bus/other-bus", //lintignore:AWSAT003,AWSAT005
			Suppress: false,
		},
		{
			// In practice, the provider configures a single region so cross-region
			// comparison doesn't occur. We suppress based on bus name only.
			Name:     "ARNs in different regions same bus name",
			Old:      "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			New:      "arn:aws:events:eu-west-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			Suppress: true,
		},
		{
			Name:     "default bus name vs default ARN",
			Old:      "default",
			New:      "arn:aws:events:us-east-1:123456789012:event-bus/default", //lintignore:AWSAT003,AWSAT005
			Suppress: true,
		},
		{
			Name:     "govcloud ARN to name",
			Old:      "arn:aws-us-gov:events:us-gov-west-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			New:      "my-bus",
			Suppress: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			if got := tfevents.SuppressEquivalentBusNameOrARN("event_bus_name", tc.Old, tc.New, nil); got != tc.Suppress {
				t.Errorf("SuppressEquivalentBusNameOrARN(%q, %q) = %t, want %t", tc.Old, tc.New, got, tc.Suppress)
			}
		})
	}
}

func TestBusNameFromNameOrARN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Input    string
		Expected string
	}{
		{
			Input:    "my-bus",
			Expected: "my-bus",
		},
		{
			Input:    "arn:aws:events:us-east-1:123456789012:event-bus/my-bus", //lintignore:AWSAT003,AWSAT005
			Expected: "my-bus",
		},
		{
			Input:    "arn:aws-us-gov:events:us-gov-west-1:123456789012:event-bus/custom_bus-123", //lintignore:AWSAT003,AWSAT005
			Expected: "custom_bus-123",
		},
		{
			Input:    "default",
			Expected: "default",
		},
		{
			Input:    "arn:aws:events:us-east-1:123456789012:event-bus/default", //lintignore:AWSAT003,AWSAT005
			Expected: "default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Input, func(t *testing.T) {
			t.Parallel()
			if got := tfevents.BusNameFromNameOrARN(tc.Input); got != tc.Expected {
				t.Errorf("BusNameFromNameOrARN(%q) = %q, want %q", tc.Input, got, tc.Expected)
			}
		})
	}
}

func TestValidCustomEventBusSourceName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value   string
		IsValid bool
	}{
		{
			Value:   "",
			IsValid: false,
		},
		{
			Value:   "default",
			IsValid: false,
		},
		{
			Value:   "aws.partner/example.com/test/" + acctest.RandStringFromCharSet(t, 227, acctest.CharSetAlpha),
			IsValid: true,
		},
		{
			Value:   "aws.partner/example.com/test/" + acctest.RandStringFromCharSet(t, 228, acctest.CharSetAlpha),
			IsValid: false,
		},
		{
			Value:   "aws.partner/example.com/test/12345ab-cdef-1235",
			IsValid: true,
		},
		{
			Value:   "/test0._1-",
			IsValid: false,
		},
		{
			Value:   "test0._1-",
			IsValid: false,
		},
	}
	for _, tc := range cases {
		_, errors := tfevents.ValidSourceName(tc.Value, "aws_cloudwatch_event_bus_event_source_name")
		isValid := len(errors) == 0
		if tc.IsValid && !isValid {
			t.Errorf("expected %q to return valid, but did not", tc.Value)
		} else if !tc.IsValid && isValid {
			t.Errorf("expected %q to not return valid, but did", tc.Value)
		}
	}
}

func TestValidCustomEventBusName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value   string
		IsValid bool
	}{
		{
			Value:   "",
			IsValid: false,
		},
		{
			Value:   "default",
			IsValid: false,
		},
		{
			Value:   acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			IsValid: true,
		},
		{
			Value:   acctest.RandStringFromCharSet(t, 257, acctest.CharSetAlpha),
			IsValid: false,
		},
		{
			Value:   "aws.partner/example.com/test/12345ab-cdef-1235",
			IsValid: true,
		},
		{
			Value:   "/test0._1-",
			IsValid: true,
		},
		{
			Value:   "test0._1-",
			IsValid: true,
		},
	}
	for _, tc := range cases {
		_, errors := tfevents.ValidCustomEventBusName(tc.Value, "aws_cloudwatch_event_bus")
		isValid := len(errors) == 0
		if tc.IsValid && !isValid {
			t.Errorf("expected %q to return valid, but did not", tc.Value)
		} else if !tc.IsValid && isValid {
			t.Errorf("expected %q to not return valid, but did", tc.Value)
		}
	}
}

func TestValidBusNameOrARN(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello.World0125",
		"aws.partner/mongodb.com/stitch.trigger/something",                   // nosemgrep:ci.semgrep.domain-names.domain-names
		"arn:aws:events:us-east-1:123456789012:event-bus/default",            // lintignore:AWSAT003,AWSAT005
		"arn:aws-eusc:events:eusc-de-east-1:123456789012:event-bus/test-bus", // lintignore:AWSAT003,AWSAT005
	}
	for _, v := range validNames {
		_, errors := tfevents.ValidBusNameOrARN(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CW event bus name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"arn:aw:events:us-east-1:123456789012:event-bus/default", // lintignore:AWSAT003,AWSAT005
	}
	for _, v := range invalidNames {
		_, errors := tfevents.ValidBusNameOrARN(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CW event bus name", v)
		}
	}
}

func TestEventBusARNPattern(t *testing.T) {
	t.Parallel()

	validARNs := []string{
		"arn:aws:events:us-east-1:123456789012:event-bus/default",            //lintignore:AWSAT003,AWSAT005
		"arn:aws-eusc:events:eusc-de-east-1:123456789012:event-bus/test-bus", //lintignore:AWSAT003,AWSAT005
		"arn:aws-us-gov:events:us-gov-west-1:123456789012:event-bus/my-bus",  //lintignore:AWSAT003,AWSAT005
		"arn:aws:events:eu-west-1:123456789012:event-bus/custom_bus-123",     //lintignore:AWSAT003,AWSAT005
	}

	for _, arn := range validARNs {
		if !tfevents.EventBusARNPattern.MatchString(arn) {
			t.Errorf("Expected %q to match EventBusARNPattern", arn)
		}
	}

	invalidARNs := []string{
		"arn:aws:events:invalid-region:123456789012:event-bus/default", //lintignore:AWSAT003,AWSAT005
		"arn:aws:events:us-east-1:123456789012:event-bus/",             //lintignore:AWSAT003,AWSAT005
		"not-an-arn",
		"arn:aws:s3:::bucket", //lintignore:AWSAT005
	}

	for _, arn := range invalidARNs {
		if tfevents.EventBusARNPattern.MatchString(arn) {
			t.Errorf("Expected %q to NOT match EventBusARNPattern", arn)
		}
	}
}

func TestPartnerEventBusPattern(t *testing.T) {
	t.Parallel()

	validPatterns := []string{
		"aws.partner/mongodb.com/stitch.trigger/something",
		"arn:aws:events:us-east-1:123456789012:event-bus/aws.partner/genesys.com/cloud/test",                 //lintignore:AWSAT003,AWSAT005
		"arn:aws-eusc:events:eusc-de-east-1:123456789012:event-bus/aws.partner/example.com/service/instance", //lintignore:AWSAT003,AWSAT005
	}

	for _, pattern := range validPatterns {
		if !tfevents.PartnerEventBusPattern.MatchString(pattern) {
			t.Errorf("Expected %q to match PartnerEventBusPattern", pattern)
		}
	}

	invalidPatterns := []string{
		"aws.partner", // too short
		"not.partner/something/else",
		"arn:aws:events:invalid-region:123456789012:event-bus/aws.partner/test/service", //lintignore:AWSAT003,AWSAT005
	}

	for _, pattern := range invalidPatterns {
		if tfevents.PartnerEventBusPattern.MatchString(pattern) {
			t.Errorf("Expected %q to NOT match PartnerEventBusPattern", pattern)
		}
	}
}

func TestValidRuleName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello.World0125",
	}
	for _, v := range validNames {
		_, errors := tfevents.ValidateRuleName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CW event rule name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"slash/in-the-middle",
		// Length > 64
		"TooLooooooooooooooooooooooooooooooooooooooooooooooooooooooongName",
	}
	for _, v := range invalidNames {
		_, errors := tfevents.ValidateRuleName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CW event rule name", v)
		}
	}
}
