package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsidentityStoreGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsidentityStoreGroupRead,

		Schema: map[string]*schema.Schema{
			"identity_store_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},

			"group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"display_name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"display_name": {
				Type:     schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"group_id"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{S}\p{N}\p{P}\t\n\r ]+$`), "must match [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}\\t\\n\\r ]"),
				),
			},
		},
	}
}

func dataSourceAwsidentityStoreGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).identitystoreconn
	// TODO
	return nil
}
