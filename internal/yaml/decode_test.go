// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package yaml_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/yaml"
)

func TestDecodeFromString(t *testing.T) {
	t.Parallel()

	type nested struct {
		A bool `yaml:"A"`
	}
	type to struct {
		A string   `yaml:"A"`
		B int      `yaml:"B"`
		C nested   `yaml:"C"`
		D []string `yaml:"D"`
	}
	var to0, to1, to2, to3 to
	to4 := to{
		A: "test1",
		B: 42,
		C: nested{A: true},
		D: []string{"test2", "test3"},
	}

	testCases := []struct {
		testName   string
		input      string
		output     any
		wantOutput any
		wantErr    bool
	}{
		{
			testName:   "empty YAML",
			input:      ``,
			output:     &to1,
			wantOutput: &to0,
		},
		{
			testName: "bad YAML",
			input:    `a`,
			output:   &to2,
			wantErr:  true,
		},
		{
			testName: "full YAML",
			input: `
---
A: test1
D:
  - test2
  - test3
C:
  A: true
B: 42
`,
			output:     &to3,
			wantOutput: &to4,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := yaml.DecodeFromString(testCase.input, testCase.output)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("DecodeFromString(%s) err %t, want %t", testCase.input, got, want)
			}
			if err == nil {
				if diff := cmp.Diff(testCase.output, testCase.wantOutput); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
