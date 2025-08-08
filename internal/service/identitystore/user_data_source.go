// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_identitystore_user", name="User")
func dataSourceUser() *schema.Resource {
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
				ConflictsWith: []string{"user_id"},
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
				AtLeastOneOf:  []string{"alternate_identifier", "user_id"},
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

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)

	var userID string

	if v, ok := d.GetOk("alternate_identifier"); ok && len(v.([]any)) > 0 {
		input := identitystore.GetUserIdInput{
			AlternateIdentifier: expandAlternateIdentifier(v.([]any)[0].(map[string]any)),
			IdentityStoreId:     aws.String(identityStoreID),
		}

		output, err := conn.GetUserId(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading IdentityStore User (%s): %s", identityStoreID, err)
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

	user, err := findUserByTwoPartKey(ctx, conn, identityStoreID, userID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IdentityStore User (%s): %s", userID, err)
	}

	d.SetId(aws.ToString(user.UserId))
	if err := d.Set("addresses", flattenAddresses(user.Addresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting addresses: %s", err)
	}
	d.Set(names.AttrDisplayName, user.DisplayName)
	if err := d.Set("emails", flattenEmails(user.Emails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting emails: %s", err)
	}
	if err := d.Set("external_ids", flattenExternalIDs(user.ExternalIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting external_ids: %s", err)
	}
	d.Set("identity_store_id", user.IdentityStoreId)
	d.Set("locale", user.Locale)
	if err := d.Set(names.AttrName, []any{flattenName(user.Name)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name: %s", err)
	}
	d.Set("nickname", user.NickName)
	if err := d.Set("phone_numbers", flattenPhoneNumbers(user.PhoneNumbers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting phone_numbers: %s", err)
	}
	d.Set("preferred_language", user.PreferredLanguage)
	d.Set("profile_url", user.ProfileUrl)
	d.Set("timezone", user.Timezone)
	d.Set("title", user.Title)
	d.Set("user_id", user.UserId)
	d.Set(names.AttrUserName, user.UserName)
	d.Set("user_type", user.UserType)

	return diags
}
