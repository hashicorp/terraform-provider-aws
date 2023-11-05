package connectcases

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_connectcases_contact_case", name="Connect Cases Contact Case")
func ResourceContactCase() *schema.Resource {
	return &schema.Resource{
		// CreateWithoutTimeout: resourceContactCaseCreate,
		// ReadWithoutTimeout:   resourceContactCaseRead,
		// UpdateWithoutTimeout: resourceContactCaseUpdate,
		// DeleteWithoutTimeout: resourceContactCaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_document": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}
