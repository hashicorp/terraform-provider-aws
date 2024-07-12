// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cognito_user", name="User")
func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUserCreate,
		ReadWithoutTimeout:   resourceUserRead,
		UpdateWithoutTimeout: resourceUserUpdate,
		DeleteWithoutTimeout: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), userResourceIDSeparator)

				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sUsername", d.Id(), userResourceIDSeparator)
				}

				d.Set(names.AttrUserPoolID, parts[0])
				d.Set(names.AttrUsername, parts[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrAttributes: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if k == "attributes.sub" || k == "attributes.%" {
						return true
					}

					return false
				},
			},
			"client_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"desired_delivery_mediums": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.DeliveryMediumType](),
				},
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"force_alias_creation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message_action": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.MessageActionType](),
			},
			"mfa_setting_list": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrPassword: {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ValidateFunc:  validation.StringLenBetween(6, 256),
				ConflictsWith: []string{"temporary_password"},
			},
			"preferred_mfa_setting": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sub": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"temporary_password": {
				Type:          schema.TypeString,
				Sensitive:     true,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(6, 256),
				ConflictsWith: []string{names.AttrPassword},
			},
			names.AttrUserPoolID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrUsername: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"validation_data": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID := d.Get(names.AttrUserPoolID).(string)
	username := d.Get(names.AttrUsername).(string)
	id := userCreateResourceID(userPoolID, username)
	input := &cognitoidentityprovider.AdminCreateUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(userPoolID),
	}

	if v, ok := d.GetOk("client_metadata"); ok {
		input.ClientMetadata = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("desired_delivery_mediums"); ok {
		input.DesiredDeliveryMediums = flex.ExpandStringyValueSet[awstypes.DeliveryMediumType](v.(*schema.Set))
	}

	if v, ok := d.GetOk("force_alias_creation"); ok {
		input.ForceAliasCreation = v.(bool)
	}

	if v, ok := d.GetOk("message_action"); ok {
		input.MessageAction = awstypes.MessageActionType(v.(string))
	}

	if v, ok := d.GetOk("temporary_password"); ok {
		input.TemporaryPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrAttributes); ok {
		input.UserAttributes = expandAttributeTypes(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("validation_data"); ok {
		input.ValidationData = expandAttributeTypes(v.(map[string]interface{}))
	}

	_, err := conn.AdminCreateUser(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cognito User (%s): %s", id, err)
	}

	d.SetId(id)

	if v := d.Get(names.AttrEnabled); !v.(bool) {
		input := &cognitoidentityprovider.AdminDisableUserInput{
			Username:   aws.String(username),
			UserPoolId: aws.String(userPoolID),
		}

		_, err := conn.AdminDisableUser(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "disabling Cognito User (%s): %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk(names.AttrPassword); ok {
		input := &cognitoidentityprovider.AdminSetUserPasswordInput{
			Password:   aws.String(v.(string)),
			Permanent:  true,
			Username:   aws.String(username),
			UserPoolId: aws.String(userPoolID),
		}

		_, err := conn.AdminSetUserPassword(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Cognito User (%s) password: %s", d.Id(), err)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, username := d.Get(names.AttrUserPoolID).(string), d.Get(names.AttrUsername).(string)
	user, err := findUserByTwoPartKey(ctx, conn, userPoolID, username)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cognito User %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cognito User (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrAttributes, flattenAttributeTypes(user.UserAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attributes: %s", err)
	}
	d.Set(names.AttrCreationDate, user.UserCreateDate.Format(time.RFC3339))
	d.Set(names.AttrEnabled, user.Enabled)
	d.Set("last_modified_date", user.UserLastModifiedDate.Format(time.RFC3339))
	if err := d.Set("mfa_setting_list", user.UserMFASettingList); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mfa_setting_list: %s", err)
	}
	d.Set("preferred_mfa_setting", user.PreferredMfaSetting)
	d.Set(names.AttrStatus, user.UserStatus)
	d.Set("sub", flattenUserSub(user.UserAttributes))
	d.Set(names.AttrUserPoolID, userPoolID)
	d.Set(names.AttrUsername, username)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, username := d.Get(names.AttrUserPoolID).(string), d.Get(names.AttrUsername).(string)

	if d.HasChange(names.AttrAttributes) {
		o, n := d.GetChange(names.AttrAttributes)
		upd, del := expandUpdateUserAttributes(o.(map[string]interface{}), n.(map[string]interface{}))

		if len(upd) > 0 {
			input := &cognitoidentityprovider.AdminUpdateUserAttributesInput{
				Username:       aws.String(username),
				UserAttributes: expandAttributeTypes(upd),
				UserPoolId:     aws.String(userPoolID),
			}

			if v, ok := d.GetOk("client_metadata"); ok {
				input.ClientMetadata = flex.ExpandStringValueMap(v.(map[string]interface{}))
			}

			_, err := conn.AdminUpdateUserAttributes(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Cognito User (%s) attributes: %s", d.Id(), err)
			}
		}

		if len(del) > 0 {
			input := &cognitoidentityprovider.AdminDeleteUserAttributesInput{
				Username:           aws.String(username),
				UserAttributeNames: tfslices.ApplyToAll(del, normalizeUserAttributeKey),
				UserPoolId:         aws.String(userPoolID),
			}

			_, err := conn.AdminDeleteUserAttributes(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Cognito User (%s) attributes: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrEnabled) {
		if d.Get(names.AttrEnabled).(bool) {
			input := &cognitoidentityprovider.AdminEnableUserInput{
				Username:   aws.String(username),
				UserPoolId: aws.String(userPoolID),
			}

			_, err := conn.AdminEnableUser(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling Cognito User (%s): %s", d.Id(), err)
			}
		} else {
			input := &cognitoidentityprovider.AdminDisableUserInput{
				Username:   aws.String(username),
				UserPoolId: aws.String(userPoolID),
			}

			_, err := conn.AdminDisableUser(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Cognito User (%s): %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("temporary_password") {
		if v := d.Get("temporary_password").(string); v != "" {
			input := &cognitoidentityprovider.AdminSetUserPasswordInput{
				Password:   aws.String(v),
				Permanent:  false,
				Username:   aws.String(username),
				UserPoolId: aws.String(userPoolID),
			}

			_, err := conn.AdminSetUserPassword(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Cognito User (%s) password: %s", d.Id(), err)
			}
		} else {
			d.Set("temporary_password", nil)
		}
	}

	if d.HasChange(names.AttrPassword) {
		if v := d.Get(names.AttrPassword).(string); v != "" {
			input := &cognitoidentityprovider.AdminSetUserPasswordInput{
				Password:   aws.String(v),
				Permanent:  true,
				Username:   aws.String(username),
				UserPoolId: aws.String(userPoolID),
			}

			_, err := conn.AdminSetUserPassword(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Cognito User (%s) password: %s", d.Id(), err)
			}
		} else {
			d.Set(names.AttrPassword, nil)
		}
	}

	return append(diags, resourceUserRead(ctx, d, meta)...)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CognitoIDPClient(ctx)

	userPoolID, username := d.Get(names.AttrUserPoolID).(string), d.Get(names.AttrUsername).(string)

	log.Printf("[DEBUG] Deleting Cognito User: %s", d.Id())
	_, err := conn.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(userPoolID),
	})

	if errs.IsA[*awstypes.UserNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cognito User (%s): %s", d.Id(), err)
	}

	return diags
}

const userResourceIDSeparator = "/"

func userCreateResourceID(userPoolID, username string) string {
	parts := []string{userPoolID, username}
	id := strings.Join(parts, userResourceIDSeparator)

	return id
}

// No userParseResourceID as pre-v5.56.0, the ID wasn't parsed -- user_pool_id and username attribute were used directly.

func findUserByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, username string) (*cognitoidentityprovider.AdminGetUserOutput, error) {
	input := &cognitoidentityprovider.AdminGetUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.AdminGetUser(ctx, input)

	if errs.IsA[*awstypes.UserNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func expandAttributeTypes(tfMap map[string]interface{}) []awstypes.AttributeType {
	if len(tfMap) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.AttributeType, 0, len(tfMap))

	for k, v := range tfMap {
		apiObjects = append(apiObjects, awstypes.AttributeType{
			Name:  aws.String(normalizeUserAttributeKey(k)),
			Value: aws.String(v.(string)),
		})
	}

	return apiObjects
}

func flattenAttributeTypes(apiObjects []awstypes.AttributeType) map[string]interface{} {
	tfMap := make(map[string]interface{})

	for _, apiObject := range apiObjects {
		if apiObject.Name != nil {
			if k, v := aws.ToString(apiObject.Name), aws.ToString(apiObject.Value); userAttributeKeyMatchesStandardAttribute(k) {
				tfMap[k] = v
			} else {
				k := strings.TrimPrefix(strings.TrimPrefix(k, attributeDevPrefix), attributeCustomPrefix)
				tfMap[k] = v
			}
		}
	}

	return tfMap
}

func flattenUserSub(apiObjects []awstypes.AttributeType) *string {
	for _, apiObject := range apiObjects {
		if aws.ToString(apiObject.Name) == "sub" {
			return apiObject.Value
		}
	}

	return nil
}

func expandUpdateUserAttributes(oldMap, newMap map[string]interface{}) (map[string]interface{}, []string) {
	upd := make(map[string]interface{})

	for k, v := range newMap {
		if old, ok := oldMap[k]; ok {
			if old.(string) != v.(string) {
				upd[k] = v
			}
			delete(oldMap, k)
		} else {
			upd[k] = v
		}
	}

	del := tfmaps.Keys(oldMap)

	return upd, del
}

const (
	attributeCustomPrefix = "custom:"
	attributeDevPrefix    = "dev:"
)

func normalizeUserAttributeKey(k string) string {
	if !userAttributeKeyMatchesStandardAttribute(k) && !strings.HasPrefix(k, attributeCustomPrefix) {
		return attributeCustomPrefix + k
	}

	return k
}

func userAttributeKeyMatchesStandardAttribute(k string) bool {
	standardAttributeKeys := []string{
		names.AttrAddress,
		"birthdate",
		names.AttrEmail,
		"email_verified",
		"gender",
		"given_name",
		"family_name",
		"locale",
		"middle_name",
		names.AttrName,
		"nickname",
		"phone_number",
		"phone_number_verified",
		"picture",
		"preferred_username",
		names.AttrProfile,
		"sub",
		"updated_at",
		"website",
		"zoneinfo",
	}

	return slices.Contains(standardAttributeKeys, k)
}
