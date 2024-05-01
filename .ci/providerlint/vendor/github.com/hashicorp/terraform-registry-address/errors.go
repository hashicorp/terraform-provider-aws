// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfaddr

import (
	"fmt"
)

type ParserError struct {
	Summary string
	Detail  string
}

func (pe *ParserError) Error() string {
	return fmt.Sprintf("%s: %s", pe.Summary, pe.Detail)
}
