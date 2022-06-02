package appsync

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var validateAuthorizerResultTTLInSeconds = validation.IntBetween(0, 3600)

const DefaultAuthorizerResultTTLInSeconds = 300

func ResourceGraphQLAPI() *schema.Resource {
	return &schema.Resource{
		Create: resourceGraphQLAPICreate,
		Read:   resourceGraphQLAPIRead,
		Update: resourceGraphQLAPIUpdate,
		Delete: resourceGraphQLAPIDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"additional_authentication_provider": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appsync.AuthenticationType_Values(), false),
						},
						"openid_connect_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auth_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"client_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"iat_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"issuer": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"user_pool_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_id_client_regex": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"aws_region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"user_pool_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"lambda_authorizer_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorizer_result_ttl_in_seconds": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      DefaultAuthorizerResultTTLInSeconds,
										ValidateFunc: validateAuthorizerResultTTLInSeconds,
									},
									"authorizer_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"identity_validation_expression": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"authentication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.AuthenticationType_Values(), false),
			},
			"schema": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"log_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field_log_level": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								appsync.FieldLogLevelAll,
								appsync.FieldLogLevelError,
								appsync.FieldLogLevelNone,
							}, false),
						},
						"exclude_verbose_content": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"openid_connect_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"client_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iat_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"issuer": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"user_pool_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id_client_regex": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"aws_region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"default_action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								appsync.DefaultActionAllow,
								appsync.DefaultActionDeny,
							}, false),
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"lambda_authorizer_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorizer_result_ttl_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      DefaultAuthorizerResultTTLInSeconds,
							ValidateFunc: validateAuthorizerResultTTLInSeconds,
						},
						"authorizer_uri": {
							Type:     schema.TypeString,
							Required: true,
						},
						"identity_validation_expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"uris": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"xray_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGraphQLAPICreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &appsync.CreateGraphqlApiInput{
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("log_config"); ok {
		input.LogConfig = expandGraphQLAPILogConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_config"); ok {
		input.OpenIDConnectConfig = expandGraphQLAPIOpenIDConnectConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_pool_config"); ok {
		input.UserPoolConfig = expandGraphQLAPIUserPoolConfig(v.([]interface{}), meta.(*conns.AWSClient).Region)
	}

	if v, ok := d.GetOk("lambda_authorizer_config"); ok {
		input.LambdaAuthorizerConfig = expandGraphQLAPILambdaAuthorizerConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandGraphQLAPIAdditionalAuthProviders(v.([]interface{}), meta.(*conns.AWSClient).Region)
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateGraphqlApi(input)
	if err != nil {
		return fmt.Errorf("error creating AppSync GraphQL API: %s", err)
	}

	d.SetId(aws.StringValue(resp.GraphqlApi.ApiId))

	if err := resourceSchemaPut(d, meta); err != nil {
		return fmt.Errorf("error creating AppSync GraphQL API (%s) Schema: %s", d.Id(), err)
	}

	return resourceGraphQLAPIRead(d, meta)
}

func resourceGraphQLAPIRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &appsync.GetGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}

	resp, err := conn.GetGraphqlApi(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync GraphQL API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.GraphqlApi.Arn)
	d.Set("authentication_type", resp.GraphqlApi.AuthenticationType)
	d.Set("name", resp.GraphqlApi.Name)

	if err := d.Set("log_config", flattenGraphQLAPILogConfig(resp.GraphqlApi.LogConfig)); err != nil {
		return fmt.Errorf("error setting log_config: %s", err)
	}

	if err := d.Set("openid_connect_config", flattenGraphQLAPIOpenIDConnectConfig(resp.GraphqlApi.OpenIDConnectConfig)); err != nil {
		return fmt.Errorf("error setting openid_connect_config: %s", err)
	}

	if err := d.Set("user_pool_config", flattenGraphQLAPIUserPoolConfig(resp.GraphqlApi.UserPoolConfig)); err != nil {
		return fmt.Errorf("error setting user_pool_config: %s", err)
	}

	if err := d.Set("lambda_authorizer_config", flattenGraphQLAPILambdaAuthorizerConfig(resp.GraphqlApi.LambdaAuthorizerConfig)); err != nil {
		return fmt.Errorf("error setting lambda_authorizer_config: %s", err)
	}

	if err := d.Set("additional_authentication_provider", flattenGraphQLAPIAdditionalAuthenticationProviders(resp.GraphqlApi.AdditionalAuthenticationProviders)); err != nil {
		return fmt.Errorf("error setting additional_authentication_provider: %s", err)
	}

	if err := d.Set("uris", aws.StringValueMap(resp.GraphqlApi.Uris)); err != nil {
		return fmt.Errorf("error setting uris: %s", err)
	}

	tags := KeyValueTags(resp.GraphqlApi.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("xray_enabled", resp.GraphqlApi.XrayEnabled); err != nil {
		return fmt.Errorf("error setting xray_enabled: %s", err)
	}

	return nil
}

func resourceGraphQLAPIUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppSync GraphQL API (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	input := &appsync.UpdateGraphqlApiInput{
		ApiId:              aws.String(d.Id()),
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("log_config"); ok {
		input.LogConfig = expandGraphQLAPILogConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_config"); ok {
		input.OpenIDConnectConfig = expandGraphQLAPIOpenIDConnectConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_pool_config"); ok {
		input.UserPoolConfig = expandGraphQLAPIUserPoolConfig(v.([]interface{}), meta.(*conns.AWSClient).Region)
	}

	if v, ok := d.GetOk("lambda_authorizer_config"); ok {
		input.LambdaAuthorizerConfig = expandGraphQLAPILambdaAuthorizerConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandGraphQLAPIAdditionalAuthProviders(v.([]interface{}), meta.(*conns.AWSClient).Region)
	}

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = aws.Bool(v.(bool))
	}

	_, err := conn.UpdateGraphqlApi(input)
	if err != nil {
		return fmt.Errorf("error updating AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	if d.HasChange("schema") {
		if err := resourceSchemaPut(d, meta); err != nil {
			return fmt.Errorf("error updating AppSync GraphQL API (%s) Schema: %s", d.Id(), err)
		}
	}

	return resourceGraphQLAPIRead(d, meta)
}

func resourceGraphQLAPIDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.DeleteGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGraphqlApi(input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	return nil
}

func expandGraphQLAPILogConfig(l []interface{}) *appsync.LogConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	logConfig := &appsync.LogConfig{
		CloudWatchLogsRoleArn: aws.String(m["cloudwatch_logs_role_arn"].(string)),
		FieldLogLevel:         aws.String(m["field_log_level"].(string)),
		ExcludeVerboseContent: aws.Bool(m["exclude_verbose_content"].(bool)),
	}

	return logConfig
}

func expandGraphQLAPIOpenIDConnectConfig(l []interface{}) *appsync.OpenIDConnectConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	openIDConnectConfig := &appsync.OpenIDConnectConfig{
		Issuer: aws.String(m["issuer"].(string)),
	}

	if v, ok := m["auth_ttl"].(int); ok && v != 0 {
		openIDConnectConfig.AuthTTL = aws.Int64(int64(v))
	}

	if v, ok := m["client_id"].(string); ok && v != "" {
		openIDConnectConfig.ClientId = aws.String(v)
	}

	if v, ok := m["iat_ttl"].(int); ok && v != 0 {
		openIDConnectConfig.IatTTL = aws.Int64(int64(v))
	}

	return openIDConnectConfig
}

func expandGraphQLAPIUserPoolConfig(l []interface{}, currentRegion string) *appsync.UserPoolConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	userPoolConfig := &appsync.UserPoolConfig{
		AwsRegion:     aws.String(currentRegion),
		DefaultAction: aws.String(m["default_action"].(string)),
		UserPoolId:    aws.String(m["user_pool_id"].(string)),
	}

	if v, ok := m["app_id_client_regex"].(string); ok && v != "" {
		userPoolConfig.AppIdClientRegex = aws.String(v)
	}

	if v, ok := m["aws_region"].(string); ok && v != "" {
		userPoolConfig.AwsRegion = aws.String(v)
	}

	return userPoolConfig
}

func expandGraphQLAPILambdaAuthorizerConfig(l []interface{}) *appsync.LambdaAuthorizerConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	lambdaAuthorizerConfig := &appsync.LambdaAuthorizerConfig{
		AuthorizerResultTtlInSeconds: aws.Int64(int64(m["authorizer_result_ttl_in_seconds"].(int))),
		AuthorizerUri:                aws.String(m["authorizer_uri"].(string)),
	}

	if v, ok := m["identity_validation_expression"].(string); ok && v != "" {
		lambdaAuthorizerConfig.IdentityValidationExpression = aws.String(v)
	}

	return lambdaAuthorizerConfig
}

func expandGraphQLAPIAdditionalAuthProviders(items []interface{}, currentRegion string) []*appsync.AdditionalAuthenticationProvider {
	if len(items) < 1 {
		return nil
	}

	additionalAuthProviders := make([]*appsync.AdditionalAuthenticationProvider, 0, len(items))
	for _, l := range items {
		if l == nil {
			continue
		}

		m := l.(map[string]interface{})
		additionalAuthProvider := &appsync.AdditionalAuthenticationProvider{
			AuthenticationType: aws.String(m["authentication_type"].(string)),
		}

		if v, ok := m["openid_connect_config"]; ok {
			additionalAuthProvider.OpenIDConnectConfig = expandGraphQLAPIOpenIDConnectConfig(v.([]interface{}))
		}

		if v, ok := m["user_pool_config"]; ok {
			additionalAuthProvider.UserPoolConfig = expandGraphQLAPICognitoUserPoolConfig(v.([]interface{}), currentRegion)
		}

		if v, ok := m["lambda_authorizer_config"]; ok {
			additionalAuthProvider.LambdaAuthorizerConfig = expandGraphQLAPILambdaAuthorizerConfig(v.([]interface{}))
		}

		additionalAuthProviders = append(additionalAuthProviders, additionalAuthProvider)
	}

	return additionalAuthProviders
}

func expandGraphQLAPICognitoUserPoolConfig(l []interface{}, currentRegion string) *appsync.CognitoUserPoolConfig {
	if len(l) < 1 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	userPoolConfig := &appsync.CognitoUserPoolConfig{
		AwsRegion:  aws.String(currentRegion),
		UserPoolId: aws.String(m["user_pool_id"].(string)),
	}

	if v, ok := m["app_id_client_regex"].(string); ok && v != "" {
		userPoolConfig.AppIdClientRegex = aws.String(v)
	}

	if v, ok := m["aws_region"].(string); ok && v != "" {
		userPoolConfig.AwsRegion = aws.String(v)
	}

	return userPoolConfig
}

func flattenGraphQLAPILogConfig(logConfig *appsync.LogConfig) []interface{} {
	if logConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloudwatch_logs_role_arn": aws.StringValue(logConfig.CloudWatchLogsRoleArn),
		"field_log_level":          aws.StringValue(logConfig.FieldLogLevel),
		"exclude_verbose_content":  aws.BoolValue(logConfig.ExcludeVerboseContent),
	}

	return []interface{}{m}
}

func flattenGraphQLAPIOpenIDConnectConfig(openIDConnectConfig *appsync.OpenIDConnectConfig) []interface{} {
	if openIDConnectConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"auth_ttl":  aws.Int64Value(openIDConnectConfig.AuthTTL),
		"client_id": aws.StringValue(openIDConnectConfig.ClientId),
		"iat_ttl":   aws.Int64Value(openIDConnectConfig.IatTTL),
		"issuer":    aws.StringValue(openIDConnectConfig.Issuer),
	}

	return []interface{}{m}
}

func flattenGraphQLAPIUserPoolConfig(userPoolConfig *appsync.UserPoolConfig) []interface{} {
	if userPoolConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"aws_region":     aws.StringValue(userPoolConfig.AwsRegion),
		"default_action": aws.StringValue(userPoolConfig.DefaultAction),
		"user_pool_id":   aws.StringValue(userPoolConfig.UserPoolId),
	}

	if userPoolConfig.AppIdClientRegex != nil {
		m["app_id_client_regex"] = aws.StringValue(userPoolConfig.AppIdClientRegex)
	}

	return []interface{}{m}
}

func flattenGraphQLAPILambdaAuthorizerConfig(lambdaAuthorizerConfig *appsync.LambdaAuthorizerConfig) []interface{} {
	if lambdaAuthorizerConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authorizer_uri": aws.StringValue(lambdaAuthorizerConfig.AuthorizerUri),
	}

	if lambdaAuthorizerConfig.AuthorizerResultTtlInSeconds != nil {
		m["authorizer_result_ttl_in_seconds"] = aws.Int64Value(lambdaAuthorizerConfig.AuthorizerResultTtlInSeconds)
	} else {
		m["authorizer_result_ttl_in_seconds"] = DefaultAuthorizerResultTTLInSeconds
	}

	if lambdaAuthorizerConfig.IdentityValidationExpression != nil {
		m["identity_validation_expression"] = aws.StringValue(lambdaAuthorizerConfig.IdentityValidationExpression)
	}

	return []interface{}{m}
}

func flattenGraphQLAPIAdditionalAuthenticationProviders(additionalAuthenticationProviders []*appsync.AdditionalAuthenticationProvider) []interface{} {
	if len(additionalAuthenticationProviders) == 0 {
		return []interface{}{}
	}

	result := make([]interface{}, len(additionalAuthenticationProviders))
	for i, provider := range additionalAuthenticationProviders {
		result[i] = map[string]interface{}{
			"authentication_type":      aws.StringValue(provider.AuthenticationType),
			"lambda_authorizer_config": flattenGraphQLAPILambdaAuthorizerConfig(provider.LambdaAuthorizerConfig),
			"openid_connect_config":    flattenGraphQLAPIOpenIDConnectConfig(provider.OpenIDConnectConfig),
			"user_pool_config":         flattenGraphQLAPICognitoUserPoolConfig(provider.UserPoolConfig),
		}
	}

	return result
}

func flattenGraphQLAPICognitoUserPoolConfig(userPoolConfig *appsync.CognitoUserPoolConfig) []interface{} {
	if userPoolConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"aws_region":   aws.StringValue(userPoolConfig.AwsRegion),
		"user_pool_id": aws.StringValue(userPoolConfig.UserPoolId),
	}

	if userPoolConfig.AppIdClientRegex != nil {
		m["app_id_client_regex"] = aws.StringValue(userPoolConfig.AppIdClientRegex)
	}

	return []interface{}{m}
}

func resourceSchemaPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	if v, ok := d.GetOk("schema"); ok {
		input := &appsync.StartSchemaCreationInput{
			ApiId:      aws.String(d.Id()),
			Definition: ([]byte)(v.(string)),
		}
		if _, err := conn.StartSchemaCreation(input); err != nil {
			return err
		}

		activeSchemaConfig := &resource.StateChangeConf{
			Pending: []string{appsync.SchemaStatusProcessing},
			Target:  []string{"SUCCESS", appsync.SchemaStatusActive}, // should be only appsync.SchemaStatusActive . I think this is a problem in documentation: https://docs.aws.amazon.com/appsync/latest/APIReference/API_GetSchemaCreationStatus.html
			Refresh: func() (interface{}, string, error) {
				result, err := conn.GetSchemaCreationStatus(&appsync.GetSchemaCreationStatusInput{
					ApiId: aws.String(d.Id()),
				})
				if err != nil {
					return 0, "", err
				}
				return result, *result.Status, nil
			},
			Timeout: d.Timeout(schema.TimeoutCreate),
		}

		if _, err := activeSchemaConfig.WaitForState(); err != nil {
			return fmt.Errorf("Error waiting for schema creation status on AppSync API %s: %s", d.Id(), err)
		}
	}

	return nil
}
