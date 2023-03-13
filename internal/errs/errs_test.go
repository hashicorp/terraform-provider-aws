package errs_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

type Error1 struct{}

func (e Error1) Error() string {
	return "Error1 Error"
}

func (e Error1) ErrorMessage() string {
	return "Error1 ErrorMessage"
}

type Error2 struct{}

func (e *Error2) Error() string {
	return "Error2 Error"
}

func (e *Error2) ErrorMessage() string {
	return "Error2 ErrorMessage"
}

func TestIsAErrorMessageContains(t *testing.T) {
	t.Parallel()

	var e1 Error1
	var e2 Error2

	if !errs.IsAErrorMessageContains[Error1](e1, "Error1") {
		t.Error("unexpected false")
	}

	if errs.IsAErrorMessageContains[Error1](e1, "Error2") {
		t.Error("unexpected true")
	}

	if errs.IsAErrorMessageContains[*Error2](e1, "Error1") {
		t.Error("unexpected true")
	}

	if !errs.IsAErrorMessageContains[*Error2](&e2, "Error2") {
		t.Error("unexpected false")
	}
}
