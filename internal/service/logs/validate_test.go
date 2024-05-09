// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidLogGroupName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ValidLogGroupName",
		"ValidLogGroup.Name",
		"valid/Log-group",
		"1234",
		"YadaValid#0123",
		"Also_valid-name",
		strings.Repeat("W", 512),
	}
	for _, v := range validNames {
		_, errors := validLogGroupName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Group name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		// length > 512
		strings.Repeat("W", 513),
	}
	for _, v := range invalidNames {
		_, errors := validLogGroupName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Group name", v)
		}
	}
}

func TestValidLogGroupNamePrefix(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"ValidLogGroupName",
		"ValidLogGroup.Name",
		"valid/Log-group",
		"1234",
		"YadaValid#0123",
		"Also_valid-name",
		strings.Repeat("W", 483),
	}
	for _, v := range validNames {
		_, errors := validLogGroupNamePrefix(v, names.AttrNamePrefix)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Group name prefix: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		// length > 483
		strings.Repeat("W", 484),
	}
	for _, v := range invalidNames {
		_, errors := validLogGroupNamePrefix(v, names.AttrNamePrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Group name prefix", v)
		}
	}
}

func TestValidLogMetricFilterName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"YadaHereAndThere",
		"Valid-5Metric_Name",
		"This . is also %% valid@!)+(",
		"1234",
		strings.Repeat("W", 512),
	}
	for _, v := range validNames {
		_, errors := validLogMetricFilterName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Metric Filter Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"*",
		// length > 512
		strings.Repeat("W", 513),
	}
	for _, v := range invalidNames {
		_, errors := validLogMetricFilterName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Metric Filter Name", v)
		}
	}
}

func TestValidLogMetricTransformationName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"YadaHereAndThere",
		"Valid-5Metric_Name",
		"This . is also %% valid@!)+(",
		"1234",
		"",
		strings.Repeat("W", 255),
	}
	for _, v := range validNames {
		_, errors := validLogMetricFilterTransformationName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid Log Metric Filter Transformation Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"*",
		// length > 255
		strings.Repeat("W", 256),
	}
	for _, v := range invalidNames {
		_, errors := validLogMetricFilterTransformationName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid Log Metric Filter Transformation Name", v)
		}
	}
}

func TestValidStreamName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"test-log-stream",
		"my_sample_log_stream",
		"012345678",
		"logstream/1234",
	}
	for _, v := range validNames {
		_, errors := validStreamName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CloudWatch LogStream name: %q", v, errors)
		}
	}

	invalidNames := []string{
		sdkacctest.RandString(513),
		"",
		"stringwith:colon",
	}
	for _, v := range invalidNames {
		_, errors := validStreamName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CloudWatch LogStream name", v)
		}
	}
}
