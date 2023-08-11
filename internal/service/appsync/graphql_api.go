// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var validateAuthorizerResultTTLInSeconds = validation.IntBetween(0, 3600)

const DefaultAuthorizerResultTTLInSeconds = 300

// @SDKResource("aws_appsync_graphql_api", name="GraphQL API")
// @Tags(identifierAttribute="arn")
func ResourceGraphQLAPI() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGraphQLAPICreate,
		ReadWithoutTimeout:   resourceGraphQLAPIRead,
		UpdateWithoutTimeout: resourceGraphQLAPIUpdate,
		DeleteWithoutTimeout: resourceGraphQLAPIDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.AuthenticationType_Values(), false),
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
			"log_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"exclude_verbose_content": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"field_log_level": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appsync.FieldLogLevel_Values(), false),
						},
					},
				},
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
			"schema": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uris": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appsync.DefaultAction_Values(), false),
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      appsync.GraphQLApiVisibilityGlobal,
				ValidateFunc: validation.StringInSlice(appsync.GraphQLApiVisibility_Values(), false),
			},
			"xray_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGraphQLAPICreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	name := d.Get("name").(string)
	input := &appsync.CreateGraphqlApiInput{
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		Name:               aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandGraphQLAPIAdditionalAuthProviders(v.([]interface{}), meta.(*conns.AWSClient).Region)
	}

	if v, ok := d.GetOk("lambda_authorizer_config"); ok {
		input.LambdaAuthorizerConfig = expandGraphQLAPILambdaAuthorizerConfig(v.([]interface{}))
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

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("visibility"); ok {
		input.Visibility = aws.String(v.(string))
	}

	output, err := conn.CreateGraphqlApiWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync GraphQL API (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.GraphqlApi.ApiId))

	if v, ok := d.GetOk("schema"); ok {
		if err := putSchema(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceGraphQLAPIRead(ctx, d, meta)...)
}

func resourceGraphQLAPIRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	api, err := FindGraphQLAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync GraphQL API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_authentication_provider", flattenGraphQLAPIAdditionalAuthenticationProviders(api.AdditionalAuthenticationProviders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting additional_authentication_provider: %s", err)
	}
	d.Set("arn", api.Arn)
	d.Set("authentication_type", api.AuthenticationType)
	if err := d.Set("lambda_authorizer_config", flattenGraphQLAPILambdaAuthorizerConfig(api.LambdaAuthorizerConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_authorizer_config: %s", err)
	}
	if err := d.Set("log_config", flattenGraphQLAPILogConfig(api.LogConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_config: %s", err)
	}
	if err := d.Set("openid_connect_config", flattenGraphQLAPIOpenIDConnectConfig(api.OpenIDConnectConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting openid_connect_config: %s", err)
	}
	d.Set("name", api.Name)
	d.Set("uris", aws.StringValueMap(api.Uris))
	if err := d.Set("user_pool_config", flattenGraphQLAPIUserPoolConfig(api.UserPoolConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_pool_config: %s", err)
	}
	d.Set("visibility", api.Visibility)
	if err := d.Set("xray_enabled", api.XrayEnabled); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting xray_enabled: %s", err)
	}

	setTagsOut(ctx, api.Tags)

	return diags
}

func resourceGraphQLAPIUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &appsync.UpdateGraphqlApiInput{
			ApiId:              aws.String(d.Id()),
			AuthenticationType: aws.String(d.Get("authentication_type").(string)),
			Name:               aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("additional_authentication_provider"); ok {
			input.AdditionalAuthenticationProviders = expandGraphQLAPIAdditionalAuthProviders(v.([]interface{}), meta.(*conns.AWSClient).Region)
		}

		if v, ok := d.GetOk("lambda_authorizer_config"); ok {
			input.LambdaAuthorizerConfig = expandGraphQLAPILambdaAuthorizerConfig(v.([]interface{}))
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

		if v, ok := d.GetOk("xray_enabled"); ok {
			input.XrayEnabled = aws.Bool(v.(bool))
		}

		_, err := conn.UpdateGraphqlApiWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppSync GraphQL API (%s): %s", d.Id(), err)
		}

		if d.HasChange("schema") {
			if v, ok := d.GetOk("schema"); ok {
				if err := putSchema(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutCreate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceGraphQLAPIRead(ctx, d, meta)...)
}

func resourceGraphQLAPIDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	log.Printf("[DEBUG] Deleting AppSync GraphQL API: %s", d.Id())
	_, err := conn.DeleteGraphqlApiWithContext(ctx, &appsync.DeleteGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	return diags
}

func putSchema(ctx context.Context, conn *appsync.AppSync, apiID, definition string, timeout time.Duration) error {
	input := &appsync.StartSchemaCreationInput{
		ApiId:      aws.String(apiID),
		Definition: ([]byte)(definition),
	}

	_, err := conn.StartSchemaCreationWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("creating AppSync GraphQL API (%s) schema: %w", apiID, err)
	}

	if err := waitSchemaCreated(ctx, conn, apiID, timeout); err != nil {
		return fmt.Errorf("waiting for AppSync GraphQL API (%s) schema create: %w", apiID, err)
	}

	return nil
}

func FindGraphQLAPIByID(ctx context.Context, conn *appsync.AppSync, id string) (*appsync.GraphqlApi, error) {
	input := &appsync.GetGraphqlApiInput{
		ApiId: aws.String(id),
	}

	output, err := conn.GetGraphqlApiWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GraphqlApi == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GraphqlApi, nil
}

func findSchemaCreationStatusByID(ctx context.Context, conn *appsync.AppSync, id string) (*appsync.GetSchemaCreationStatusOutput, error) {
	input := &appsync.GetSchemaCreationStatusInput{
		ApiId: aws.String(id),
	}

	output, err := conn.GetSchemaCreationStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
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

func statusSchemaCreation(ctx context.Context, conn *appsync.AppSync, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSchemaCreationStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitSchemaCreated(ctx context.Context, conn *appsync.AppSync, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{appsync.SchemaStatusProcessing},
		Target:  []string{appsync.SchemaStatusActive, appsync.SchemaStatusSuccess},
		Refresh: statusSchemaCreation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appsync.GetSchemaCreationStatusOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Details)))
	}

	return err
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
