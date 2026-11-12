// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
)

func stringMatchDiag(pattern string) {
	// ruleid: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(pattern, "must match pattern")

	// ruleid: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(`^[a-z]+$`, "must contain lowercase letters")

	// ruleid: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(pattern, "must match pattern")

	// ruleid: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(`^[0-9]+$`, "must contain digits")

	// ok: use-lazyvalidation-string-match-diag
	validation.ToDiagFunc(validation.StringLenBetween(1, 128))

	// ok: use-lazyvalidation-string-match-diag
	lazyvalidation.StringMatchDiag(pattern, "must match pattern")

	// ok: use-lazyvalidation-string-match
	lazyvalidation.StringMatch(pattern, "must match pattern")
}
