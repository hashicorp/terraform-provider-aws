// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

type FirstError struct{}

func (e FirstError) Error() string {
	return "First Error"
}

func (e FirstError) ErrorMessage() string {
	return "First ErrorMessage"
}

type SecondError struct{}

func (e *SecondError) Error() string {
	return "Second Error"
}

func (e *SecondError) ErrorMessage() string {
	return "Second ErrorMessage"
}

func TestIsAErrorMessageContains(t *testing.T) {
	t.Parallel()

	var e1 FirstError
	var e2 SecondError

	if !errs.IsAErrorMessageContains[FirstError](e1, "First") {
		t.Error("unexpected false")
	}

	if errs.IsAErrorMessageContains[FirstError](e1, "Second") {
		t.Error("unexpected true")
	}

	if errs.IsAErrorMessageContains[*SecondError](e1, "First") {
		t.Error("unexpected true")
	}

	if !errs.IsAErrorMessageContains[*SecondError](&e2, "Second") {
		t.Error("unexpected false")
	}
}
