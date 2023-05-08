package iam

import (
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func suppressOpenIDURL(k, old, new string, d *schema.ResourceData) bool {
	oldUrl, err := url.Parse(old)
	if err != nil {
		return false
	}

	newUrl, err := url.Parse(new)
	if err != nil {
		return false
	}

	oldUrl.Scheme = "https"

	return oldUrl.String() == newUrl.String()
}
