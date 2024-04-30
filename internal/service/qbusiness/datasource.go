// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/document"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @FrameworkResource("aws_qbusiness_datasource", name="Datasource")
// @Tags(identifierAttribute="arn")
func newResourceDatasource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDatasource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDatasource = "Datasource"
)

type resourceDatasource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceDatasource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_datasource"
}

func (r *resourceDatasource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
}

func (r *resourceDatasource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

func (r *resourceDatasource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *resourceDatasource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceDatasource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceDatasource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func keySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 200),
			validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`),
				"must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters")),
	}
}

func valueSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
				"long_value": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"string_list_value": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 2048,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"string_value": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
					),
				},
			},
		},
	}
}

func documentAttributeConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": keySchema(),
				"operator": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.DocumentEnrichmentConditionOperator](),
				},
				"value": valueSchema(),
			},
		},
	}
}

func hookConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"invocation_condition": documentAttributeConditionSchema(),
			"lambda_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_bucket_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name")),
			},
		}},
	}
}

/*
// @SDKResource("aws_qbusiness_datasource", name="Datasource")
// @Tags(identifierAttribute="arn")

	func ResourceDatasource() *schema.Resource {
		return &schema.Resource{

			CreateWithoutTimeout: resourceDatasourceCreate,
			ReadWithoutTimeout:   resourceDatasourceRead,
			UpdateWithoutTimeout: resourceDatasourceUpdate,
			DeleteWithoutTimeout: resourceDatasourceDelete,

			Importer: &schema.ResourceImporter{
				StateContext: schema.ImportStatePassthroughContext,
			},

			Timeouts: &schema.ResourceTimeout{
				Delete:  schema.DefaultTimeout(40 * time.Minute),
				Default: schema.DefaultTimeout(10 * time.Minute),
			},

			Schema: map[string]*schema.Schema{
				"application_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Identifier of the Amazon Q application the data source will be attached to.",
					ValidateFunc: validation.All(
						validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
					),
				},
				"arn": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ARN of the Amazon Q datasource.",
				},
				"configuration": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Configuration information (JSON) to connect to your data source repository.",
					ValidateFunc: validation.StringIsJSON,
				},
				"datasource_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Datasource identifier",
				},
				"description": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Description for the data source connector.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 1000),
						validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
					),
				},
				"display_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Name for the data source connector.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 100),
						validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
					),
				},
				"document_enrichment_configuration": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Description: "Configuration information for altering document metadata and content during the document ingestion process.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"inline_configurations": {
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Description: "Information to alter document attributes or metadata fields and content when ingesting documents into Amazon Q",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"configuration": {
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 100,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"condition": documentAttributeConditionSchema(),
													"document_content_operator": {
														Type:             schema.TypeString,
														Optional:         true,
														ValidateDiagFunc: enum.Validate[types.DocumentContentOperator](),
													},
													"target": {
														Type:     schema.TypeList,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"key": keySchema(),
																"attribute_value_operator": {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[types.AttributeValueOperator](),
																},
																"value": valueSchema(),
															},
														},
													},
												},
											},
										},
									},
								},
							},
							"post_extraction_hook_configuration": hookConfigurationSchema(),
							"pre_extraction_hook_configuration":  hookConfigurationSchema(),
						},
					},
				},
				"iam_service_role_arn": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "ARN of an IAM role with permission to access the data source and required resources.",
					ValidateFunc: verify.ValidARN,
				},
				"index_id": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Identifier of the index that you want to use with the data source connector.",
					ValidateFunc: validation.All(
						validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
					),
				},
				"sync_schedule": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Frequency for Amazon Q to check the documents in your data source repository and update your index.",
					ValidateFunc: validation.All(
						validation.StringLenBetween(0, 998),
						validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
					),
				},
				"vpc_configuration": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "Information for an VPC to connect to your data source.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"security_group_ids": {
								Type:     schema.TypeSet,
								Required: true,
								MaxItems: 200,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"subnet_ids": {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 200,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			},
			CustomizeDiff: verify.SetTagsDiff,
		}
	}
*/
func resourceDatasourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id := d.Get("application_id").(string)
	index_id := d.Get("index_id").(string)

	var conf map[string]interface{}

	err := json.Unmarshal([]byte(d.Get("configuration").(string)), &conf)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "unmarshal configuration: %s", err)
	}

	input := qbusiness.CreateDataSourceInput{
		ApplicationId: aws.String(application_id),
		IndexId:       aws.String(index_id),
		DisplayName:   aws.String(d.Get("display_name").(string)),
		Configuration: document.NewLazyDocument(conf),
		RoleArn:       aws.String(d.Get("iam_service_role_arn").(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("document_enrichment_configuration"); ok {
		input.DocumentEnrichmentConfiguration = expandDocumentEnrichmentConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("sync_schedule"); ok {
		input.SyncSchedule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_configuration"); ok {
		input.VpcConfiguration = expandVPCConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDataSource(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amazon Q datasource: %s", err)
	}

	d.SetId(application_id + "/" + index_id + "/" + aws.ToString(output.DataSourceId))

	if _, err := waitDatasourceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness datasource (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceDatasourceRead(ctx, d, meta)...)
}

func resourceDatasourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	output, err := FindDatasourceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] qbusiness datasource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading qbusiness datasource (%s): %s", d.Id(), err)
	}

	conf, err := output.Configuration.MarshalSmithyDocument()

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "marshal qbusiness datasource configuration: %s", err)
	}

	d.Set("application_id", output.ApplicationId)
	d.Set("arn", output.DataSourceArn)
	d.Set("configuration", string(conf))
	d.Set("document_enrichment_configuration", flattenDocumentEnrichmentConfiguration(output.DocumentEnrichmentConfiguration))
	d.Set("iam_service_role_arn", output.RoleArn)
	d.Set("index_id", output.IndexId)
	d.Set("datasource_id", output.DataSourceId)
	d.Set("display_name", output.DisplayName)
	d.Set("description", output.Description)
	d.Set("sync_schedule", output.SyncSchedule)
	d.Set("vpc_configuration", flattenVPCConfiguration(output.VpcConfiguration))

	return diags
}

func resourceDatasourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, index_id, datasource_id, err := parseDatasourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parse qbusiness datasource ID: %s", err)
	}

	input := qbusiness.UpdateDataSourceInput{
		ApplicationId: aws.String(application_id),
		IndexId:       aws.String(index_id),
		DataSourceId:  aws.String(datasource_id),
	}

	if d.HasChange("document_enrichment_configuration") {
		input.DocumentEnrichmentConfiguration = expandDocumentEnrichmentConfiguration(d.Get("document_enrichment_configuration").([]interface{}))
	}

	if d.HasChange("configuration") {
		var conf map[string]interface{}
		err = json.Unmarshal([]byte(d.Get("configuration").(string)), &conf)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unmarshal qbusiness datasource configuration: %s", err)
		}

		input.Configuration = document.NewLazyDocument(conf)
	}

	if d.HasChange("display_name") {
		input.DisplayName = aws.String(d.Get("display_name").(string))
	}

	if d.HasChange("iam_service_role_arn") {
		input.RoleArn = aws.String(d.Get("iam_service_role_arn").(string))
	}

	if d.HasChange("sync_schedule") {
		input.SyncSchedule = aws.String(d.Get("sync_schedule").(string))
	}

	if d.HasChange("vpc_configuration") {
		input.VpcConfiguration = expandVPCConfiguration(d.Get("vpc_configuration").([]interface{}))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	_, err = conn.UpdateDataSource(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness datasource (%s): %s", d.Id(), err)
	}

	if _, err := waitDatasourceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness datasource (%s) to be updated: %s", d.Id(), err)
	}

	return append(diags, resourceDatasourceRead(ctx, d, meta)...)
}

func resourceDatasourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessClient(ctx)

	application_id, index_id, datasource_id, err := parseDatasourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parse qbusiness datasource ID: %s", err)
	}

	input := qbusiness.DeleteDataSourceInput{
		ApplicationId: aws.String(application_id),
		IndexId:       aws.String(index_id),
		DataSourceId:  aws.String(datasource_id),
	}

	_, err = conn.DeleteDataSource(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness datasource (%s): %s", d.Id(), err)
	}

	if _, err := waitDatasourceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness datasource (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func parseDatasourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid datasource ID: %s", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func FindDatasourceByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetDataSourceOutput, error) {
	application_id, index_id, datasource_id, err := parseDatasourceID(id)

	if err != nil {
		return nil, err
	}

	input := qbusiness.GetDataSourceInput{
		ApplicationId: aws.String(application_id),
		IndexId:       aws.String(index_id),
		DataSourceId:  aws.String(datasource_id),
	}

	output, err := conn.GetDataSource(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func flattenDocumentEnrichmentConfiguration(cfg *types.DocumentEnrichmentConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if cfg.InlineConfigurations != nil {
		m["inline_configurations"] = flattenInlineDocumentEnrichmentConfigurations(cfg.InlineConfigurations)
	}

	if cfg.PostExtractionHookConfiguration != nil {
		m["post_extraction_hook_configuration"] = flattenHookConfiguration(cfg.PostExtractionHookConfiguration)
	}

	if cfg.PreExtractionHookConfiguration != nil {
		m["pre_extraction_hook_configuration"] = flattenHookConfiguration(cfg.PreExtractionHookConfiguration)
	}

	return []interface{}{m}
}

func flattenInlineDocumentEnrichmentConfigurations(cfg []types.InlineDocumentEnrichmentConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	var conf []interface{}

	for _, c := range cfg {
		conf = append(conf, map[string]interface{}{
			"condition":                 flattenDocumentAttributeCondition(c.Condition),
			"document_content_operator": string(c.DocumentContentOperator),
			"target":                    flattenDocumentAttributeTarget(c.Target),
		})
	}

	m["configuration"] = conf

	return []interface{}{m}
}

func expandDocumentEnrichmentConfiguration(v []interface{}) *types.DocumentEnrichmentConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &types.DocumentEnrichmentConfiguration{
		InlineConfigurations:            expandInlineDocumentEnrichmentConfigurations(m["inline_configurations"].([]interface{})),
		PostExtractionHookConfiguration: expandHookConfiguration(m["post_extraction_hook_configuration"].([]interface{})),
		PreExtractionHookConfiguration:  expandHookConfiguration(m["pre_extraction_hook_configuration"].([]interface{})),
	}
}

func expandInlineDocumentEnrichmentConfigurations(v []interface{}) []types.InlineDocumentEnrichmentConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	var conf []types.InlineDocumentEnrichmentConfiguration

	for _, c := range m["configuration"].([]interface{}) {
		conf = append(conf, types.InlineDocumentEnrichmentConfiguration{
			Condition:               expandDocumentAttributeCondition(c.(map[string]interface{})["condition"].([]interface{})),
			DocumentContentOperator: types.DocumentContentOperator(c.(map[string]interface{})["document_content_operator"].(string)),
			Target:                  expandDocumentAttributeTarget(c.(map[string]interface{})["target"].([]interface{})),
		})
	}
	return conf
}

func flattenHookConfiguration(cfg *types.HookConfiguration) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["invocation_condition"] = flattenDocumentAttributeCondition(cfg.InvocationCondition)
	m["role_arn"] = aws.ToString(cfg.RoleArn)
	m["lambda_arn"] = aws.ToString(cfg.LambdaArn)
	m["s3_bucket_name"] = aws.ToString(cfg.S3BucketName)

	return []interface{}{m}
}

func expandHookConfiguration(v []interface{}) *types.HookConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &types.HookConfiguration{
		InvocationCondition: expandDocumentAttributeCondition(m["invocation_condition"].([]interface{})),
		RoleArn:             aws.String(m["role_arn"].(string)),
		LambdaArn:           aws.String(m["lambda_arn"].(string)),
		S3BucketName:        aws.String(m["s3_bucket_name"].(string)),
	}
}

func flattenDocumentAttributeCondition(cfg *types.DocumentAttributeCondition) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["key"] = aws.ToString(cfg.Key)
	m["operator"] = string(cfg.Operator)
	m["value"] = flattenValueSchema(cfg.Value)

	return []interface{}{m}
}

func expandDocumentAttributeTarget(v []interface{}) *types.DocumentAttributeTarget {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &types.DocumentAttributeTarget{
		Key:                    aws.String(m["key"].(string)),
		AttributeValueOperator: types.AttributeValueOperator(m["attribute_value_operator"].(string)),
		Value:                  expandValueSchema(m["value"].([]interface{})),
	}
}

func flattenDocumentAttributeTarget(cfg *types.DocumentAttributeTarget) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["key"] = aws.ToString(cfg.Key)
	m["attribute_value_operator"] = string(cfg.AttributeValueOperator)
	m["value"] = flattenValueSchema(cfg.Value)

	return []interface{}{m}
}

func expandDocumentAttributeCondition(v []interface{}) *types.DocumentAttributeCondition {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &types.DocumentAttributeCondition{
		Key:      aws.String(m["key"].(string)),
		Operator: types.DocumentEnrichmentConditionOperator(m["operator"].(string)),
		Value:    expandValueSchema(m["value"].([]interface{})),
	}
}

func flattenValueSchema(v types.DocumentAttributeValue) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	switch v := v.(type) {
	case *types.DocumentAttributeValueMemberDateValue:
		m["date_value"] = v.Value.Format(time.RFC3339)
	case *types.DocumentAttributeValueMemberLongValue:
		m["long_value"] = v.Value
	case *types.DocumentAttributeValueMemberStringListValue:
		m["string_list_value"] = flex.FlattenStringValueList(v.Value)
	case *types.DocumentAttributeValueMemberStringValue:
		m["string_value"] = v.Value
	}

	return []interface{}{m}
}

func expandValueSchema(v []interface{}) types.DocumentAttributeValue {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	if date_value, ok := m["date_value"].(string); ok {
		if t, err := time.Parse(time.RFC3339, date_value); err == nil {
			return &types.DocumentAttributeValueMemberDateValue{
				Value: t,
			}
		}
	}

	if string_list_value, ok := m["string_list_value"].([]interface{}); ok && len(string_list_value) > 0 {
		return &types.DocumentAttributeValueMemberStringListValue{
			Value: flex.ExpandStringValueList(string_list_value),
		}
	}

	if string_value, ok := m["string_value"].(string); ok && string_value != "" {
		return &types.DocumentAttributeValueMemberStringValue{
			Value: string_value,
		}
	}

	if long_value, ok := m["long_value"].(int); ok {
		return &types.DocumentAttributeValueMemberLongValue{
			Value: int64(long_value),
		}
	}

	return nil
}

func expandVPCConfiguration(cfg []interface{}) *types.DataSourceVpcConfiguration {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := &types.DataSourceVpcConfiguration{}

	if v, ok := conf["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := conf["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		out.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return out
}

func flattenVPCConfiguration(rs *types.DataSourceVpcConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.SecurityGroupIds != nil {
		m["security_group_ids"] = flex.FlattenStringValueSet(rs.SecurityGroupIds)
	}
	if rs.SubnetIds != nil {
		m["subnet_ids"] = flex.FlattenStringValueSet(rs.SubnetIds)
	}

	return []interface{}{m}
}
