// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validNamespaceName = validation.All(
	validation.StringLenBetween(1, 1024),
	validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z._-]+$`), ""),
)
