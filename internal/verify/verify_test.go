// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"testing"
)

func TestCheckYAMLString(t *testing.T) {
	t.Parallel()

	var err error
	var actual string

	validYaml := `---
abc:
  def: 123
  xyz:
    -
      a: "ホリネズミ"
      b: "1"
`

	actual, err = checkYAMLString(validYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the same YAML string back
	if actual != validYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validYaml)
	}

	invalidYaml := `abc: [`

	actual, err = checkYAMLString(invalidYaml)
	if err == nil {
		t.Fatalf("Expected to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the invalid YAML to be shown back to us again.
	if actual != invalidYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, invalidYaml)
	}
}
