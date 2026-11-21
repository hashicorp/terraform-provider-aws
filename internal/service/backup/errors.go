// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup

const (
	errCodeAccessDeniedException          = "AccessDeniedException"
	errCodeInvalidParameterValueException = "InvalidParameterValueException"

	// AWS Backup reports a vault that does not exist as AccessDeniedException with this message.
	// An IAM or SCP deny uses the same code with a different message.
	errMessageVaultNotFound = "Insufficient privileges to perform this action"
)
