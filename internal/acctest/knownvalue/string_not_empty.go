// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

func StringNotEmpty() knownvalue.Check {
	return knownvalue.StringFunc(func(v string) error {
		if len(v) == 0 {
			return errors.New("expected non-empty string for StringNotEmpty check")
		}
		return nil
	})
}
