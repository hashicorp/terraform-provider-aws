// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func suppressOpenIDURL(k, old, new string, d *schema.ResourceData) bool {
	oldURL, err := url.Parse(normalizeOpenIDURL(old))
	if err != nil {
		return false
	}

	newURL, err := url.Parse(normalizeOpenIDURL(new))
	if err != nil {
		return false
	}

	return oldURL.String() == newURL.String()
}
