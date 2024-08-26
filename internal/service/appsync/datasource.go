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
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_datasource", name="Data Source")
func resourceDataSource() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
						names.AttrRegion: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrTableName: {
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
						names.AttrEndpoint: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRegion: {
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
										Default:          awstypes.AuthorizationTypeAwsIam,
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
						names.AttrEndpoint: {
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
						names.AttrFunctionARN: {
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
						names.AttrEndpoint: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "http_config", "lambda_config", "elasticsearch_config"},
			},
			names.AttrName: {
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
									names.AttrDatabaseName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"db_cluster_identifier": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									names.AttrSchema: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrSourceType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.RelationalDatabaseSourceTypeRdsHttpEndpoint,
							ValidateDiagFunc: enum.Validate[awstypes.RelationalDatabaseSourceType](),
						},
					},
				},
				ConflictsWith: []string{"dynamodb_config", "elasticsearch_config", "opensearchservice_config", "http_config", "lambda_config"},
			},
			names.AttrServiceRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DataSourceType](),
				StateFunc:        sdkv2.ToUpperSchemaStateFunc,
			},
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)
	region := meta.(*conns.AWSClient).Region

	apiID := d.Get("api_id").(string)
	name := d.Get(names.AttrName).(string)
	id := dataSourceCreateResourceID(apiID, name)
	input := &appsync.CreateDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
		Type:  awstypes.DataSourceType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	if v, ok := d.GetOk(names.AttrServiceRoleARN); ok {
		input.ServiceRoleArn = aws.String(v.(string))
	}

	_, err := conn.CreateDataSource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Data Source (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)

	apiID, name, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dataSource, err := findDataSourceByTwoPartKey(ctx, conn, apiID, name)

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Datasource %q not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Data Source (%s): %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set(names.AttrARN, dataSource.DataSourceArn)
	d.Set(names.AttrDescription, dataSource.Description)
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
	d.Set(names.AttrName, dataSource.Name)
	if err := d.Set("opensearchservice_config", flattenOpenSearchServiceDataSourceConfig(dataSource.OpenSearchServiceConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting opensearchservice_config: %s", err)
	}
	if err := d.Set("relational_database_config", flattenRelationalDatabaseDataSourceConfig(dataSource.RelationalDatabaseConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting relational_database_config: %s", err)
	}
	d.Set(names.AttrServiceRoleARN, dataSource.ServiceRoleArn)
	d.Set(names.AttrType, dataSource.Type)

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncClient(ctx)
	region := meta.(*conns.AWSClient).Region

	apiID, name, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &appsync.UpdateDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
		Type:  awstypes.DataSourceType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	if v, ok := d.GetOk(names.AttrServiceRoleARN); ok {
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

	apiID, name, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Appsync Data Source: %s", d.Id())
	_, err = conn.DeleteDataSource(ctx, &appsync.DeleteDataSourceInput{
		ApiId: aws.String(apiID),
		Name:  aws.String(name),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Data Source (%s): %s", d.Id(), err)
	}

	return diags
}

const dataSourceResourceIDSeparator = "-"

func dataSourceCreateResourceID(apiID, name string) string {
	parts := []string{apiID, name}
	id := strings.Join(parts, dataSourceResourceIDSeparator)

	return id
}

func dataSourceParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, dataSourceResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected API-ID%[2]sDATA-SOURCE-NAME", id, dataSourceResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDataSourceByTwoPartKey(ctx context.Context, conn *appsync.Client, apiID, name string) (*awstypes.DataSource, error) {
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

func expandDynamoDBDataSourceConfig(tfList []interface{}, currentRegion string) *awstypes.DynamodbDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.DynamodbDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		TableName: aws.String(tfMap[names.AttrTableName].(string)),
	}

	if v, ok := tfMap["delta_sync_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.DeltaSyncConfig = expandDeltaSyncConfig(v)
	}

	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.AwsRegion = aws.String(v.(string))
	}

	if v, ok := tfMap["use_caller_credentials"]; ok {
		apiObject.UseCallerCredentials = v.(bool)
	}

	if v, ok := tfMap["versioned"]; ok {
		apiObject.Versioned = v.(bool)
	}

	return apiObject
}

func expandDeltaSyncConfig(tfList []interface{}) *awstypes.DeltaSyncConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.DeltaSyncConfig{}

	if v, ok := tfMap["base_table_ttl"].(int); ok {
		apiObject.BaseTableTTL = int64(v)
	}

	if v, ok := tfMap["delta_sync_table_ttl"].(int); ok {
		apiObject.DeltaSyncTableTTL = int64(v)
	}

	if v, ok := tfMap["delta_sync_table_name"].(string); ok {
		apiObject.DeltaSyncTableName = aws.String(v)
	}

	return apiObject
}

func flattenDynamoDBDataSourceConfig(apiObject *awstypes.DynamodbDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrRegion:         aws.ToString(apiObject.AwsRegion),
		names.AttrTableName:      aws.ToString(apiObject.TableName),
		"use_caller_credentials": apiObject.UseCallerCredentials,
		"versioned":              apiObject.Versioned,
	}

	if apiObject.DeltaSyncConfig != nil {
		tfMap["delta_sync_config"] = flattenDeltaSyncConfig(apiObject.DeltaSyncConfig)
	}

	return []interface{}{tfMap}
}

func flattenDeltaSyncConfig(apiObject *awstypes.DeltaSyncConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"base_table_ttl":       apiObject.BaseTableTTL,
		"delta_sync_table_ttl": apiObject.DeltaSyncTableTTL,
	}

	if apiObject.DeltaSyncTableName != nil {
		tfMap["delta_sync_table_name"] = aws.ToString(apiObject.DeltaSyncTableName)
	}

	return []interface{}{tfMap}
}

func expandElasticsearchDataSourceConfig(tfList []interface{}, currentRegion string) *awstypes.ElasticsearchDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.ElasticsearchDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(tfMap[names.AttrEndpoint].(string)),
	}

	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.AwsRegion = aws.String(v.(string))
	}

	return apiObject
}

func expandOpenSearchServiceDataSourceConfig(tfList []interface{}, currentRegion string) *awstypes.OpenSearchServiceDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.OpenSearchServiceDataSourceConfig{
		AwsRegion: aws.String(currentRegion),
		Endpoint:  aws.String(tfMap[names.AttrEndpoint].(string)),
	}

	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.AwsRegion = aws.String(v.(string))
	}

	return apiObject
}

func flattenElasticsearchDataSourceConfig(apiObject *awstypes.ElasticsearchDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEndpoint: aws.ToString(apiObject.Endpoint),
		names.AttrRegion:   aws.ToString(apiObject.AwsRegion),
	}

	return []interface{}{tfMap}
}

func flattenOpenSearchServiceDataSourceConfig(apiObject *awstypes.OpenSearchServiceDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEndpoint: aws.ToString(apiObject.Endpoint),
		names.AttrRegion:   aws.ToString(apiObject.AwsRegion),
	}

	return []interface{}{tfMap}
}

func expandHTTPDataSourceConfig(tfList []interface{}) *awstypes.HttpDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.HttpDataSourceConfig{
		Endpoint: aws.String(tfMap[names.AttrEndpoint].(string)),
	}

	if v, ok := tfMap["authorization_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.AuthorizationConfig = expandAuthorizationConfig(v)
	}

	return apiObject
}

func flattenHTTPDataSourceConfig(apiObject *awstypes.HttpDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrEndpoint: aws.ToString(apiObject.Endpoint),
	}

	if apiObject.AuthorizationConfig != nil {
		tfMap["authorization_config"] = flattenAuthorizationConfig(apiObject.AuthorizationConfig)
	}

	return []interface{}{tfMap}
}

func expandAuthorizationConfig(tfList []interface{}) *awstypes.AuthorizationConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AuthorizationConfig{
		AuthorizationType: awstypes.AuthorizationType(tfMap["authorization_type"].(string)),
	}

	if v, ok := tfMap["aws_iam_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.AwsIamConfig = expandIAMConfig(v)
	}

	return apiObject
}

func flattenAuthorizationConfig(apiObject *awstypes.AuthorizationConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"authorization_type": apiObject.AuthorizationType,
	}

	if apiObject.AwsIamConfig != nil {
		tfMap["aws_iam_config"] = flattenIAMConfig(apiObject.AwsIamConfig)
	}

	return []interface{}{tfMap}
}

func expandIAMConfig(tfList []interface{}) *awstypes.AwsIamConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.AwsIamConfig{}

	if v, ok := tfMap["signing_region"].(string); ok && v != "" {
		apiObject.SigningRegion = aws.String(v)
	}

	if v, ok := tfMap["signing_service_name"].(string); ok && v != "" {
		apiObject.SigningServiceName = aws.String(v)
	}

	return apiObject
}

func flattenIAMConfig(apiObject *awstypes.AwsIamConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"signing_region":       aws.ToString(apiObject.SigningRegion),
		"signing_service_name": aws.ToString(apiObject.SigningServiceName),
	}

	return []interface{}{tfMap}
}

func expandLambdaDataSourceConfig(tfList []interface{}) *awstypes.LambdaDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.LambdaDataSourceConfig{
		LambdaFunctionArn: aws.String(tfMap[names.AttrFunctionARN].(string)),
	}

	return apiObject
}

func flattenLambdaDataSourceConfig(apiObject *awstypes.LambdaDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrFunctionARN: aws.ToString(apiObject.LambdaFunctionArn),
	}

	return []interface{}{tfMap}
}

func expandRelationalDatabaseDataSourceConfig(tfList []interface{}, currentRegion string) *awstypes.RelationalDatabaseDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.RelationalDatabaseDataSourceConfig{
		RelationalDatabaseSourceType: awstypes.RelationalDatabaseSourceType(tfMap[names.AttrSourceType].(string)),
		RdsHttpEndpointConfig:        expandRDSHTTPEndpointConfig(tfMap["http_endpoint_config"].([]interface{}), currentRegion),
	}

	return apiObject
}

func flattenRelationalDatabaseDataSourceConfig(apiObject *awstypes.RelationalDatabaseDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrSourceType:   apiObject.RelationalDatabaseSourceType,
		"http_endpoint_config": flattenRDSHTTPEndpointConfig(apiObject.RdsHttpEndpointConfig),
	}

	return []interface{}{tfMap}
}

func expandEventBridgeDataSourceConfig(tfList []interface{}) *awstypes.EventBridgeDataSourceConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.EventBridgeDataSourceConfig{
		EventBusArn: aws.String(tfMap["event_bus_arn"].(string)),
	}

	return apiObject
}

func flattenEventBridgeDataSourceConfig(apiObject *awstypes.EventBridgeDataSourceConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"event_bus_arn": aws.ToString(apiObject.EventBusArn),
	}

	return []interface{}{tfMap}
}

func expandRDSHTTPEndpointConfig(tfList []interface{}, currentRegion string) *awstypes.RdsHttpEndpointConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.RdsHttpEndpointConfig{
		AwsRegion: aws.String(currentRegion),
	}

	if v, ok := tfMap[names.AttrRegion]; ok && v.(string) != "" {
		apiObject.AwsRegion = aws.String(v.(string))
	}

	if v, ok := tfMap["aws_secret_store_arn"]; ok && v.(string) != "" {
		apiObject.AwsSecretStoreArn = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrDatabaseName]; ok && v.(string) != "" {
		apiObject.DatabaseName = aws.String(v.(string))
	}

	if v, ok := tfMap["db_cluster_identifier"]; ok && v.(string) != "" {
		apiObject.DbClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := tfMap[names.AttrSchema]; ok && v.(string) != "" {
		apiObject.Schema = aws.String(v.(string))
	}

	return apiObject
}

func flattenRDSHTTPEndpointConfig(apiObject *awstypes.RdsHttpEndpointConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AwsRegion != nil {
		tfMap[names.AttrRegion] = aws.ToString(apiObject.AwsRegion)
	}

	if apiObject.AwsSecretStoreArn != nil {
		tfMap["aws_secret_store_arn"] = aws.ToString(apiObject.AwsSecretStoreArn)
	}

	if apiObject.DatabaseName != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(apiObject.DatabaseName)
	}

	if apiObject.DbClusterIdentifier != nil {
		tfMap["db_cluster_identifier"] = aws.ToString(apiObject.DbClusterIdentifier)
	}

	if apiObject.Schema != nil {
		tfMap[names.AttrSchema] = aws.ToString(apiObject.Schema)
	}

	return []interface{}{tfMap}
}
