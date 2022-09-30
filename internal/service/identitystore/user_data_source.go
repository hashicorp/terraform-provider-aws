package identitystore

import (
	"context"
	"errors"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	var userId string

	if v, ok := d.GetOk("user_id"); ok && v.(string) != "" {
		userId = v.(string)
	} else {
		var uniqueAttribute types.UniqueAttribute

		filters := expandFilters(d.Get("filter").(*schema.Set).List())

		if len(filters) > 1 {
			panic("too many filters -- must be one unique attribute or an external identifier")
		}

		if uniqueAttribute.AttributePath == nil {
			uniqueAttribute = types.UniqueAttribute{
				AttributePath:  filters[0].AttributePath,
				AttributeValue: document.NewLazyDocument(aws.ToString(filters[0].AttributeValue)),
			}
		}

		input := &identitystore.GetUserIdInput{
			AlternateIdentifier: &types.AlternateIdentifierMemberUniqueAttribute{
				Value: uniqueAttribute,
			},
			IdentityStoreId: aws.String(identityStoreId),
		}

		output, err := conn.GetUserId(ctx, input)

		if err != nil {
			var e *types.ResourceNotFoundException
			if errors.As(err, &e) {
				return diag.Errorf("no Identity Store User found matching criteria; try different search")
			} else {
				return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreId, err)
			}
		}

		userId = aws.ToString(output.UserId)
	}

	user, err := findUserByID(ctx, conn, identityStoreId, userId)

	if err != nil {
		if _, ok := err.(*resource.NotFoundError); ok {
			return diag.Errorf("no Identity Store User found matching criteria; try different search")
		}

		return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreId, err)
	}

	d.SetId(userId)
	d.Set("user_id", userId)
	d.Set("user_name", user.UserName)

	return nil
}
