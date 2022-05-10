package keyspaces

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"regexp"
)

func ResourceTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyspaceCreate,
		ReadWithoutTimeout:   resourceKeyspaceRead,
		UpdateWithoutTimeout: resourceKeyspaceUpdate,
		DeleteWithoutTimeout: resourceKeyspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// To be updated
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"keyspace_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 48),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,47}$`),
						"The name must consist of alphanumerics and underscores.",
					),
				),
			},
			"table_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 48),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,47}$`),
						"The name must consist of alphanumerics and underscores.",
					),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}
