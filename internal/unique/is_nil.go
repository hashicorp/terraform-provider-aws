// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package unique

import (
	"unique"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// IsHandleNil checks whether a Handle has had a value assigned.
func IsHandleNil[T comparable](h unique.Handle[T]) bool {
	return isZero(h)
}

func isZero[T comparable](v T) bool {
	return v == inttypes.Zero[T]()
}
