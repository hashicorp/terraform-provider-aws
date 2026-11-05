// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func suppressOpenIDURL(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimPrefix(old, "https://") == strings.TrimPrefix(new, "https://")
}
