// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"fmt"
	"math/rand"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"tf-test-elb",
	}

	for _, s := range validNames {
		_, errors := validName(s, names.AttrName)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid ELB name: %v", s, errors)
		}
	}

	invalidNames := []string{
		"tf.test.elb.1",
		"tf-test-elb-tf-test-elb-tf-test-elb",
		"-tf-test-elb",
		"tf-test-elb-",
	}

	for _, s := range invalidNames {
		_, errors := validName(s, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name: %v", s, errors)
		}
	}
}

func TestValidLoadBalancerNameCannotBeginWithHyphen(t *testing.T) {
	t.Parallel()

	var n = "-Testing123"
	_, errors := validName(n, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestValidLoadBalancerNameCanBeAnEmptyString(t *testing.T) {
	t.Parallel()

	var n = ""
	_, errors := validName(n, "SampleKey")

	if len(errors) != 0 {
		t.Fatalf("Expected the ELB Name to pass validation")
	}
}

func TestValidLoadBalancerNameCannotBeLongerThan32Characters(t *testing.T) {
	t.Parallel()

	var n = "Testing123dddddddddddddddddddvvvv"
	_, errors := validName(n, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestValidLoadBalancerNameCannotHaveSpecialCharacters(t *testing.T) {
	t.Parallel()

	var n = "Testing123%%"
	_, errors := validName(n, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestValidLoadBalancerNameCannotEndWithHyphen(t *testing.T) {
	t.Parallel()

	var n = "Testing123-"
	_, errors := validName(n, "SampleKey")

	if len(errors) != 1 {
		t.Fatalf("Expected the ELB Name to trigger a validation error")
	}
}

func TestValidNamePrefix(t *testing.T) {
	t.Parallel()

	validNamePrefixes := []string{
		"test-",
	}

	for _, s := range validNamePrefixes {
		_, errors := validNamePrefix(s, names.AttrNamePrefix)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid ELB name prefix: %v", s, errors)
		}
	}

	invalidNamePrefixes := []string{
		"tf.test.elb.",
		"tf-test",
		"-test",
	}

	for _, s := range invalidNamePrefixes {
		_, errors := validNamePrefix(s, names.AttrNamePrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name prefix: %v", s, errors)
		}
	}
}

func TestValidLoadBalancerAccessLogsInterval(t *testing.T) {
	t.Parallel()

	type testCases struct {
		Value    int
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    0,
			ErrCount: 1,
		},
		{
			Value:    10,
			ErrCount: 1,
		},
		{
			Value:    -1,
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validAccessLogsInterval(tc.Value, names.AttrInterval)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}
}

func TestValidLoadBalancerHealthCheckTarget(t *testing.T) {
	t.Parallel()

	type testCase struct {
		Value    string
		ErrCount int
	}

	randomRunes := func(n int) string {
		// A complete set of modern Katakana characters.
		runes := []rune("アイウエオ" +
			"カキクケコガギグゲゴサシスセソザジズゼゾ" +
			"タチツテトダヂヅデドナニヌネノハヒフヘホ" +
			"バビブベボパピプペポマミムメモヤユヨラリ" +
			"ルレロワヰヱヲン")

		s := make([]rune, n)
		for i := range s {
			s[i] = runes[rand.Intn(len(runes))]
		}
		return string(s)
	}

	validCases := []testCase{
		{
			Value:    "TCP:1234",
			ErrCount: 0,
		},
		{
			Value:    "http:80/test",
			ErrCount: 0,
		},
		{
			Value:    fmt.Sprintf("HTTP:8080/%s", randomRunes(5)),
			ErrCount: 0,
		},
		{
			Value:    "SSL:8080",
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := validHeathCheckTarget(tc.Value, names.AttrTarget)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}

	invalidCases := []testCase{
		{
			Value:    "",
			ErrCount: 1,
		},
		{
			Value:    "TCP:",
			ErrCount: 1,
		},
		{
			Value:    "TCP:1234/",
			ErrCount: 1,
		},
		{
			Value:    "SSL:8080/",
			ErrCount: 1,
		},
		{
			Value:    "HTTP:8080",
			ErrCount: 1,
		},
		{
			Value:    "incorrect-value",
			ErrCount: 1,
		},
		{
			Value:    "TCP:123456",
			ErrCount: 1,
		},
		{
			Value:    "incorrect:80/",
			ErrCount: 1,
		},
		{
			Value: fmt.Sprintf("HTTP:8080/%s%s",
				sdkacctest.RandStringFromCharSet(512, sdkacctest.CharSetAlpha), randomRunes(512)),
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := validHeathCheckTarget(tc.Value, names.AttrTarget)
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}
}
