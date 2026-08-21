// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"testing"
)

func TestRepositoryURIDualStack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard URI",
			input:    "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-repo", //lintignore:AWSAT003
			expected: "123456789012.dkr-ecr.us-west-2.on.aws/my-repo",        //lintignore:AWSAT003
		},
		{
			name:     "different region",
			input:    "123456789012.dkr.ecr.us-west-2.amazonaws.com/app/backend", //lintignore:AWSAT003
			expected: "123456789012.dkr-ecr.us-west-2.on.aws/app/backend",        //lintignore:AWSAT003
		},
		{
			name:     "nested repo path",
			input:    "123456789012.dkr.ecr.us-west-2.amazonaws.com/org/app/backend", //lintignore:AWSAT003
			expected: "123456789012.dkr-ecr.us-west-2.on.aws/org/app/backend",        //lintignore:AWSAT003
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := repositoryURIDualStack(tt.input)
			if result != tt.expected {
				t.Errorf("repositoryURIDualStack(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
