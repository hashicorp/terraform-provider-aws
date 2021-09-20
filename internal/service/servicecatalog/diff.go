package servicecatalog

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func suppressEquivalentJSONEmptyNilDiffs(k, old, new string, d *schema.ResourceData) bool {
	if old == "[]" && new == "" {
		return true
	}

	if old == "" && new == "[]" {
		return true
	}

	return verify.SuppressEquivalentJSONDiffs(k, old, new, d)
}
