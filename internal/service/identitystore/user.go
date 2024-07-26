// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_identitystore_user")
func ResourceUser() *schema.Resource {
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

const (
	ResNameUser = "User"
)

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	in := &identitystore.CreateUserInput{
		DisplayName:     aws.String(d.Get(names.AttrDisplayName).(string)),
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserName:        aws.String(d.Get(names.AttrUserName).(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && len(v.([]interface{})) > 0 {
		in.Addresses = expandAddresses(v.([]interface{}))
	}

	if v, ok := d.GetOk("emails"); ok && len(v.([]interface{})) > 0 {
		in.Emails = expandEmails(v.([]interface{}))
	}

	if v, ok := d.GetOk("locale"); ok {
		in.Locale = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Name = expandName(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("phone_numbers"); ok && len(v.([]interface{})) > 0 {
		in.PhoneNumbers = expandPhoneNumbers(v.([]interface{}))
	}

	if v, ok := d.GetOk("nickname"); ok {
		in.NickName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_language"); ok {
		in.PreferredLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("profile_url"); ok {
		in.ProfileUrl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timezone"); ok {
		in.Timezone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("title"); ok {
		in.Title = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_type"); ok {
		in.UserType = aws.String(v.(string))
	}

	out, err := conn.CreateUser(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionCreating, ResNameUser, d.Get("identity_store_id").(string), err)
	}

	if out == nil || out.UserId == nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionCreating, ResNameUser, d.Get("identity_store_id").(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s/%s", aws.ToString(out.IdentityStoreId), aws.ToString(out.UserId)))

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	identityStoreId, userId, err := resourceUserParseID(d.Id())

	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, ResNameUser, d.Id(), err)
	}

	out, err := FindUserByTwoPartKey(ctx, conn, identityStoreId, userId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IdentityStore User (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionReading, ResNameUser, d.Id(), err)
	}

	d.Set(names.AttrDisplayName, out.DisplayName)
	d.Set("identity_store_id", out.IdentityStoreId)
	d.Set("locale", out.Locale)
	d.Set("nickname", out.NickName)
	d.Set("preferred_language", out.PreferredLanguage)
	d.Set("profile_url", out.ProfileUrl)
	d.Set("timezone", out.Timezone)
	d.Set("title", out.Title)
	d.Set("user_id", out.UserId)
	d.Set(names.AttrUserName, out.UserName)
	d.Set("user_type", out.UserType)

	if err := d.Set("addresses", flattenAddresses(out.Addresses)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	if err := d.Set("emails", flattenEmails(out.Emails)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	if err := d.Set("external_ids", flattenExternalIds(out.ExternalIds)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	if err := d.Set(names.AttrName, []interface{}{flattenName(out.Name)}); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	if err := d.Set("phone_numbers", flattenPhoneNumbers(out.PhoneNumbers)); err != nil {
		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionSetting, ResNameUser, d.Id(), err)
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	in := &identitystore.UpdateUserInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserId:          aws.String(d.Get("user_id").(string)),
		Operations:      nil,
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
		Expand func(interface{}) interface{}
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
			Expand: func(value interface{}) interface{} {
				addresses := expandAddresses(value.([]interface{}))

				var result []interface{}

				// The API requires a null to unset the list, so in the case
				// of no addresses, a nil result is preferable.
				for _, address := range addresses {
					m := map[string]interface{}{}

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
			Expand: func(value interface{}) interface{} {
				emails := expandEmails(value.([]interface{}))

				var result []interface{}

				// The API requires a null to unset the list, so in the case
				// of no emails, a nil result is preferable.
				for _, email := range emails {
					m := map[string]interface{}{}

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
			Expand: func(value interface{}) interface{} {
				emails := expandPhoneNumbers(value.([]interface{}))

				var result []interface{}

				// The API requires a null to unset the list, so in the case
				// of no emails, a nil result is preferable.
				for _, email := range emails {
					m := map[string]interface{}{}

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

			in.Operations = append(in.Operations, types.AttributeOperation{
				AttributePath:  aws.String(fieldToUpdate.Field),
				AttributeValue: document.NewLazyDocument(value),
			})
		}
	}

	if len(in.Operations) > 0 {
		log.Printf("[DEBUG] Updating IdentityStore User (%s): %#v", d.Id(), in)
		_, err := conn.UpdateUser(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionUpdating, ResNameUser, d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).IdentityStoreClient(ctx)

	log.Printf("[INFO] Deleting IdentityStore User %s", d.Id())

	_, err := conn.DeleteUser(ctx, &identitystore.DeleteUserInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		UserId:          aws.String(d.Get("user_id").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.IdentityStore, create.ErrActionDeleting, ResNameUser, d.Id(), err)
	}

	return diags
}

func FindUserByTwoPartKey(ctx context.Context, conn *identitystore.Client, identityStoreID, userID string) (*identitystore.DescribeUserOutput, error) {
	in := &identitystore.DescribeUserInput{
		IdentityStoreId: aws.String(identityStoreID),
		UserId:          aws.String(userID),
	}

	out, err := conn.DescribeUser(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.UserId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func resourceUserParseID(id string) (identityStoreId, userId string, err error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = errors.New("expected a resource id in the form: identity-store-id/user-id")
		return
	}

	return parts[0], parts[1], nil
}
