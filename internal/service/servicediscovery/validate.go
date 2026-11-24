// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
)

var validNamespaceName = validation.All(
	validation.StringLenBetween(1, 1024),
	lazyvalidation.StringMatch(`^[0-9A-Za-z._-]+$`, "valid characters: a-z, A-Z, 0-9, . (period), _ (underscore), and - (hyphen)"),
)
