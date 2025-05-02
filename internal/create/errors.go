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
	ErrActionCancelling             = "cancelling"
	ErrActionChecking               = "checking"
	ErrActionCheckingDestroyed      = "checking destroyed"
	ErrActionCheckingExistence      = "checking existence"
	ErrActionCheckingNotRecreated   = "checking not recreated"
	ErrActionCheckingRecreated      = "checking recreated"
	ErrActionClosing                = "closing"
	ErrActionConfiguring            = "configuring"
	ErrActionCreating               = "creating"
	ErrActionDeleting               = "deleting"
	ErrActionExpandingResourceId    = "expanding resource id"
	ErrActionFlatteningResourceId   = "flattening resource id"
	ErrActionImporting              = "importing"
	ErrActionOpening                = "opening"
	ErrActionReading                = "reading"
	ErrActionRenewing               = "renewing"
	ErrActionSetting                = "setting"
	ErrActionUpdating               = "updating"
	ErrActionValidating             = "validating"
	ErrActionWaitingForCreation     = "waiting for creation"
	ErrActionWaitingForDeletion     = "waiting for delete"
	ErrActionWaitingForUpdate       = "waiting for update"
	ErrActionWaitingForCancellation = "waiting for cancellation"
)

// ProblemStandardMessage is a standardized message for reporting errors and warnings
func ProblemStandardMessage(service, action, resource, id string, gotError error) string {
	hf, err := names.FullHumanFriendly(service)

	if err != nil {
		return fmt.Sprintf("finding human-friendly name for service (%s) while creating error (%s, %s, %s, %s): %s", service, action, resource, id, gotError, err)
	}

	if gotError == nil {
		if id == "" {
			return fmt.Sprintf("%s %s %s", action, hf, resource)
		}
		return fmt.Sprintf("%s %s %s (%s)", action, hf, resource, id)
	}

	if id == "" {
		return fmt.Sprintf("%s %s %s: %s", action, hf, resource, gotError)
	}
	return fmt.Sprintf("%s %s %s (%s): %s", action, hf, resource, id, gotError)
}

func AddError(d *fwdiag.Diagnostics, service, action, resource, id string, gotError error) {
	d.AddError(
		ProblemStandardMessage(service, action, resource, id, nil),
		gotError.Error(),
	)
}

// Error returns an errors.Error with a standardized error message
func Error(service, action, resource, id string, gotError error) error {
	return errors.New(ProblemStandardMessage(service, action, resource, id, gotError))
}

// AppendDiagError returns diag.Diagnostics with an additional diag.Diagnostic containing
// an error using a standardized problem message
func AppendDiagError(diags diag.Diagnostics, service, action, resource, id string, gotError error) diag.Diagnostics {
	return append(diags,
		diagError(service, action, resource, id, gotError),
	)
}

func diagError(service, action, resource, id string, gotError error) diag.Diagnostic {
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  ProblemStandardMessage(service, action, resource, id, gotError),
	}
}

func AppendDiagSettingError(diags diag.Diagnostics, service, resource, id, argument string, gotError error) diag.Diagnostics {
	return append(diags,
		diagError(service, fmt.Sprintf("%s %s", ErrActionSetting, argument), resource, id, gotError),
	)
}

func AppendDiagWarningMessage(diags diag.Diagnostics, service, action, resource, id, message string) diag.Diagnostics {
	return append(diags,
		diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  ProblemStandardMessage(service, action, resource, id, errors.New(message)),
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
