// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_user")
func DataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceUserRead,

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
						names.AttrRegion: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"street_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"alternate_identifier": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExternalID: {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrID: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrIssuer: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"unique_attribute": {
							Type:         schema.TypeList,
							Optional:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"alternate_identifier.0.external_id", "alternate_identifier.0.unique_attribute"},
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
					},
				},
				ConflictsWith: []string{names.AttrFilter, "user_id"},
			},
			names.AttrDisplayName: {
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
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
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
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: {
				Deprecated:    "Use the alternate_identifier attribute instead.",
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				AtLeastOneOf:  []string{"alternate_identifier", names.AttrFilter, "user_id"},
				ConflictsWith: []string{"alternate_identifier"},
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]*$`), "must match [0-9A-Za-z-]"),
				),
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
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
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexache.MustCompile(`^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`), "must match ([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"),
				),
				AtLeastOneOf:  []string{"alternate_identifier", names.AttrFilter, "user_id"},
				ConflictsWith: []string{"alternate_identifier"},
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_type": {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	if v, ok := d.GetOk(names.AttrFilter); ok && len(v.([]interface{})) > 0 {
		// Use ListUsers for backwards compat.
		input := &identitystore.ListUsersInput{
			Filters:         expandFilters(d.Get(names.AttrFilter).([]interface{})),
			IdentityStoreId: aws.String(identityStoreID),
		}
		paginator := identitystore.NewListUsersPaginator(conn, input)
		var results []types.User

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)

			if err != nil {
				return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreID, err)
			}

			for _, user := range page.Users {
				if v, ok := d.GetOk("user_id"); ok && v.(string) != aws.ToString(user.UserId) {
					continue
				}

				results = append(results, user)
			}
		}

		if len(results) == 0 {
			return sdkdiag.AppendErrorf(diags, "no Identity Store User found matching criteria\n%v; try different search", input.Filters)
		}

		if len(results) > 1 {
			return sdkdiag.AppendErrorf(diags, "multiple Identity Store Users found matching criteria\n%v; try different search", input.Filters)
		}

		user := results[0]

		d.SetId(aws.ToString(user.UserId))
		d.Set(names.AttrDisplayName, user.DisplayName)
		d.Set("identity_store_id", user.IdentityStoreId)
		d.Set("locale", user.Locale)
		d.Set("nickname", user.NickName)
		d.Set("preferred_language", user.PreferredLanguage)
		d.Set("profile_url", user.ProfileUrl)
		d.Set("timezone", user.Timezone)
		d.Set("title", user.Title)
		d.Set("user_id", user.UserId)
		d.Set(names.AttrUserName, user.UserName)
		d.Set("user_type", user.UserType)

		if err := d.Set("addresses", flattenAddresses(user.Addresses)); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
		}

		if err := d.Set("emails", flattenEmails(user.Emails)); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
		}

		if err := d.Set("external_ids", flattenExternalIds(user.ExternalIds)); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
		}

		if err := d.Set(names.AttrName, []interface{}{flattenName(user.Name)}); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
		}

		if err := d.Set("phone_numbers", flattenPhoneNumbers(user.PhoneNumbers)); err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
		}

		return diags
	}

	var userID string

	if v, ok := d.GetOk("alternate_identifier"); ok && len(v.([]interface{})) > 0 {
		input := &identitystore.GetUserIdInput{
			AlternateIdentifier: expandAlternateIdentifier(v.([]interface{})[0].(map[string]interface{})),
			IdentityStoreId:     aws.String(identityStoreID),
		}

		output, err := conn.GetUserId(ctx, input)

		if err != nil {
			var e *types.ResourceNotFoundException
			if errors.As(err, &e) {
				return sdkdiag.AppendErrorf(diags, "no Identity Store User found matching criteria; try different search")
			} else {
				return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreID, err)
			}
		}

		userID = aws.ToString(output.UserId)
	}

	if v, ok := d.GetOk("user_id"); ok && v.(string) != "" {
		if userID != "" && userID != v.(string) {
			// We were given a filter, and it found a user different to this one.
			return sdkdiag.AppendErrorf(diags, "no Identity Store User found matching criteria; try different search")
		}

		userID = v.(string)
	}

	user, err := FindUserByTwoPartKey(ctx, conn, identityStoreID, userID)

	if err != nil {
		if tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "no Identity Store User found matching criteria; try different search")
		}

		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, DSNameUser, identityStoreID, err)
	}

	d.SetId(aws.ToString(user.UserId))

	d.Set(names.AttrDisplayName, user.DisplayName)
	d.Set("identity_store_id", user.IdentityStoreId)
	d.Set("locale", user.Locale)
	d.Set("nickname", user.NickName)
	d.Set("preferred_language", user.PreferredLanguage)
	d.Set("profile_url", user.ProfileUrl)
	d.Set("timezone", user.Timezone)
	d.Set("title", user.Title)
	d.Set("user_id", user.UserId)
	d.Set(names.AttrUserName, user.UserName)
	d.Set("user_type", user.UserType)

	if err := d.Set("addresses", flattenAddresses(user.Addresses)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
	}

	if err := d.Set("emails", flattenEmails(user.Emails)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
	}

	if err := d.Set("external_ids", flattenExternalIds(user.ExternalIds)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
	}

	if err := d.Set(names.AttrName, []interface{}{flattenName(user.Name)}); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
	}

	if err := d.Set("phone_numbers", flattenPhoneNumbers(user.PhoneNumbers)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, DSNameUser, d.Id(), err)
	}

	return diags
}
