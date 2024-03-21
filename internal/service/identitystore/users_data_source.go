// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/YakDriver/regexache"
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

// @SDKDataSource("aws_identitystore_users", name="Users")
func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUsersRead,

		Schema: map[string]*schema.Schema{
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"country": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"formatted": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"locality": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"postal_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"primary": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"street_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"emails": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"external_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"identity_store_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"locale": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"family_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"formatted": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"given_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"honorific_prefix": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"honorific_suffix": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"middle_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"nickname": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"phone_numbers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"value": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"preferred_language": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"profile_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timezone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

const (
	DSNameUsers = "Users Data Source"
)

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	input := &identitystore.ListUsersInput{
		IdentityStoreId: aws.String(identityStoreID),
	}

	paginator := identitystore.NewListUsersPaginator(conn, input)
	users := make([]types.User, 0)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return create.DiagError(names.IdentityStore, create.ErrActionReading, DSNameUsers, identityStoreID, err)
		}

		users = append(users, page.Users...)
	}

	d.SetId(identityStoreID)

	if err := d.Set("users", flattenUsers(users)); err != nil {
		return create.DiagError(names.IdentityStore, create.ErrActionSetting, DSNameUsers, d.Id(), err)
	}

	return nil
}

func flattenUsers(users []types.User) []map[string]interface{} {
	if len(users) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	for _, user := range users {
		u := make(map[string]interface{}, 1)

		u["display_name"] = aws.ToString(user.DisplayName)
		u["identity_store_id"] = aws.ToString(user.IdentityStoreId)
		u["locale"] = aws.ToString(user.Locale)
		u["nickname"] = aws.ToString(user.NickName)
		u["preferred_language"] = aws.ToString(user.PreferredLanguage)
		u["profile_url"] = aws.ToString(user.ProfileUrl)
		u["timezone"] = aws.ToString(user.Timezone)
		u["title"] = aws.ToString(user.Title)
		u["user_id"] = aws.ToString(user.UserId)
		u["user_name"] = aws.ToString(user.UserName)
		u["user_type"] = aws.ToString(user.UserType)

		u["addresses"] = flattenAddresses(user.Addresses)
		u["emails"] = flattenEmails(user.Emails)
		u["external_ids"] = flattenExternalIds(user.ExternalIds)
		u["name"] = []interface{}{flattenName(user.Name)}
		u["phone_numbers"] = flattenPhoneNumbers(user.PhoneNumbers)

		result = append(result, u)
	}

	return result
}
