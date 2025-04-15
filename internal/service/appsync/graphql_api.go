// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultAuthorizerResultTTLInSeconds = 300
)

// @SDKResource("aws_appsync_graphql_api", name="GraphQL API")
// @Tags(identifierAttribute="arn")
func resourceGraphQLAPI() *schema.Resource {
	validateAuthorizerResultTTLInSeconds := validation.IntBetween(0, 3600)

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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
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
										Default:      defaultAuthorizerResultTTLInSeconds,
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
									names.AttrClientID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"iat_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrIssuer: {
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
									names.AttrUserPoolID: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"api_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.GraphQLApiType](),
				ForceNew:         true,
				Default:          awstypes.GraphQLApiTypeGraphql,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
			},
			"enhanced_metrics_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_source_level_metrics_behavior": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DataSourceLevelMetricsBehavior](),
						},
						"operation_level_metrics_config": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OperationLevelMetricsConfig](),
						},
						"resolver_level_metrics_behavior": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResolverLevelMetricsBehavior](),
						},
					},
				},
			},
			"introspection_config": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.GraphQLApiIntrospectionConfigEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.GraphQLApiIntrospectionConfig](),
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
							Default:      defaultAuthorizerResultTTLInSeconds,
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FieldLogLevel](),
						},
					},
				},
			},
			"merged_api_execution_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[A-Za-z_][0-9A-Za-z_]*`), ""),
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
						names.AttrClientID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iat_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrIssuer: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"query_depth_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 75),
			},
			"resolver_count_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 10000),
			},
			names.AttrSchema: {
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
						names.AttrDefaultAction: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DefaultAction](),
						},
						names.AttrUserPoolID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"visibility": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.GraphQLApiVisibilityGlobal,
				ValidateDiagFunc: enum.Validate[awstypes.GraphQLApiVisibility](),
			},
			"xray_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceGraphQLAPICreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appsync.CreateGraphqlApiInput{
		AuthenticationType: awstypes.AuthenticationType(d.Get("authentication_type").(string)),
		Name:               aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("additional_authentication_provider"); ok {
		input.AdditionalAuthenticationProviders = expandAdditionalAuthenticationProviders(v.([]any), meta.(*conns.AWSClient).Region(ctx))
	}

	if v, ok := d.GetOk("api_type"); ok {
		input.ApiType = awstypes.GraphQLApiType(v.(string))
	}

	if v, ok := d.GetOk("enhanced_metrics_config"); ok {
		input.EnhancedMetricsConfig = expandEnhancedMetricsConfig(v.([]any))
	}

	if v, ok := d.GetOk("introspection_config"); ok {
		input.IntrospectionConfig = awstypes.GraphQLApiIntrospectionConfig(v.(string))
	}

	if v, ok := d.GetOk("lambda_authorizer_config"); ok {
		input.LambdaAuthorizerConfig = expandLambdaAuthorizerConfig(v.([]any))
	}

	if v, ok := d.GetOk("log_config"); ok {
		input.LogConfig = expandLogConfig(v.([]any))
	}

	if v, ok := d.GetOk("merged_api_execution_role_arn"); ok {
		input.MergedApiExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("openid_connect_config"); ok {
		input.OpenIDConnectConfig = expandOpenIDConnectConfig(v.([]any))
	}

	if v, ok := d.GetOk("query_depth_limit"); ok {
		input.QueryDepthLimit = int32(v.(int))
	}

	if v, ok := d.GetOk("resolver_count_limit"); ok {
		input.ResolverCountLimit = int32(v.(int))
	}

	if v, ok := d.GetOk("user_pool_config"); ok {
		input.UserPoolConfig = expandUserPoolConfig(v.([]any), meta.(*conns.AWSClient).Region(ctx))
	}

	if v, ok := d.GetOk("xray_enabled"); ok {
		input.XrayEnabled = v.(bool)
	}

	if v, ok := d.GetOk("visibility"); ok {
		input.Visibility = awstypes.GraphQLApiVisibility(v.(string))
	}

	output, err := conn.CreateGraphqlApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppSync GraphQL API (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GraphqlApi.ApiId))

	if v, ok := d.GetOk(names.AttrSchema); ok {
		if err := putSchema(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceGraphQLAPIRead(ctx, d, meta)...)
}

func resourceGraphQLAPIRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	api, err := findGraphQLAPIByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync GraphQL API (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_authentication_provider", flattenAdditionalAuthenticationProviders(api.AdditionalAuthenticationProviders)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting additional_authentication_provider: %s", err)
	}
	d.Set("api_type", api.ApiType)
	d.Set(names.AttrARN, api.Arn)
	d.Set("authentication_type", api.AuthenticationType)
	if err := d.Set("enhanced_metrics_config", flattenEnhancedMetricsConfig(api.EnhancedMetricsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting enhanced_metrics_config: %s", err)
	}
	d.Set("introspection_config", api.IntrospectionConfig)
	if err := d.Set("lambda_authorizer_config", flattenLambdaAuthorizerConfig(api.LambdaAuthorizerConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_authorizer_config: %s", err)
	}
	if err := d.Set("log_config", flattenLogConfig(api.LogConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_config: %s", err)
	}
	d.Set("merged_api_execution_role_arn", api.MergedApiExecutionRoleArn)
	d.Set(names.AttrName, api.Name)
	if err := d.Set("openid_connect_config", flattenOpenIDConnectConfig(api.OpenIDConnectConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting openid_connect_config: %s", err)
	}
	d.Set("query_depth_limit", api.QueryDepthLimit)
	d.Set("resolver_count_limit", api.ResolverCountLimit)
	d.Set("uris", api.Uris)
	if err := d.Set("user_pool_config", flattenUserPoolConfig(api.UserPoolConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user_pool_config: %s", err)
	}
	d.Set("visibility", api.Visibility)
	d.Set("xray_enabled", api.XrayEnabled)

	setTagsOut(ctx, api.Tags)

	return diags
}

func resourceGraphQLAPIUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &appsync.UpdateGraphqlApiInput{
			ApiId:              aws.String(d.Id()),
			AuthenticationType: awstypes.AuthenticationType(d.Get("authentication_type").(string)),
			Name:               aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("additional_authentication_provider"); ok {
			input.AdditionalAuthenticationProviders = expandAdditionalAuthenticationProviders(v.([]any), meta.(*conns.AWSClient).Region(ctx))
		}

		if v, ok := d.GetOk("enhanced_metrics_config"); ok {
			input.EnhancedMetricsConfig = expandEnhancedMetricsConfig(v.([]any))
		}

		if v, ok := d.GetOk("introspection_config"); ok {
			input.IntrospectionConfig = awstypes.GraphQLApiIntrospectionConfig(v.(string))
		}

		if v, ok := d.GetOk("lambda_authorizer_config"); ok {
			input.LambdaAuthorizerConfig = expandLambdaAuthorizerConfig(v.([]any))
		}

		if v, ok := d.GetOk("log_config"); ok {
			input.LogConfig = expandLogConfig(v.([]any))
		}

		if v, ok := d.GetOk("merged_api_execution_role_arn"); ok {
			input.MergedApiExecutionRoleArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("openid_connect_config"); ok {
			input.OpenIDConnectConfig = expandOpenIDConnectConfig(v.([]any))
		}

		if v, ok := d.GetOk("query_depth_limit"); ok {
			input.QueryDepthLimit = int32(v.(int))
		}

		if v, ok := d.GetOk("resolver_count_limit"); ok {
			input.ResolverCountLimit = int32(v.(int))
		}

		if v, ok := d.GetOk("user_pool_config"); ok {
			input.UserPoolConfig = expandUserPoolConfig(v.([]any), meta.(*conns.AWSClient).Region(ctx))
		}

		if v, ok := d.GetOk("xray_enabled"); ok {
			input.XrayEnabled = v.(bool)
		}

		_, err := conn.UpdateGraphqlApi(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppSync GraphQL API (%s): %s", d.Id(), err)
		}

		if d.HasChange(names.AttrSchema) {
			if v, ok := d.GetOk(names.AttrSchema); ok {
				if err := putSchema(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceGraphQLAPIRead(ctx, d, meta)...)
}

func resourceGraphQLAPIDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	log.Printf("[DEBUG] Deleting AppSync GraphQL API: %s", d.Id())
	input := appsync.DeleteGraphqlApiInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteGraphqlApi(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppSync GraphQL API (%s): %s", d.Id(), err)
	}

	return diags
}

func putSchema(ctx context.Context, conn *appsync.Client, apiID, definition string, timeout time.Duration) error {
	input := &appsync.StartSchemaCreationInput{
		ApiId:      aws.String(apiID),
		Definition: ([]byte)(definition),
	}

	_, err := conn.StartSchemaCreation(ctx, input)

	if err != nil {
		return fmt.Errorf("creating AppSync GraphQL API (%s) schema: %w", apiID, err)
	}

	if _, err := waitSchemaCreated(ctx, conn, apiID, timeout); err != nil {
		return fmt.Errorf("waiting for AppSync GraphQL API (%s) schema create: %w", apiID, err)
	}

	return nil
}

func findGraphQLAPIByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.GraphqlApi, error) {
	input := &appsync.GetGraphqlApiInput{
		ApiId: aws.String(id),
	}

	output, err := conn.GetGraphqlApi(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func findSchemaCreationStatusByID(ctx context.Context, conn *appsync.Client, id string) (*appsync.GetSchemaCreationStatusOutput, error) {
	input := &appsync.GetSchemaCreationStatusInput{
		ApiId: aws.String(id),
	}

	output, err := conn.GetSchemaCreationStatus(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func statusSchemaCreation(ctx context.Context, conn *appsync.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSchemaCreationStatusByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitSchemaCreated(ctx context.Context, conn *appsync.Client, id string, timeout time.Duration) (*appsync.GetSchemaCreationStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SchemaStatusProcessing),
		Target:  enum.Slice(awstypes.SchemaStatusActive, awstypes.SchemaStatusSuccess),
		Refresh: statusSchemaCreation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appsync.GetSchemaCreationStatusOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Details)))
		return output, err
	}

	return nil, err
}

func expandLogConfig(tfList []any) *awstypes.LogConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.LogConfig{
		CloudWatchLogsRoleArn: aws.String(tfMap["cloudwatch_logs_role_arn"].(string)),
		ExcludeVerboseContent: tfMap["exclude_verbose_content"].(bool),
		FieldLogLevel:         awstypes.FieldLogLevel(tfMap["field_log_level"].(string)),
	}

	return apiObject
}

func expandOpenIDConnectConfig(tfList []any) *awstypes.OpenIDConnectConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.OpenIDConnectConfig{
		Issuer: aws.String(tfMap[names.AttrIssuer].(string)),
	}

	if v, ok := tfMap["auth_ttl"].(int); ok && v != 0 {
		apiObject.AuthTTL = int64(v)
	}

	if v, ok := tfMap[names.AttrClientID].(string); ok && v != "" {
		apiObject.ClientId = aws.String(v)
	}

	if v, ok := tfMap["iat_ttl"].(int); ok && v != 0 {
		apiObject.IatTTL = int64(v)
	}

	return apiObject
}

func expandUserPoolConfig(tfList []any, currentRegion string) *awstypes.UserPoolConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.UserPoolConfig{
		AwsRegion:     aws.String(currentRegion),
		DefaultAction: awstypes.DefaultAction(tfMap[names.AttrDefaultAction].(string)),
		UserPoolId:    aws.String(tfMap[names.AttrUserPoolID].(string)),
	}

	if v, ok := tfMap["app_id_client_regex"].(string); ok && v != "" {
		apiObject.AppIdClientRegex = aws.String(v)
	}

	if v, ok := tfMap["aws_region"].(string); ok && v != "" {
		apiObject.AwsRegion = aws.String(v)
	}

	return apiObject
}

func expandEnhancedMetricsConfig(tfList []any) *awstypes.EnhancedMetricsConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.EnhancedMetricsConfig{
		DataSourceLevelMetricsBehavior: awstypes.DataSourceLevelMetricsBehavior(tfMap["data_source_level_metrics_behavior"].(string)),
		OperationLevelMetricsConfig:    awstypes.OperationLevelMetricsConfig(tfMap["operation_level_metrics_config"].(string)),
		ResolverLevelMetricsBehavior:   awstypes.ResolverLevelMetricsBehavior(tfMap["resolver_level_metrics_behavior"].(string)),
	}

	return apiObject
}

func expandLambdaAuthorizerConfig(tfList []any) *awstypes.LambdaAuthorizerConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.LambdaAuthorizerConfig{
		AuthorizerResultTtlInSeconds: int32(tfMap["authorizer_result_ttl_in_seconds"].(int)),
		AuthorizerUri:                aws.String(tfMap["authorizer_uri"].(string)),
	}

	if v, ok := tfMap["identity_validation_expression"].(string); ok && v != "" {
		apiObject.IdentityValidationExpression = aws.String(v)
	}

	return apiObject
}

func expandAdditionalAuthenticationProviders(tfList []any, currentRegion string) []awstypes.AdditionalAuthenticationProvider {
	if len(tfList) < 1 {
		return nil
	}

	apiObjects := make([]awstypes.AdditionalAuthenticationProvider, 0)
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.AdditionalAuthenticationProvider{
			AuthenticationType: awstypes.AuthenticationType(tfMap["authentication_type"].(string)),
		}

		if v, ok := tfMap["lambda_authorizer_config"]; ok {
			apiObject.LambdaAuthorizerConfig = expandLambdaAuthorizerConfig(v.([]any))
		}

		if v, ok := tfMap["openid_connect_config"]; ok {
			apiObject.OpenIDConnectConfig = expandOpenIDConnectConfig(v.([]any))
		}

		if v, ok := tfMap["user_pool_config"]; ok {
			apiObject.UserPoolConfig = expandCognitoUserPoolConfig(v.([]any), currentRegion)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandCognitoUserPoolConfig(tfList []any, currentRegion string) *awstypes.CognitoUserPoolConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.CognitoUserPoolConfig{
		AwsRegion:  aws.String(currentRegion),
		UserPoolId: aws.String(tfMap[names.AttrUserPoolID].(string)),
	}

	if v, ok := tfMap["app_id_client_regex"].(string); ok && v != "" {
		apiObject.AppIdClientRegex = aws.String(v)
	}

	if v, ok := tfMap["aws_region"].(string); ok && v != "" {
		apiObject.AwsRegion = aws.String(v)
	}

	return apiObject
}

func flattenLogConfig(apiObject *awstypes.LogConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"cloudwatch_logs_role_arn": aws.ToString(apiObject.CloudWatchLogsRoleArn),
		"exclude_verbose_content":  apiObject.ExcludeVerboseContent,
		"field_log_level":          apiObject.FieldLogLevel,
	}

	return []any{tfMap}
}

func flattenOpenIDConnectConfig(apiObject *awstypes.OpenIDConnectConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"auth_ttl":         apiObject.AuthTTL,
		names.AttrClientID: aws.ToString(apiObject.ClientId),
		"iat_ttl":          apiObject.IatTTL,
		names.AttrIssuer:   aws.ToString(apiObject.Issuer),
	}

	return []any{tfMap}
}

func flattenUserPoolConfig(apiObject *awstypes.UserPoolConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"aws_region":            aws.ToString(apiObject.AwsRegion),
		names.AttrDefaultAction: apiObject.DefaultAction,
		names.AttrUserPoolID:    aws.ToString(apiObject.UserPoolId),
	}

	if apiObject.AppIdClientRegex != nil {
		tfMap["app_id_client_regex"] = aws.ToString(apiObject.AppIdClientRegex)
	}

	return []any{tfMap}
}

func flattenEnhancedMetricsConfig(apiObject *awstypes.EnhancedMetricsConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"data_source_level_metrics_behavior": apiObject.DataSourceLevelMetricsBehavior,
		"operation_level_metrics_config":     apiObject.OperationLevelMetricsConfig,
		"resolver_level_metrics_behavior":    apiObject.ResolverLevelMetricsBehavior,
	}

	return []any{tfMap}
}

func flattenLambdaAuthorizerConfig(apiObject *awstypes.LambdaAuthorizerConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"authorizer_result_ttl_in_seconds": apiObject.AuthorizerResultTtlInSeconds,
		"authorizer_uri":                   aws.ToString(apiObject.AuthorizerUri),
	}

	if apiObject.IdentityValidationExpression != nil {
		tfMap["identity_validation_expression"] = aws.ToString(apiObject.IdentityValidationExpression)
	}

	return []any{tfMap}
}

func flattenAdditionalAuthenticationProviders(apiObjects []awstypes.AdditionalAuthenticationProvider) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	tfList := make([]any, len(apiObjects))
	for i, apiObject := range apiObjects {
		tfList[i] = map[string]any{
			"authentication_type":      apiObject.AuthenticationType,
			"lambda_authorizer_config": flattenLambdaAuthorizerConfig(apiObject.LambdaAuthorizerConfig),
			"openid_connect_config":    flattenOpenIDConnectConfig(apiObject.OpenIDConnectConfig),
			"user_pool_config":         flattenCognitoUserPoolConfig(apiObject.UserPoolConfig),
		}
	}

	return tfList
}

func flattenCognitoUserPoolConfig(apiObject *awstypes.CognitoUserPoolConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"aws_region":         aws.ToString(apiObject.AwsRegion),
		names.AttrUserPoolID: aws.ToString(apiObject.UserPoolId),
	}

	if apiObject.AppIdClientRegex != nil {
		tfMap["app_id_client_regex"] = aws.ToString(apiObject.AppIdClientRegex)
	}

	return []any{tfMap}
}
