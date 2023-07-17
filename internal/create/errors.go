// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"errors"
	"fmt"
	"log"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ErrActionChecking             = "checking"
	ErrActionCheckingDestroyed    = "checking destroyed"
	ErrActionCheckingExistence    = "checking existence"
	ErrActionCheckingNotRecreated = "checking not recreated"
	ErrActionCheckingRecreated    = "checking recreated"
	ErrActionCreating             = "creating"
	ErrActionDeleting             = "deleting"
	ErrActionImporting            = "importing"
	ErrActionReading              = "reading"
	ErrActionSetting              = "setting"
	ErrActionUpdating             = "updating"
	ErrActionWaitingForCreation   = "waiting for creation"
	ErrActionWaitingForDeletion   = "waiting for delete"
	ErrActionWaitingForUpdate     = "waiting for update"
	ErrActionExpandingResourceId  = "expanding resource id"
	ErrActionFlatteningResourceId = "flattening resource id"
)

// ProblemStandardMessage is a standardized message for reporting errors and warnings
func ProblemStandardMessage(service, action, resource, id string, gotError error) string {
	hf, err := names.FullHumanFriendly(service)

	if err != nil {
		return fmt.Sprintf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %s", service, action, resource, id, gotError, err)
	}

	if gotError == nil {
		return fmt.Sprintf("%s %s %s (%s)", action, hf, resource, id)
	}

	return fmt.Sprintf("%s %s %s (%s): %s", action, hf, resource, id, gotError)
}

// Error returns an errors.Error with a standardized error message
func Error(service, action, resource, id string, gotError error) error {
	return errors.New(ProblemStandardMessage(service, action, resource, id, gotError))
}

// AddError returns diag.Diagnostics with an additional diag.Diagnostic containing
// an error using a standardized problem message
func AddError(diags diag.Diagnostics, service, action, resource, id string, gotError error) diag.Diagnostics {
	return append(diags, newError(service, action, resource, id, gotError))
}

// DiagError returns a 1-length diag.Diagnostics with a diag.Error-level diag.Diagnostic
// with a standardized error message
func DiagError(service, action, resource, id string, gotError error) diag.Diagnostics {
	return diag.Diagnostics{
		newError(service, action, resource, id, gotError),
	}
}

func newError(service, action, resource, id string, gotError error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  ProblemStandardMessage(service, action, resource, id, gotError),
	}
}

func DiagErrorFramework(service, action, resource, id string, gotError error) fwdiag.Diagnostic {
	return fwdiag.NewErrorDiagnostic(
		ProblemStandardMessage(service, action, resource, id, nil),
		gotError.Error(),
	)
}

func DiagErrorMessage(service, action, resource, id, message string) diag.Diagnostics {
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  ProblemStandardMessage(service, action, resource, id, fmt.Errorf(message)),
		},
	}
}

// ErrorSetting returns an errors.Error with a standardized error message when setting
// arguments and attributes values.
func SettingError(service, resource, id, argument string, gotError error) error {
	return errors.New(ProblemStandardMessage(service, fmt.Sprintf("%s %s", ErrActionSetting, argument), resource, id, gotError))
}

// DiagSettingError returns an errors.Error with a standardized error message when setting
// arguments and attributes values.
func DiagSettingError(service, resource, id, argument string, gotError error) diag.Diagnostics {
	return DiagError(service, fmt.Sprintf("%s %s", ErrActionSetting, argument), resource, id, gotError)
}

// AddWarning returns diag.Diagnostics with an additional diag.Diagnostic containing
// a warning using a standardized problem message
func AddWarning(diags diag.Diagnostics, service, action, resource, id string, gotError error) diag.Diagnostics {
	return append(diags,
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  ProblemStandardMessage(service, action, resource, id, gotError),
		},
	)
}

func AddWarningMessage(diags diag.Diagnostics, service, action, resource, id, message string) diag.Diagnostics {
	return append(diags,
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  ProblemStandardMessage(service, action, resource, id, fmt.Errorf(message)),
		},
	)
}

// AddWarningNotFoundRemoveState returns diag.Diagnostics with an additional diag.Diagnostic containing
// a warning using a standardized problem message
func AddWarningNotFoundRemoveState(service, action, resource, id string) diag.Diagnostics {
	return append(diag.Diagnostics{},
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  ProblemStandardMessage(service, action, resource, id, errors.New("not found, removing from state")),
		},
	)
}

// WarnLog logs to the default logger a standardized problem message
func WarnLog(service, action, resource, id string, gotError error) {
	log.Printf("[WARN] %s", ProblemStandardMessage(service, action, resource, id, gotError))
}

func LogNotFoundRemoveState(service, action, resource, id string) {
	WarnLog(service, action, resource, id, errors.New("not found, removing from state"))
}
