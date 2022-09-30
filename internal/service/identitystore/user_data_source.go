package identitystore

import (
	"context"
	"errors"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
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
			"external_id": {
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      1,
				AtLeastOneOf:  []string{"external_id", "filter", "user_id"},
				ConflictsWith: []string{"filter"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"issuer": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      1,
				AtLeastOneOf:  []string{"external_id", "filter", "user_id"},
				ConflictsWith: []string{"external_id"},
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"external_id", "filter", "user_id"},
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

	if v, ok := d.GetOk("filter"); ok && v.(*schema.Set).Len() > 0 {
		input := &identitystore.GetUserIdInput{
			AlternateIdentifier: &types.AlternateIdentifierMemberUniqueAttribute{
				Value: *expandUniqueAttribute(v.(*schema.Set).List()[0].(map[string]interface{})),
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
	} else if v, ok := d.GetOk("external_id"); ok && v.(*schema.Set).Len() > 0 {
		input := &identitystore.GetUserIdInput{
			AlternateIdentifier: &types.AlternateIdentifierMemberExternalId{
				Value: *expandExternalId(v.(*schema.Set).List()[0].(map[string]interface{})),
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

	if v, ok := d.GetOk("user_id"); ok && v.(string) != "" {
		if userId != "" && userId != v.(string) {
			// We were given a filter, and it found a user different to this one.
			return diag.Errorf("no Identity Store User found matching criteria; try different search")
		}

		userId = v.(string)
	}

	user, err := findUserByID(ctx, conn, identityStoreId, userId)

	if err != nil {
		if _, ok := err.(*resource.NotFoundError); ok {
			return diag.Errorf("no Identity Store User found matching criteria; try different search")
		}

		return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreId, err)
	}

	d.SetId(userId)
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return nil
}
