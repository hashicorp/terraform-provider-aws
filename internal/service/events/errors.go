// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

const (
	errCodeValidationException = "ValidationException"
)

const (
	// KMS key policy changes are eventually consistent, so EventBridge can be denied kms:Decrypt shortly after the policy grants it.
	errMessageKMSAccessDenied = "KMS returned an access denied exception"
)
