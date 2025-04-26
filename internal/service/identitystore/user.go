// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/document"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_identitystore_user", name="User")
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"addresses": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"country": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"formatted": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"locality": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"postal_code": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrRegion: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"street_address": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
					},
				},
			},
			names.AttrDisplayName: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"emails": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						names.AttrValue: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
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
				ForceNew: true,
			},
			"locale": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			names.AttrName: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"family_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"formatted": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"given_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"honorific_prefix": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"honorific_suffix": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						"middle_name": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
					},
				},
			},
			"nickname": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"phone_numbers": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
						names.AttrValue: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
						},
					},
				},
			},
			"preferred_language": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"profile_url": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"timezone": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"title": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrUserName: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 128)),
			},
			"user_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID := d.Get("identity_store_id").(string)
	username := d.Get(names.AttrUserName).(string)
	input := &identitystore.CreateUserInput{
		DisplayName:     aws.String(d.Get(names.AttrDisplayName).(string)),
		IdentityStoreId: aws.String(identityStoreID),
		UserName:        aws.String(username),
	}

	if v, ok := d.GetOk("addresses"); ok && len(v.([]any)) > 0 {
		input.Addresses = expandAddresses(v.([]any))
	}

	if v, ok := d.GetOk("emails"); ok && len(v.([]any)) > 0 {
		input.Emails = expandEmails(v.([]any))
	}

	if v, ok := d.GetOk("locale"); ok {
		input.Locale = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Name = expandName(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("phone_numbers"); ok && len(v.([]any)) > 0 {
		input.PhoneNumbers = expandPhoneNumbers(v.([]any))
	}

	if v, ok := d.GetOk("nickname"); ok {
		input.NickName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_language"); ok {
		input.PreferredLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("profile_url"); ok {
		input.ProfileUrl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timezone"); ok {
		input.Timezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("title"); ok {
		input.Title = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_type"); ok {
		input.UserType = aws.String(v.(string))
	}

	output, err := conn.CreateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IdentityStore User (%s): %s", username, err)
	}

	d.SetId(userCreateResourceID(identityStoreID, aws.ToString(output.UserId)))

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findUserByTwoPartKey(ctx, conn, identityStoreID, userID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IdentityStore User (%s): %s", d.Id(), err)
	}

	if err := d.Set("addresses", flattenAddresses(out.Addresses)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting addresses: %s", err)
	}
	d.Set(names.AttrDisplayName, out.DisplayName)
	if err := d.Set("emails", flattenEmails(out.Emails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting emails: %s", err)
	}
	if err := d.Set("external_ids", flattenExternalIDs(out.ExternalIds)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting external_ids: %s", err)
	}
	d.Set("identity_store_id", out.IdentityStoreId)
	d.Set("locale", out.Locale)
	if err := d.Set(names.AttrName, []any{flattenName(out.Name)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name: %s", err)
	}
	d.Set("nickname", out.NickName)
	if err := d.Set("phone_numbers", flattenPhoneNumbers(out.PhoneNumbers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting phone_numbers: %s", err)
	}
	d.Set("preferred_language", out.PreferredLanguage)
	d.Set("profile_url", out.ProfileUrl)
	d.Set("timezone", out.Timezone)
	d.Set("title", out.Title)
	d.Set("user_id", out.UserId)
	d.Set(names.AttrUserName, out.UserName)
	d.Set("user_type", out.UserType)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &identitystore.UpdateUserInput{
		IdentityStoreId: aws.String(identityStoreID),
		UserId:          aws.String(userID),
	}

	// IMPLEMENTATION NOTE.
	//
	// Complex types, such as the `emails` field, don't allow field by field
	// updates, and require that the entire sub-object is modified.
	//
	// In those sub-objects, to remove a field, it must not be present at all
	// in the updated attribute value.
	//
	// However, structs such as types.Email don't specify omitempty in their
	// struct tags, so the document.NewLazyDocument marshaller will write out
	// nulls.
	//
	// This is why, for those complex fields, a custom Expand function is
	// provided that converts the Go SDK type (e.g. types.Email) into a field
	// by field representation of what the API would expect.

	fieldsToUpdate := []struct {
		// Attribute corresponds to the provider schema.
		Attribute string

		// Field corresponds to the AWS API schema.
		Field string

		// Expand, when not nil, is used to transform the value of the field
		// given in Attribute before it's passed to the UpdateOperation.
		Expand func(any) any
	}{
		{
			Attribute: names.AttrDisplayName,
			Field:     "displayName",
		},
		{
			Attribute: "locale",
			Field:     "locale",
		},
		{
			Attribute: "name.0.family_name",
			Field:     "name.familyName",
		},
		{
			Attribute: "name.0.formatted",
			Field:     "name.formatted",
		},
		{
			Attribute: "name.0.given_name",
			Field:     "name.givenName",
		},
		{
			Attribute: "name.0.honorific_prefix",
			Field:     "name.honorificPrefix",
		},
		{
			Attribute: "name.0.honorific_suffix",
			Field:     "name.honorificSuffix",
		},
		{
			Attribute: "name.0.middle_name",
			Field:     "name.middleName",
		},
		{
			Attribute: "nickname",
			Field:     "nickName",
		},
		{
			Attribute: "preferred_language",
			Field:     "preferredLanguage",
		},
		{
			Attribute: "profile_url",
			Field:     "profileUrl",
		},
		{
			Attribute: "timezone",
			Field:     "timezone",
		},
		{
			Attribute: "title",
			Field:     "title",
		},
		{
			Attribute: "user_type",
			Field:     "userType",
		},
		{
			Attribute: "addresses",
			Field:     "addresses",
			Expand: func(value any) any {
				addresses := expandAddresses(value.([]any))

				var result []any

				// The API requires a null to unset the list, so in the case
				// of no addresses, a nil result is preferable.
				for _, address := range addresses {
					m := map[string]any{}

					if v := address.Country; v != nil {
						m["country"] = v
					}

					if v := address.Formatted; v != nil {
						m["formatted"] = v
					}

					if v := address.Locality; v != nil {
						m["locality"] = v
					}

					if v := address.PostalCode; v != nil {
						m["postalCode"] = v
					}

					m["primary"] = address.Primary

					if v := address.Region; v != nil {
						m[names.AttrRegion] = v
					}

					if v := address.StreetAddress; v != nil {
						m["streetAddress"] = v
					}

					if v := address.Type; v != nil {
						m[names.AttrType] = v
					}

					result = append(result, m)
				}

				return result
			},
		},
		{
			Attribute: "emails",
			Field:     "emails",
			Expand: func(value any) any {
				emails := expandEmails(value.([]any))

				var result []any

				// The API requires a null to unset the list, so in the case
				// of no emails, a nil result is preferable.
				for _, email := range emails {
					m := map[string]any{}

					m["primary"] = email.Primary

					if v := email.Type; v != nil {
						m[names.AttrType] = v
					}

					if v := email.Value; v != nil {
						m[names.AttrValue] = v
					}

					result = append(result, m)
				}

				return result
			},
		},
		{
			Attribute: "phone_numbers",
			Field:     "phoneNumbers",
			Expand: func(value any) any {
				emails := expandPhoneNumbers(value.([]any))

				var result []any

				// The API requires a null to unset the list, so in the case
				// of no emails, a nil result is preferable.
				for _, email := range emails {
					m := map[string]any{}

					m["primary"] = email.Primary

					if v := email.Type; v != nil {
						m[names.AttrType] = v
					}

					if v := email.Value; v != nil {
						m[names.AttrValue] = v
					}

					result = append(result, m)
				}

				return result
			},
		},
	}

	for _, fieldToUpdate := range fieldsToUpdate {
		if d.HasChange(fieldToUpdate.Attribute) {
			value := d.Get(fieldToUpdate.Attribute)

			if expand := fieldToUpdate.Expand; expand != nil {
				value = expand(value)
			}

			// The API doesn't allow empty attribute values. To unset an
			// attribute, set it to null.
			if value == "" {
				value = nil
			}

			input.Operations = append(input.Operations, types.AttributeOperation{
				AttributePath:  aws.String(fieldToUpdate.Field),
				AttributeValue: document.NewLazyDocument(value),
			})
		}
	}

	if len(input.Operations) > 0 {
		_, err := conn.UpdateUser(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IdentityStore User (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreID, userID, err := userParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IdentityStore User: %s", d.Id())
	_, err = conn.DeleteUser(ctx, &identitystore.DeleteUserInput{
		IdentityStoreId: aws.String(identityStoreID),
		UserId:          aws.String(userID),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IdentityStore User (%s): %s", d.Id(), err)
	}

	return diags
}

const userResourceIDSeparator = "/"

func userCreateResourceID(identityStoreID, userID string) string {
	parts := []string{identityStoreID, userID}
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

func userParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, userResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected identity-store-id%[2]sgroup-id", id, groupResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findUserByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, userID string) (*identitystore.DescribeUserOutput, error) {
	input := &identitystore.DescribeUserInput{
		IdentityStoreId: aws.String(identityStoreID),
		UserId:          aws.String(userID),
	}

	return findUser(ctx, conn, input)
}

func findUser(ctx context.Context, conn *identitystore.Client, input *identitystore.DescribeUserInput) (*identitystore.DescribeUserOutput, error) {
	output, err := conn.DescribeUser(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
