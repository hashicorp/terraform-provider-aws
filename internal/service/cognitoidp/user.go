package cognitoidp

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Update: resourceUserUpdate,
		Delete: resourceUserDelete,

		Importer: &schema.ResourceImporter{
			State: resourceUserImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminCreateUser.html
		Schema: map[string]*schema.Schema{
			"attributes": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if k == "attributes.sub" || k == "attributes.%" {
						return true
					}

					return false
				},
				Optional: true,
			},
			"client_metadata": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"desired_delivery_mediums": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.DeliveryMediumType_Values(), false),
				},
				Optional: true,
			},
			"enabled": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.MessageActionType_Values(), false),
			},
			"mfa_setting_list": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"preferred_mfa_setting": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sub": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": {
				Type:          schema.TypeString,
				Sensitive:     true,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(6, 256),
				ConflictsWith: []string{"temporary_password"},
			},
			"temporary_password": {
				Type:          schema.TypeString,
				Sensitive:     true,
				Optional:      true,
				ValidateFunc:  validation.StringLenBetween(6, 256),
				ConflictsWith: []string{"password"},
			},
			"validation_data": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	username := d.Get("username").(string)
	userPoolId := d.Get("user_pool_id").(string)

	params := &cognitoidentityprovider.AdminCreateUserInput{
		Username:   aws.String(username),
		UserPoolId: aws.String(userPoolId),
	}

	if v, ok := d.GetOk("client_metadata"); ok {
		metadata := v.(map[string]interface{})
		params.ClientMetadata = expandUserClientMetadata(metadata)
	}

	if v, ok := d.GetOk("desired_delivery_mediums"); ok {
		mediums := v.(*schema.Set)
		params.DesiredDeliveryMediums = expandUserDesiredDeliveryMediums(mediums)
	}

	if v, ok := d.GetOk("force_alias_creation"); ok {
		params.ForceAliasCreation = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("message_action"); ok {
		params.MessageAction = aws.String(v.(string))
	}

	if v, ok := d.GetOk("attributes"); ok {
		attributes := v.(map[string]interface{})
		params.UserAttributes = expandAttribute(attributes)
	}

	if v, ok := d.GetOk("validation_data"); ok {
		attributes := v.(map[string]interface{})
		// aws sdk uses the same type for both validation data and user attributes
		// https://docs.aws.amazon.com/sdk-for-go/api/service/cognitoidentityprovider/#AdminCreateUserInput
		params.ValidationData = expandAttribute(attributes)
	}

	if v, ok := d.GetOk("temporary_password"); ok {
		params.TemporaryPassword = aws.String(v.(string))
	}

	log.Print("[DEBUG] Creating Cognito User")

	resp, err := conn.AdminCreateUser(params)
	if err != nil {
		return fmt.Errorf("error creating Cognito User (%s/%s): %w", userPoolId, username, err)
	}

	d.SetId(fmt.Sprintf("%s/%s", aws.StringValue(params.UserPoolId), aws.StringValue(resp.User.Username)))

	if v := d.Get("enabled"); !v.(bool) {
		disableParams := &cognitoidentityprovider.AdminDisableUserInput{
			Username:   aws.String(d.Get("username").(string)),
			UserPoolId: aws.String(d.Get("user_pool_id").(string)),
		}

		_, err := conn.AdminDisableUser(disableParams)
		if err != nil {
			return fmt.Errorf("error disabling Cognito User (%s): %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("password"); ok {
		setPasswordParams := &cognitoidentityprovider.AdminSetUserPasswordInput{
			Username:   aws.String(d.Get("username").(string)),
			UserPoolId: aws.String(d.Get("user_pool_id").(string)),
			Password:   aws.String(v.(string)),
			Permanent:  aws.Bool(true),
		}

		_, err := conn.AdminSetUserPassword(setPasswordParams)
		if err != nil {
			return fmt.Errorf("error setting Cognito User's password (%s): %w", d.Id(), err)
		}
	}

	return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	log.Println("[DEBUG] Reading Cognito User")

	params := &cognitoidentityprovider.AdminGetUserInput{
		Username:   aws.String(d.Get("username").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	user, err := conn.AdminGetUser(params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeUserNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResUser, d.Get("username").(string))
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResUser, d.Get("username").(string), err)
	}

	if err := d.Set("attributes", flattenUserAttributes(user.UserAttributes)); err != nil {
		return fmt.Errorf("failed setting user attributes (%s): %w", d.Id(), err)
	}

	if err := d.Set("mfa_setting_list", user.UserMFASettingList); err != nil {
		return fmt.Errorf("failed setting user's mfa settings (%s): %w", d.Id(), err)
	}

	d.Set("preferred_mfa_setting", user.PreferredMfaSetting)
	d.Set("status", user.UserStatus)
	d.Set("enabled", user.Enabled)
	d.Set("creation_date", user.UserCreateDate.Format(time.RFC3339))
	d.Set("last_modified_date", user.UserLastModifiedDate.Format(time.RFC3339))
	d.Set("sub", retrieveUserSub(user.UserAttributes))

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	log.Println("[DEBUG] Updating Cognito User")

	if d.HasChange("attributes") {
		old, new := d.GetChange("attributes")

		upd, del := computeUserAttributesUpdate(old, new)

		if len(upd) > 0 {
			params := &cognitoidentityprovider.AdminUpdateUserAttributesInput{
				Username:       aws.String(d.Get("username").(string)),
				UserPoolId:     aws.String(d.Get("user_pool_id").(string)),
				UserAttributes: expandAttribute(upd),
			}

			if v, ok := d.GetOk("client_metadata"); ok {
				metadata := v.(map[string]interface{})
				params.ClientMetadata = expandUserClientMetadata(metadata)
			}

			_, err := conn.AdminUpdateUserAttributes(params)
			if err != nil {
				return fmt.Errorf("error updating Cognito User Attributes (%s): %w", d.Id(), err)
			}
		}
		if len(del) > 0 {
			params := &cognitoidentityprovider.AdminDeleteUserAttributesInput{
				Username:           aws.String(d.Get("username").(string)),
				UserPoolId:         aws.String(d.Get("user_pool_id").(string)),
				UserAttributeNames: expandUserAttributesDelete(del),
			}
			_, err := conn.AdminDeleteUserAttributes(params)
			if err != nil {
				return fmt.Errorf("error updating Cognito User Attributes (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)

		if enabled {
			enableParams := &cognitoidentityprovider.AdminEnableUserInput{
				Username:   aws.String(d.Get("username").(string)),
				UserPoolId: aws.String(d.Get("user_pool_id").(string)),
			}
			_, err := conn.AdminEnableUser(enableParams)
			if err != nil {
				return fmt.Errorf("error enabling Cognito User (%s): %w", d.Id(), err)
			}
		} else {
			disableParams := &cognitoidentityprovider.AdminDisableUserInput{
				Username:   aws.String(d.Get("username").(string)),
				UserPoolId: aws.String(d.Get("user_pool_id").(string)),
			}
			_, err := conn.AdminDisableUser(disableParams)
			if err != nil {
				return fmt.Errorf("error disabling Cognito User (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("temporary_password") {
		password := d.Get("temporary_password").(string)

		if password != "" {
			setPasswordParams := &cognitoidentityprovider.AdminSetUserPasswordInput{
				Username:   aws.String(d.Get("username").(string)),
				UserPoolId: aws.String(d.Get("user_pool_id").(string)),
				Password:   aws.String(password),
				Permanent:  aws.Bool(false),
			}

			_, err := conn.AdminSetUserPassword(setPasswordParams)
			if err != nil {
				return fmt.Errorf("error changing Cognito User's temporary password (%s): %w", d.Id(), err)
			}
		} else {
			d.Set("temporary_password", nil)
		}
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)

		if password != "" {
			setPasswordParams := &cognitoidentityprovider.AdminSetUserPasswordInput{
				Username:   aws.String(d.Get("username").(string)),
				UserPoolId: aws.String(d.Get("user_pool_id").(string)),
				Password:   aws.String(password),
				Permanent:  aws.Bool(true),
			}

			_, err := conn.AdminSetUserPassword(setPasswordParams)
			if err != nil {
				return fmt.Errorf("error changing Cognito User's password (%s): %w", d.Id(), err)
			}
		} else {
			d.Set("password", nil)
		}
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	log.Print("[DEBUG] Deleting Cognito User")

	params := &cognitoidentityprovider.AdminDeleteUserInput{
		Username:   aws.String(d.Get("username").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	_, err := conn.AdminDeleteUser(params)
	if err != nil {
		return fmt.Errorf("error deleting Cognito User (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("error importing Cognito User. Must specify user_pool_id/username")
	}
	userPoolId := idSplit[0]
	name := idSplit[1]
	d.Set("user_pool_id", userPoolId)
	d.Set("username", name)
	return []*schema.ResourceData{d}, nil
}

func expandAttribute(tfMap map[string]interface{}) []*cognitoidentityprovider.AttributeType {
	if len(tfMap) == 0 {
		return nil
	}

	apiList := make([]*cognitoidentityprovider.AttributeType, 0, len(tfMap))

	for k, v := range tfMap {
		if !UserAttributeKeyMatchesStandardAttribute(k) && !strings.HasPrefix(k, "custom:") {
			k = fmt.Sprintf("custom:%v", k)
		}
		apiList = append(apiList, &cognitoidentityprovider.AttributeType{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return apiList
}

func expandUserAttributesDelete(input []*string) []*string {
	result := make([]*string, 0, len(input))

	for _, v := range input {
		if !UserAttributeKeyMatchesStandardAttribute(*v) && !strings.HasPrefix(*v, "custom:") {
			formattedV := fmt.Sprintf("custom:%v", *v)
			result = append(result, &formattedV)
		} else {
			result = append(result, v)
		}
	}

	return result
}

func flattenUserAttributes(apiList []*cognitoidentityprovider.AttributeType) map[string]interface{} {
	tfMap := make(map[string]interface{})

	for _, apiAttribute := range apiList {
		if apiAttribute.Name != nil {
			if UserAttributeKeyMatchesStandardAttribute(*apiAttribute.Name) {
				tfMap[aws.StringValue(apiAttribute.Name)] = aws.StringValue(apiAttribute.Value)
			} else {
				name := strings.TrimPrefix(strings.TrimPrefix(aws.StringValue(apiAttribute.Name), "dev:"), "custom:")
				tfMap[name] = aws.StringValue(apiAttribute.Value)
			}
		}
	}

	return tfMap
}

// computeUserAttributesUpdate computes which user attributes should be updated and which ones should be deleted.
// We should do it like this because we cannot set a list of user attributes in cognito.
// We can either perfor update or delete operation
func computeUserAttributesUpdate(old interface{}, new interface{}) (map[string]interface{}, []*string) {
	oldMap := old.(map[string]interface{})
	newMap := new.(map[string]interface{})

	upd := make(map[string]interface{})

	for k, v := range newMap {
		if oldV, ok := oldMap[k]; ok {
			if oldV.(string) != v.(string) {
				upd[k] = v
			}
			delete(oldMap, k)
		} else {
			upd[k] = v
		}
	}

	del := make([]*string, 0, len(oldMap))
	for k := range oldMap {
		del = append(del, aws.String(k))
	}

	return upd, del
}

func expandUserDesiredDeliveryMediums(tfSet *schema.Set) []*string {
	apiList := []*string{}

	for _, elem := range tfSet.List() {
		apiList = append(apiList, aws.String(elem.(string)))
	}

	return apiList
}

func retrieveUserSub(apiList []*cognitoidentityprovider.AttributeType) string {
	for _, attr := range apiList {
		if aws.StringValue(attr.Name) == "sub" {
			return aws.StringValue(attr.Value)
		}
	}

	return ""
}

// For ClientMetadata we only need expand since AWS doesn't store its value
func expandUserClientMetadata(tfMap map[string]interface{}) map[string]*string {
	apiMap := map[string]*string{}
	for k, v := range tfMap {
		apiMap[k] = aws.String(v.(string))
	}

	return apiMap
}

func UserAttributeKeyMatchesStandardAttribute(input string) bool {
	if len(input) == 0 {
		return false
	}

	var standardAttributeKeys = []string{
		"address",
		"birthdate",
		"email",
		"email_verified",
		"gender",
		"given_name",
		"family_name",
		"locale",
		"middle_name",
		"name",
		"nickname",
		"phone_number",
		"phone_number_verified",
		"picture",
		"preferred_username",
		"profile",
		"sub",
		"updated_at",
		"website",
		"zoneinfo",
	}

	for _, attribute := range standardAttributeKeys {
		if input == attribute {
			return true
		}
	}
	return false
}
