package cognitoidp

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceUserPoolClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceUserPoolClientCreate,
		Read:   resourceUserPoolClientRead,
		Update: resourceUserPoolClientUpdate,
		Delete: resourceUserPoolClientDelete,

		Importer: &schema.ResourceImporter{
			State: resourceUserPoolClientImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPoolClient.html
		Schema: map[string]*schema.Schema{
			"access_token_validity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 86400),
			},
			"allowed_oauth_flows": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.OAuthFlowType_Values(), false),
				},
			},
			"allowed_oauth_flows_user_pool_client": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allowed_oauth_scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					// https://docs.aws.amazon.com/cognito/latest/developerguide/authorization-endpoint.html
					// System reserved scopes are openid, email, phone, profile, and aws.cognito.signin.user.admin.
					// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_CreateUserPoolClient.html#CognitoUserPools-CreateUserPoolClient-request-AllowedOAuthScopes
					// Constraints seem like to be designed for custom scopes which are not supported yet?
				},
			},
			"analytics_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"application_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ExactlyOneOf: []string{"analytics_configuration.0.application_id", "analytics_configuration.0.application_arn"},
						},
						"application_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ExactlyOneOf:  []string{"analytics_configuration.0.application_id", "analytics_configuration.0.application_arn"},
							ConflictsWith: []string{"analytics_configuration.0.external_id", "analytics_configuration.0.role_arn"},
							ValidateFunc:  verify.ValidARN,
						},
						"external_id": {
							Type:          schema.TypeString,
							ConflictsWith: []string{"analytics_configuration.0.application_arn"},
							Optional:      true,
						},
						"role_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ConflictsWith: []string{"analytics_configuration.0.application_arn"},
							ValidateFunc:  verify.ValidARN,
						},
						"user_data_shared": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"callback_urls": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"default_redirect_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1024),
					validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
						"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
				),
			},
			"enable_token_revocation": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"explicit_auth_flows": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cognitoidentityprovider.ExplicitAuthFlowsType_Values(), false),
				},
			},
			"generate_secret": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"id_token_validity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 86400),
			},
			"logout_urls": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 1024),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`[\w\s+=,.@-]+`),
						"must satisfy regular expression pattern: `[\\w\\s+=,.@-]+`"),
				),
			},
			"prevent_user_existence_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(cognitoidentityprovider.PreventUserExistenceErrorTypes_Values(), false),
			},
			"read_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"refresh_token_validity": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(0, 315360000),
			},
			"supported_identity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 32),
						validation.StringMatch(regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
							"must satisfy regular expression pattern: [\\p{L}\\p{M}\\p{S}\\p{N}\\p{P}]+`"),
					),
				},
			},
			"token_validity_units": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeHours,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
						"id_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeHours,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
						"refresh_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cognitoidentityprovider.TimeUnitsTypeDays,
							ValidateFunc: validation.StringInSlice(cognitoidentityprovider.TimeUnitsType_Values(), false),
						},
					},
				},
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"write_attributes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceUserPoolClientCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.CreateUserPoolClientInput{
		ClientName: aws.String(d.Get("name").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("generate_secret"); ok {
		params.GenerateSecret = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("explicit_auth_flows"); ok {
		params.ExplicitAuthFlows = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("refresh_token_validity"); ok {
		params.RefreshTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("access_token_validity"); ok {
		params.AccessTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("id_token_validity"); ok {
		params.IdTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allowed_oauth_flows"); ok {
		params.AllowedOAuthFlows = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("allowed_oauth_flows_user_pool_client"); ok {
		params.AllowedOAuthFlowsUserPoolClient = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("allowed_oauth_scopes"); ok {
		params.AllowedOAuthScopes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("callback_urls"); ok {
		params.CallbackURLs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_redirect_uri"); ok {
		params.DefaultRedirectURI = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logout_urls"); ok {
		params.LogoutURLs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("supported_identity_providers"); ok {
		params.SupportedIdentityProviders = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("analytics_configuration"); ok {
		params.AnalyticsConfiguration = expandUserPoolClientAnalyticsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("token_validity_units"); ok {
		params.TokenValidityUnits = expandUserPoolClientTokenValidityUnitsType(v.([]interface{}))
	}

	if v, ok := d.GetOk("prevent_user_existence_errors"); ok {
		params.PreventUserExistenceErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enable_token_revocation"); ok {
		params.EnableTokenRevocation = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Client: %s", params)

	resp, err := conn.CreateUserPoolClient(params)

	if err != nil {
		return fmt.Errorf("error creating Cognito User Pool Client (%s): %w", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(resp.UserPoolClient.ClientId))

	return resourceUserPoolClientRead(d, meta)
}

func resourceUserPoolClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Reading Cognito User Pool Client: %s", params)

	resp, err := conn.DescribeUserPoolClient(params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIDP, names.ErrActionReading, ResUserPoolClient, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIDP, names.ErrActionReading, ResUserPoolClient, d.Id(), err)
	}

	userPoolClient := resp.UserPoolClient
	d.Set("user_pool_id", userPoolClient.UserPoolId)
	d.Set("name", userPoolClient.ClientName)
	d.Set("explicit_auth_flows", flex.FlattenStringSet(userPoolClient.ExplicitAuthFlows))
	d.Set("read_attributes", flex.FlattenStringSet(userPoolClient.ReadAttributes))
	d.Set("write_attributes", flex.FlattenStringSet(userPoolClient.WriteAttributes))
	d.Set("refresh_token_validity", userPoolClient.RefreshTokenValidity)
	d.Set("access_token_validity", userPoolClient.AccessTokenValidity)
	d.Set("id_token_validity", userPoolClient.IdTokenValidity)
	d.Set("client_secret", userPoolClient.ClientSecret)
	d.Set("allowed_oauth_flows", flex.FlattenStringSet(userPoolClient.AllowedOAuthFlows))
	d.Set("allowed_oauth_flows_user_pool_client", userPoolClient.AllowedOAuthFlowsUserPoolClient)
	d.Set("allowed_oauth_scopes", flex.FlattenStringSet(userPoolClient.AllowedOAuthScopes))
	d.Set("callback_urls", flex.FlattenStringSet(userPoolClient.CallbackURLs))
	d.Set("default_redirect_uri", userPoolClient.DefaultRedirectURI)
	d.Set("logout_urls", flex.FlattenStringSet(userPoolClient.LogoutURLs))
	d.Set("prevent_user_existence_errors", userPoolClient.PreventUserExistenceErrors)
	d.Set("supported_identity_providers", flex.FlattenStringSet(userPoolClient.SupportedIdentityProviders))
	d.Set("enable_token_revocation", userPoolClient.EnableTokenRevocation)

	if err := d.Set("analytics_configuration", flattenUserPoolClientAnalyticsConfig(userPoolClient.AnalyticsConfiguration)); err != nil {
		return fmt.Errorf("error setting analytics_configuration: %w", err)
	}

	if err := d.Set("token_validity_units", flattenUserPoolClientTokenValidityUnitsType(userPoolClient.TokenValidityUnits)); err != nil {
		return fmt.Errorf("error setting token_validity_units: %w", err)
	}

	return nil
}

func resourceUserPoolClientUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.UpdateUserPoolClientInput{
		ClientId:              aws.String(d.Id()),
		UserPoolId:            aws.String(d.Get("user_pool_id").(string)),
		EnableTokenRevocation: aws.Bool(d.Get("enable_token_revocation").(bool)),
	}

	if v, ok := d.GetOk("name"); ok {
		params.ClientName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("explicit_auth_flows"); ok {
		params.ExplicitAuthFlows = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("read_attributes"); ok {
		params.ReadAttributes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("write_attributes"); ok {
		params.WriteAttributes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("refresh_token_validity"); ok {
		params.RefreshTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("access_token_validity"); ok {
		params.AccessTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("id_token_validity"); ok {
		params.IdTokenValidity = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("allowed_oauth_flows"); ok {
		params.AllowedOAuthFlows = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("allowed_oauth_flows_user_pool_client"); ok {
		params.AllowedOAuthFlowsUserPoolClient = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("allowed_oauth_scopes"); ok {
		params.AllowedOAuthScopes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("callback_urls"); ok {
		params.CallbackURLs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_redirect_uri"); ok {
		params.DefaultRedirectURI = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logout_urls"); ok {
		params.LogoutURLs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("prevent_user_existence_errors"); ok {
		params.PreventUserExistenceErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_identity_providers"); ok {
		params.SupportedIdentityProviders = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("analytics_configuration"); ok {
		params.AnalyticsConfiguration = expandUserPoolClientAnalyticsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("token_validity_units"); ok {
		params.TokenValidityUnits = expandUserPoolClientTokenValidityUnitsType(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating Cognito User Pool Client: %s", params)

	_, err := verify.RetryOnAWSCode(cognitoidentityprovider.ErrCodeConcurrentModificationException, func() (interface{}, error) {
		return conn.UpdateUserPoolClient(params)
	})
	if err != nil {
		return fmt.Errorf("error updating Cognito User Pool Client (%s): %w", d.Id(), err)
	}

	return resourceUserPoolClientRead(d, meta)
}

func resourceUserPoolClientDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIDPConn

	params := &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Cognito User Pool Client: %s", params)

	_, err := conn.DeleteUserPoolClient(params)

	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Cognito User Pool Client (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceUserPoolClientImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 || len(d.Id()) < 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of resource: %s. Please follow 'user-pool-id/client-id'", d.Id())
	}
	userPoolId := strings.Split(d.Id(), "/")[0]
	clientId := strings.Split(d.Id(), "/")[1]
	d.SetId(clientId)
	d.Set("user_pool_id", userPoolId)
	log.Printf("[DEBUG] Importing client %s for user pool %s", clientId, userPoolId)

	return []*schema.ResourceData{d}, nil
}

func expandUserPoolClientAnalyticsConfig(l []interface{}) *cognitoidentityprovider.AnalyticsConfigurationType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	analyticsConfig := &cognitoidentityprovider.AnalyticsConfigurationType{}

	if v, ok := m["role_arn"]; ok && v != "" {
		analyticsConfig.RoleArn = aws.String(v.(string))
	}

	if v, ok := m["external_id"]; ok && v != "" {
		analyticsConfig.ExternalId = aws.String(v.(string))
	}

	if v, ok := m["application_id"]; ok && v != "" {
		analyticsConfig.ApplicationId = aws.String(v.(string))
	}

	if v, ok := m["application_arn"]; ok && v != "" {
		analyticsConfig.ApplicationArn = aws.String(v.(string))
	}

	if v, ok := m["user_data_shared"]; ok {
		analyticsConfig.UserDataShared = aws.Bool(v.(bool))
	}

	return analyticsConfig
}

func flattenUserPoolClientAnalyticsConfig(analyticsConfig *cognitoidentityprovider.AnalyticsConfigurationType) []interface{} {
	if analyticsConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"user_data_shared": aws.BoolValue(analyticsConfig.UserDataShared),
	}

	if analyticsConfig.ExternalId != nil {
		m["external_id"] = aws.StringValue(analyticsConfig.ExternalId)
	}

	if analyticsConfig.RoleArn != nil {
		m["role_arn"] = aws.StringValue(analyticsConfig.RoleArn)
	}

	if analyticsConfig.ApplicationId != nil {
		m["application_id"] = aws.StringValue(analyticsConfig.ApplicationId)
	}

	if analyticsConfig.ApplicationArn != nil {
		m["application_arn"] = aws.StringValue(analyticsConfig.ApplicationArn)
	}

	return []interface{}{m}
}

func expandUserPoolClientTokenValidityUnitsType(l []interface{}) *cognitoidentityprovider.TokenValidityUnitsType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	tokenValidityConfig := &cognitoidentityprovider.TokenValidityUnitsType{}

	if v, ok := m["access_token"]; ok {
		tokenValidityConfig.AccessToken = aws.String(v.(string))
	}

	if v, ok := m["id_token"]; ok {
		tokenValidityConfig.IdToken = aws.String(v.(string))
	}

	if v, ok := m["refresh_token"]; ok {
		tokenValidityConfig.RefreshToken = aws.String(v.(string))
	}

	return tokenValidityConfig
}

func flattenUserPoolClientTokenValidityUnitsType(tokenValidityConfig *cognitoidentityprovider.TokenValidityUnitsType) []interface{} {
	if tokenValidityConfig == nil {
		return nil
	}

	//tokenValidityConfig is never nil and if everything is empty it causes diffs
	if tokenValidityConfig.IdToken == nil && tokenValidityConfig.AccessToken == nil && tokenValidityConfig.RefreshToken == nil {
		return nil
	}

	m := map[string]interface{}{}

	if tokenValidityConfig.IdToken != nil {
		m["id_token"] = aws.StringValue(tokenValidityConfig.IdToken)
	}

	if tokenValidityConfig.AccessToken != nil {
		m["access_token"] = aws.StringValue(tokenValidityConfig.AccessToken)
	}

	if tokenValidityConfig.RefreshToken != nil {
		m["refresh_token"] = aws.StringValue(tokenValidityConfig.RefreshToken)
	}

	return []interface{}{m}
}
