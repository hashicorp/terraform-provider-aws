package events

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidCustomEventBusSourceName(t *testing.T) {
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
			Value:   "aws.partner/example.com/test/" + sdkacctest.RandStringFromCharSet(227, sdkacctest.CharSetAlpha),
			IsValid: true,
		},
		{
			Value:   "aws.partner/example.com/test/" + sdkacctest.RandStringFromCharSet(228, sdkacctest.CharSetAlpha),
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
		_, errors := validSourceName(tc.Value, "aws_cloudwatch_event_bus_event_source_name")
		isValid := len(errors) == 0
		if tc.IsValid && !isValid {
			t.Errorf("expected %q to return valid, but did not", tc.Value)
		} else if !tc.IsValid && isValid {
			t.Errorf("expected %q to not return valid, but did", tc.Value)
		}
	}
}

func TestValidCustomEventBusName(t *testing.T) {
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
			Value:   sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			IsValid: true,
		},
		{
			Value:   sdkacctest.RandStringFromCharSet(257, sdkacctest.CharSetAlpha),
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
		_, errors := validCustomEventBusName(tc.Value, "aws_cloudwatch_event_bus")
		isValid := len(errors) == 0
		if tc.IsValid && !isValid {
			t.Errorf("expected %q to return valid, but did not", tc.Value)
		} else if !tc.IsValid && isValid {
			t.Errorf("expected %q to not return valid, but did", tc.Value)
		}
	}
}

func TestValidBusNameOrARN(t *testing.T) {
	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello.World0125",
		"aws.partner/mongodb.com/stitch.trigger/something",        // nosemgrep: domain-names
		"arn:aws:events:us-east-1:123456789012:event-bus/default", // lintignore:AWSAT003,AWSAT005
	}
	for _, v := range validNames {
		_, errors := validBusNameOrARN(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CW event rule name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"special@character",
		"arn:aw:events:us-east-1:123456789012:event-bus/default", // lintignore:AWSAT003,AWSAT005
	}
	for _, v := range invalidNames {
		_, errors := validBusNameOrARN(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CW event rule name", v)
		}
	}
}

func TestValidRuleName(t *testing.T) {
	validNames := []string{
		"HelloWorl_d",
		"hello-world",
		"hello.World0125",
	}
	for _, v := range validNames {
		_, errors := validateRuleName(v, "name")
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
		_, errors := validateRuleName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CW event rule name", v)
		}
	}
}
