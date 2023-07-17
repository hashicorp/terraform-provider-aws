// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
										Type:         schema.TypeString,
										Optional:     true,
										Default:      appsync.AuthorizationTypeAwsIam,
										ValidateFunc: validation.StringInSlice(appsync.AuthorizationType_Values(), true),
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`), "must match [_A-Za-z][_0-9A-Za-z]*"),
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      appsync.RelationalDatabaseSourceTypeRdsHttpEndpoint,
							ValidateFunc: validation.StringInSlice(appsync.RelationalDatabaseSourceType_Values(), true),
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
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.DataSourceType_Values(), true),
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
			},
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get("name").(string)
	input := &appsync.CreateDataSourceInput{
		ApiId: aws.String(d.Get("api_id").(string)),
		Name:  aws.String(name),
		Type:  aws.String(d.Get("type").(string)),
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

	_, err := conn.CreateDataSourceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Data Source (%s): %s", name, err)
	}

	d.SetId(d.Get("api_id").(string) + "-" + d.Get("name").(string))

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

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
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)
	region := meta.(*conns.AWSClient).Region

	apiID, name, err := DecodeID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
		Type:  aws.String(d.Get("type").(string)),
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

	_, err = conn.UpdateDataSourceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Data Source (%s): %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, name, err := DecodeID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.DeleteDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}

	_, err = conn.DeleteDataSourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Data Source (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDataSourceByTwoPartKey(ctx context.Context, conn *appsync.AppSync, apiID, name string) (*appsync.DataSource, error) {
	input := &appsync.GetDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	}

	output, err := conn.GetDataSourceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
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

func expandDynamoDBDataSourceConfig(l []interface{}, currentRegion string) *appsync.DynamodbDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.DynamodbDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		TableName: aws.String(configured["table_name"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	if v, ok := configured["use_caller_credentials"]; ok {
		result.UseCallerCredentials = aws.Bool(v.(bool))
	}

	if v, ok := configured["versioned"]; ok {
		result.Versioned = aws.Bool(v.(bool))
	}

	if v, ok := configured["delta_sync_config"].([]interface{}); ok && len(v) > 0 {
		result.DeltaSyncConfig = expandDynamoDBDataSourceDeltaSyncConfig(v)
	}

	return result
}

func expandDynamoDBDataSourceDeltaSyncConfig(l []interface{}) *appsync.DeltaSyncConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.DeltaSyncConfig{}

	if v, ok := configured["base_table_ttl"].(int); ok {
		result.BaseTableTTL = aws.Int64(int64(v))
	}

	if v, ok := configured["delta_sync_table_ttl"].(int); ok {
		result.DeltaSyncTableTTL = aws.Int64(int64(v))
	}

	if v, ok := configured["delta_sync_table_name"].(string); ok {
		result.DeltaSyncTableName = aws.String(v)
	}

	return result
}

func flattenDynamoDBDataSourceConfig(config *appsync.DynamodbDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"region":     aws.StringValue(config.AwsRegion),
		"table_name": aws.StringValue(config.TableName),
	}

	if config.UseCallerCredentials != nil {
		result["use_caller_credentials"] = aws.BoolValue(config.UseCallerCredentials)
	}

	if config.Versioned != nil {
		result["versioned"] = aws.BoolValue(config.Versioned)
	}

	if config.DeltaSyncConfig != nil {
		result["delta_sync_config"] = flattenDynamoDBDataSourceDeltaSyncConfig(config.DeltaSyncConfig)
	}

	return []map[string]interface{}{result}
}

func flattenDynamoDBDataSourceDeltaSyncConfig(config *appsync.DeltaSyncConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.DeltaSyncTableName != nil {
		result["delta_sync_table_name"] = aws.StringValue(config.DeltaSyncTableName)
	}

	if config.BaseTableTTL != nil {
		result["base_table_ttl"] = aws.Int64Value(config.BaseTableTTL)
	}

	if config.DeltaSyncTableTTL != nil {
		result["delta_sync_table_ttl"] = aws.Int64Value(config.DeltaSyncTableTTL)
	}

	return []map[string]interface{}{result}
}

func expandElasticsearchDataSourceConfig(l []interface{}, currentRegion string) *appsync.ElasticsearchDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.ElasticsearchDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	return result
}

func expandOpenSearchServiceDataSourceConfig(l []interface{}, currentRegion string) *appsync.OpenSearchServiceDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.OpenSearchServiceDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["region"]; ok && v.(string) != "" {
		result.AwsRegion = aws.String(v.(string))
	}

	return result
}

func flattenElasticsearchDataSourceConfig(config *appsync.ElasticsearchDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.StringValue(config.Endpoint),
		"region":   aws.StringValue(config.AwsRegion),
	}

	return []map[string]interface{}{result}
}

func flattenOpenSearchServiceDataSourceConfig(config *appsync.OpenSearchServiceDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.StringValue(config.Endpoint),
		"region":   aws.StringValue(config.AwsRegion),
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceConfig(l []interface{}) *appsync.HttpDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.HttpDataSourceConfig{
		Endpoint: aws.String(configured["endpoint"].(string)),
	}

	if v, ok := configured["authorization_config"].([]interface{}); ok && len(v) > 0 {
		result.AuthorizationConfig = expandHTTPDataSourceAuthorizationConfig(v)
	}

	return result
}

func flattenHTTPDataSourceConfig(config *appsync.HttpDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"endpoint": aws.StringValue(config.Endpoint),
	}

	if config.AuthorizationConfig != nil {
		result["authorization_config"] = flattenHTTPDataSourceAuthorizationConfig(config.AuthorizationConfig)
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceAuthorizationConfig(l []interface{}) *appsync.AuthorizationConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.AuthorizationConfig{
		AuthorizationType: aws.String(configured["authorization_type"].(string)),
	}

	if v, ok := configured["aws_iam_config"].([]interface{}); ok && len(v) > 0 {
		result.AwsIamConfig = expandHTTPDataSourceIAMConfig(v)
	}

	return result
}

func flattenHTTPDataSourceAuthorizationConfig(config *appsync.AuthorizationConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"authorization_type": aws.StringValue(config.AuthorizationType),
	}

	if config.AwsIamConfig != nil {
		result["aws_iam_config"] = flattenHTTPDataSourceIAMConfig(config.AwsIamConfig)
	}

	return []map[string]interface{}{result}
}

func expandHTTPDataSourceIAMConfig(l []interface{}) *appsync.AwsIamConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.AwsIamConfig{}

	if v, ok := configured["signing_region"].(string); ok && v != "" {
		result.SigningRegion = aws.String(v)
	}

	if v, ok := configured["signing_service_name"].(string); ok && v != "" {
		result.SigningServiceName = aws.String(v)
	}

	return result
}

func flattenHTTPDataSourceIAMConfig(config *appsync.AwsIamConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"signing_region":       aws.StringValue(config.SigningRegion),
		"signing_service_name": aws.StringValue(config.SigningServiceName),
	}

	return []map[string]interface{}{result}
}

func expandLambdaDataSourceConfig(l []interface{}) *appsync.LambdaDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.LambdaDataSourceConfig{
		LambdaFunctionArn: aws.String(configured["function_arn"].(string)),
	}

	return result
}

func flattenLambdaDataSourceConfig(config *appsync.LambdaDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"function_arn": aws.StringValue(config.LambdaFunctionArn),
	}

	return []map[string]interface{}{result}
}

func expandRelationalDatabaseDataSourceConfig(l []interface{}, currentRegion string) *appsync.RelationalDatabaseDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.RelationalDatabaseDataSourceConfig{
		RelationalDatabaseSourceType: aws.String(configured["source_type"].(string)),
		RdsHttpEndpointConfig:        testAccDataSourceConfig_expandRDSHTTPEndpoint(configured["http_endpoint_config"].([]interface{}), currentRegion),
	}

	return result
}

func flattenRelationalDatabaseDataSourceConfig(config *appsync.RelationalDatabaseDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"source_type":          aws.StringValue(config.RelationalDatabaseSourceType),
		"http_endpoint_config": flattenRDSHTTPEndpointConfig(config.RdsHttpEndpointConfig),
	}

	return []map[string]interface{}{result}
}

func expandEventBridgeDataSourceConfig(l []interface{}) *appsync.EventBridgeDataSourceConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.EventBridgeDataSourceConfig{
		EventBusArn: aws.String(configured["event_bus_arn"].(string)),
	}

	return result
}

func flattenEventBridgeDataSourceConfig(config *appsync.EventBridgeDataSourceConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"event_bus_arn": aws.StringValue(config.EventBusArn),
	}

	return []map[string]interface{}{result}
}

func testAccDataSourceConfig_expandRDSHTTPEndpoint(l []interface{}, currentRegion string) *appsync.RdsHttpEndpointConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.RdsHttpEndpointConfig{
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

func flattenRDSHTTPEndpointConfig(config *appsync.RdsHttpEndpointConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.AwsRegion != nil {
		result["region"] = aws.StringValue(config.AwsRegion)
	}

	if config.AwsSecretStoreArn != nil {
		result["aws_secret_store_arn"] = aws.StringValue(config.AwsSecretStoreArn)
	}

	if config.DatabaseName != nil {
		result["database_name"] = aws.StringValue(config.DatabaseName)
	}

	if config.DbClusterIdentifier != nil {
		result["db_cluster_identifier"] = aws.StringValue(config.DbClusterIdentifier)
	}

	if config.Schema != nil {
		result["schema"] = aws.StringValue(config.Schema)
	}

	return []map[string]interface{}{result}
}
