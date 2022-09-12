package identitystore

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},

			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameUser = "User Data Source"
)

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreConn

	identityStoreId := d.Get("identity_store_id").(string)

	// Filters has been marked as deprecated in favour of GetUserId, which
	// allows only a single filter. Keep using it to maintain backwards
	// compatibility of the data source.

	input := &identitystore.ListUsersInput{
		IdentityStoreId: aws.String(identityStoreId),
		Filters:         expandFilters(d.Get("filter").(*schema.Set).List()),
	}

	var results []types.User

	paginator := identitystore.NewListUsersPaginator(conn, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreId, err)
		}

		for _, user := range page.Users {
			if v, ok := d.GetOk("user_id"); ok && v.(string) != aws.ToString(user.UserId) {
				continue
			}

			results = append(results, user)
		}
	}

	if len(results) == 0 {
		return diag.Errorf("no Identity Store User found matching criteria\n%v; try different search", input.Filters)
	}

	if len(results) > 1 {
		return diag.Errorf("multiple Identity Store Users found matching criteria\n%v; try different search", input.Filters)
	}

	user := results[0]

	d.SetId(aws.ToString(user.UserId))
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return nil
}
