// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_appsync_datasource")
func ResourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSourceCreate,
		ReadWithoutTimeout:   resourceDataSourceRead,
		UpdateWithoutTimeout: resourceDataSourceUpdate,
		DeleteWithoutTimeout: resourceDataSourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dynamodb_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delta_sync_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"base_table_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"delta_sync_table_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"delta_sync_table_ttl": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"use_caller_credentials": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"versioned": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{"elasticsearch_config", "http_config", "lambda_config", "relational_database_config", "opensearchservice_config"},
			},
			"elasticsearch_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "http_config", "lambda_config", "opensearchservice_config"},
			},
			"event_bridge_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_bus_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "elasticsearch_config", "http_config", "lambda_config", "relational_database_config"},
			},
			"http_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorization_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_type": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          string(awstypes.AuthorizationTypeAwsIam),
										ValidateDiagFunc: enum.Validate[awstypes.AuthorizationType](),
									},
									"aws_iam_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"signing_region": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"signing_service_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "elasticsearch_config", "opensearchservice_config", "lambda_config", "relational_database_config"},
			},
			"lambda_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "elasticsearch_config", "opensearchservice_config", "http_config", "relational_database_config"},
			},
			"opensearchservice_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "http_config", "lambda_config", "elasticsearch_config"},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[A-Za-z_][0-9A-Za-z_]*`), "must match [A-Za-z_][0-9A-Za-z_]*"),
			},
			"relational_database_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_secret_store_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"database_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"db_cluster_identifier": {
										Type:     schema.TypeString,
										Required: true,
									},
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"schema": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"source_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          string(awstypes.RelationalDatabaseSourceTypeRdsHttpEndpoint),
							ValidateDiagFunc: enum.Validate[awstypes.RelationalDatabaseSourceType](),
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "elasticsearch_config", "opensearchservice_config", "http_config", "lambda_config"},
			},
			"service_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DataSourceType](),
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
			},
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get("name").(string)
	input := &appsync.CreateDataSourceInput{
		ApiId: aws.String(d.Get("api_id").(string)),
		Name:  aws.String(name),
		Type:  awstypes.DataSourceType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dynamodb_config"); ok {
		input.DynamodbConfig = expandDynamoDBDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("elasticsearch_config"); ok {
		input.ElasticsearchConfig = expandElasticsearchDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("event_bridge_config"); ok {
		input.EventBridgeConfig = expandEventBridgeDataSourceConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("http_config"); ok {
		input.HttpConfig = expandHTTPDataSourceConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("lambda_config"); ok {
		input.LambdaConfig = expandLambdaDataSourceConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("opensearchservice_config"); ok {
		input.OpenSearchServiceConfig = expandOpenSearchServiceDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("relational_database_config"); ok {
		input.RelationalDatabaseConfig = expandRelationalDatabaseDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("service_role_arn"); ok {
		input.ServiceRoleArn = aws.String(v.(string))
	}

	_, err := conn.CreateDataSource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Data Source (%s): %s", name, err)
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("name").(string))

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, name, err := DecodeID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dataSource, err := FindDataSourceByTwoPartKey(ctx, conn, apiID, name)

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Datasource %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Data Source (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("arn", dataSource.DataSourceArn)
	d.Set("description", dataSource.Description)
	if err := d.Set("dynamodb_config", flattenDynamoDBDataSourceConfig(dataSource.DynamodbConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dynamodb_config: %s", err)
	}
	if err := d.Set("elasticsearch_config", flattenElasticsearchDataSourceConfig(dataSource.ElasticsearchConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting elasticsearch_config: %s", err)
	}
	if err := d.Set("event_bridge_config", flattenEventBridgeDataSourceConfig(dataSource.EventBridgeConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_bridge_config: %s", err)
	}
	if err := d.Set("http_config", flattenHTTPDataSourceConfig(dataSource.HttpConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting http_config: %s", err)
	}
	if err := d.Set("lambda_config", flattenLambdaDataSourceConfig(dataSource.LambdaConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lambda_config: %s", err)
	}
	d.Set("name", dataSource.Name)
	if err := d.Set("opensearchservice_config", flattenOpenSearchServiceDataSourceConfig(dataSource.OpenSearchServiceConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting opensearchservice_config: %s", err)
	}
	if err := d.Set("relational_database_config", flattenRelationalDatabaseDataSourceConfig(dataSource.RelationalDatabaseConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting relational_database_config: %s", err)
	}
	d.Set("service_role_arn", dataSource.ServiceRoleArn)
	d.Set("type", dataSource.Type)

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)
	region := meta.(*conns.AWSClient).Region

	apiID, name, err := DecodeID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
		Type:  awstypes.DataSourceType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("dynamodb_config"); ok {
		input.DynamodbConfig = expandDynamoDBDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("elasticsearch_config"); ok {
		input.ElasticsearchConfig = expandElasticsearchDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("http_config"); ok {
		input.HttpConfig = expandHTTPDataSourceConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("lambda_config"); ok {
		input.LambdaConfig = expandLambdaDataSourceConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("opensearchservice_config"); ok {
		input.OpenSearchServiceConfig = expandOpenSearchServiceDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("relational_database_config"); ok {
		input.RelationalDatabaseConfig = expandRelationalDatabaseDataSourceConfig(v.([]interface{}), region)
	}

	if v, ok := d.GetOk("service_role_arn"); ok {
		input.ServiceRoleArn = aws.String(v.(string))
	}

	_, err = conn.UpdateDataSource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Data Source (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, name, err := DecodeID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.DeleteDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}

	_, err = conn.DeleteDataSource(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Data Source (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDataSourceByTwoPartKey(ctx context.Context, conn *appsync.Client, apiID, name string) (*awstypes.DataSource, error) {
	input := &appsync.GetDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}

	output, err := conn.GetDataSource(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSource, nil
}

func DecodeID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format ApiID-DataSourceName, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}

func expandDynamoDBDataSourceConfig(l []interface{}, currentRegion string) *awstypes.DynamodbDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.DynamodbDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		TableName: aws.String(configured["table_name"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	if v, ok := configured["use_caller_credentials"]; ok {
		result.UseCallerCredentials = v.(bool)
	}

	if v, ok := configured["versioned"]; ok {
		result.Versioned = v.(bool)
	}

	if v, ok := configured["delta_sync_config"].([]interface{}); ok && len(v) > 0 {
		result.DeltaSyncConfig = expandDynamoDBDataSourceDeltaSyncConfig(v)
	}

	return result
}

func expandDynamoDBDataSourceDeltaSyncConfig(l []interface{}) *awstypes.DeltaSyncConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.DeltaSyncConfig{}

	if v, ok := configured["base_table_ttl"].(int); ok {
		result.BaseTableTTL = int64(v)
	}

	if v, ok := configured["delta_sync_table_ttl"].(int); ok {
		result.DeltaSyncTableTTL = int64(v)
	}

	if v, ok := configured["delta_sync_table_name"].(string); ok {
		result.DeltaSyncTableName = aws.String(v)
	}

	return result
}

func flattenDynamoDBDataSourceConfig(config *awstypes.DynamodbDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"region":     aws.ToString(config.AwsRegion),
		"table_name": aws.ToString(config.TableName),
	}

	result["use_caller_credentials"] = config.UseCallerCredentials
	result["versioned"] = config.Versioned

	if config.DeltaSyncConfig != nil {
		result["delta_sync_config"] = flattenDynamoDBDataSourceDeltaSyncConfig(config.DeltaSyncConfig)
	}

	return []map[string]interface{}{result}
}

func flattenDynamoDBDataSourceDeltaSyncConfig(config *awstypes.DeltaSyncConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.DeltaSyncTableName != nil {
		result["delta_sync_table_name"] = aws.ToString(config.DeltaSyncTableName)
	}

	result["base_table_ttl"] = config.BaseTableTTL
	result["delta_sync_table_ttl"] = config.DeltaSyncTableTTL

	return []map[string]interface{}{result}
}

func expandElasticsearchDataSourceConfig(l []interface{}, currentRegion string) *awstypes.ElasticsearchDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.ElasticsearchDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	return result
}

func expandOpenSearchServiceDataSourceConfig(l []interface{}, currentRegion string) *awstypes.OpenSearchServiceDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.OpenSearchServiceDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	return result
}

func flattenElasticsearchDataSourceConfig(config *awstypes.ElasticsearchDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.ToString(config.Endpoint),
		"region":   aws.ToString(config.AwsRegion),
	}

	return []map[string]interface{}{result}
}

func flattenOpenSearchServiceDataSourceConfig(config *awstypes.OpenSearchServiceDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.ToString(config.Endpoint),
		"region":   aws.ToString(config.AwsRegion),
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceConfig(l []interface{}) *awstypes.HttpDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.HttpDataSourceConfig{
		Endpoint: aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["authorization_config"].([]interface{}); ok && len(v) > 0 {
		result.AuthorizationConfig = expandHTTPDataSourceAuthorizationConfig(v)
	}

	return result
}

func flattenHTTPDataSourceConfig(config *awstypes.HttpDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.ToString(config.Endpoint),
	}

	if config.AuthorizationConfig != nil {
		result["authorization_config"] = flattenHTTPDataSourceAuthorizationConfig(config.AuthorizationConfig)
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceAuthorizationConfig(l []interface{}) *awstypes.AuthorizationConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.AuthorizationConfig{
		AuthorizationType: awstypes.AuthorizationType(configured["authorization_type"].(string)),
	}

	if v, ok := configured["aws_iam_config"].([]interface{}); ok && len(v) > 0 {
		result.AwsIamConfig = expandHTTPDataSourceIAMConfig(v)
	}

	return result
}

func flattenHTTPDataSourceAuthorizationConfig(config *awstypes.AuthorizationConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"authorization_type": string(config.AuthorizationType),
	}

	if config.AwsIamConfig != nil {
		result["aws_iam_config"] = flattenHTTPDataSourceIAMConfig(config.AwsIamConfig)
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceIAMConfig(l []interface{}) *awstypes.AwsIamConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.AwsIamConfig{}

	if v, ok := configured["signing_region"].(string); ok && v != "" {
		result.SigningRegion = aws.String(v)
	}

	if v, ok := configured["signing_service_name"].(string); ok && v != "" {
		result.SigningServiceName = aws.String(v)
	}

	return result
}

func flattenHTTPDataSourceIAMConfig(config *awstypes.AwsIamConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"signing_region":       aws.ToString(config.SigningRegion),
		"signing_service_name": aws.ToString(config.SigningServiceName),
	}

	return []map[string]interface{}{result}
}

func expandLambdaDataSourceConfig(l []interface{}) *awstypes.LambdaDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.LambdaDataSourceConfig{
		LambdaFunctionArn: aws.String(configured["function_arn"].(string)),
	}

	return result
}

func flattenLambdaDataSourceConfig(config *awstypes.LambdaDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"function_arn": aws.ToString(config.LambdaFunctionArn),
	}

	return []map[string]interface{}{result}
}

func expandRelationalDatabaseDataSourceConfig(l []interface{}, currentRegion string) *awstypes.RelationalDatabaseDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.RelationalDatabaseDataSourceConfig{
		RelationalDatabaseSourceType: awstypes.RelationalDatabaseSourceType(configured["source_type"].(string)),
		RdsHttpEndpointConfig:        testAccDataSourceConfig_expandRDSHTTPEndpoint(configured["http_endpoint_config"].([]interface{}), currentRegion),
	}

	return result
}

func flattenRelationalDatabaseDataSourceConfig(config *awstypes.RelationalDatabaseDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"source_type":          string(config.RelationalDatabaseSourceType),
		"http_endpoint_config": flattenRDSHTTPEndpointConfig(config.RdsHttpEndpointConfig),
	}

	return []map[string]interface{}{result}
}

func expandEventBridgeDataSourceConfig(l []interface{}) *awstypes.EventBridgeDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.EventBridgeDataSourceConfig{
		EventBusArn: aws.String(configured["event_bus_arn"].(string)),
	}

	return result
}

func flattenEventBridgeDataSourceConfig(config *awstypes.EventBridgeDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"event_bus_arn": aws.ToString(config.EventBusArn),
	}

	return []map[string]interface{}{result}
}

func testAccDataSourceConfig_expandRDSHTTPEndpoint(l []interface{}, currentRegion string) *awstypes.RdsHttpEndpointConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &awstypes.RdsHttpEndpointConfig{
		AwsRegion: aws.String(currentRegion),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	if v, ok := configured["aws_secret_store_arn"]; ok && v.(string) != "" {
		result.AwsSecretStoreArn = aws.String(v.(string))
	}

	if v, ok := configured["database_name"]; ok && v.(string) != "" {
		result.DatabaseName = aws.String(v.(string))
	}

	if v, ok := configured["db_cluster_identifier"]; ok && v.(string) != "" {
		result.DbClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := configured["schema"]; ok && v.(string) != "" {
		result.Schema = aws.String(v.(string))
	}

	return result
}

func flattenRDSHTTPEndpointConfig(config *awstypes.RdsHttpEndpointConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.AwsRegion != nil {
		result["region"] = aws.ToString(config.AwsRegion)
	}

	if config.AwsSecretStoreArn != nil {
		result["aws_secret_store_arn"] = aws.ToString(config.AwsSecretStoreArn)
	}

	if config.DatabaseName != nil {
		result["database_name"] = aws.ToString(config.DatabaseName)
	}

	if config.DbClusterIdentifier != nil {
		result["db_cluster_identifier"] = aws.ToString(config.DbClusterIdentifier)
	}

	if config.Schema != nil {
		result["schema"] = aws.ToString(config.Schema)
	}

	return []map[string]interface{}{result}
}
