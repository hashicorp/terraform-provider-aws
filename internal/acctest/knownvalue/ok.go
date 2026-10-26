// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

var (
	_ knownvalue.Check = ok{}
)

type ok struct{}

func (ok) CheckValue(any) error {
	return nil
}

func (ok) String() string {
	return "ok"
}

// OK returns a Check that always passes.
// Used to verify that a value is present, but not to verify the value itself.
func OK() knownvalue.Check {
	return ok{}
}
