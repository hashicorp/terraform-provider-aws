// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validNamespaceName = validation.All(
	validation.StringLenBetween(1, 1024),
	validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z._-]+$`), "valid characters: a-z, A-Z, 0-9, . (period), _ (underscore), and - (hyphen)"),
)
