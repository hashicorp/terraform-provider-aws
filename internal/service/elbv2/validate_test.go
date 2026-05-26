// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
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
		"internal-tf-test-elb",
	}

	for _, s := range invalidNames {
		_, errors := validName(s, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name: %v", s, errors)
		}
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
		"internal-",
	}

	for _, s := range invalidNamePrefixes {
		_, errors := validNamePrefix(s, names.AttrNamePrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid ELB name prefix: %v", s, errors)
		}
	}
}

func TestValidTargetGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tf.test.elb.target.1",
			ErrCount: 1,
		},
		{
			Value:    "-tf-test-target",
			ErrCount: 1,
		},
		{
			Value:    "tf-test-target-",
			ErrCount: 1,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(33, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validTargetGroupName(tc.Value, "aws_lb_target_group")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS LB Target Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidTargetGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tf.lb",
			ErrCount: 1,
		},
		{
			Value:    "-tf-lb",
			ErrCount: 1,
		},
		{
			Value:    sdkacctest.RandStringFromCharSet(32, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}
	for _, tc := range cases {
		_, errors := validTargetGroupNamePrefix(tc.Value, "aws_lb_target_group")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the AWS LB Target Group Name to trigger a validation error for %q", tc.Value)
		}
	}
}

func TestValidateTargetGroupHealthCheckIntervalTimeout(t *testing.T) {
	t.Parallel()

	healthCheckPath := cty.GetAttrPath(names.AttrHealthCheck).IndexInt(0)

	testCases := []struct {
		name     string
		interval int
		timeout  int
		wantErr  string
	}{
		{
			name:     "interval_greater_than_timeout",
			interval: 6,
			timeout:  5,
		},
		{
			name:     "interval_equals_timeout",
			interval: 5,
			timeout:  5,
			wantErr:  `Attribute "health_check[0].interval" must be greater than "health_check[0].timeout".`,
		},
		{
			name:     "interval_less_than_timeout",
			interval: 4,
			timeout:  5,
			wantErr:  `Attribute "health_check[0].interval" must be greater than "health_check[0].timeout".`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTargetGroupHealthCheckIntervalTimeout(healthCheckPath, tc.interval, tc.timeout)

			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if got := err.Error(); !strings.Contains(got, tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, got)
			}
		})
	}
}
