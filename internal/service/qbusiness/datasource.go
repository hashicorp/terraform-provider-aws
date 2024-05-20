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
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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

func valueSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"date_value": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "A date expressed as an ISO 8601 string.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName(names.AttrName),
						path.MatchRelative().AtParent().AtName("long_value"),
						path.MatchRelative().AtParent().AtName("string_list_value"),
						path.MatchRelative().AtParent().AtName("string_value"),
					),
				},
			},
			"long_value": schema.Int64Attribute{
				Description: "A long integer value.",
				Optional:    true,
			},
			"string_list_value": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Description: "A list of string values.",
				Optional:    true,
			},
			"string_value": schema.StringAttribute{
				Description: "A string value.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
				},
			},
		},
	}
}

func conditionSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			names.AttrKey: schema.StringAttribute{
				Description: "The identifier of the document attribute used for the condition.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`), "must be a valid document attribute"),
				},
			},
			"operator": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.DocumentEnrichmentConditionOperator](),
				Required:    true,
				Description: "Operator of the document attribute used for the condition.",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.DocumentEnrichmentConditionOperator](),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"value": valueSchema(),
		},
	}
}

func hookConfigurationSchema() schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"lambda_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of a role with permission to run a Lambda function during ingestion.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile(`^arn:aws[a-zA-Z-]*:lambda:[a-z-]*-[0-9]:[0-9]{12}:function:[a-zA-Z0-9-_]+(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?(:[a-zA-Z0-9-_]+)?$`), "must be a valid Lambda ARN"),
				},
			},
			"role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of a role with permission to run PreExtractionHookConfiguration and PostExtractionHookConfiguration for altering document metadata and content during the document ingestion process.",
				Optional:    true,
			},
			"s3_bucket_name": schema.StringAttribute{
				Description: "Stores the original, raw documents or the structured, parsed documents before and after altering them.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9]$`), "must be a valid bucket name"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"invocation_condition": conditionSchema(),
		},
	}
}

func (r *resourceDatasource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_datasource"
}

func (r *resourceDatasource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Description: "A description of the Amazon Q datasource.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrConfiguration: schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Description: "Configuration information (JSON) to connect to your data source repository.",
				Required:    true,
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the datasource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			"index_id": schema.StringAttribute{
				Description: "The identifier of the index that you want to use with the data source connector.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid index ID"),
				},
			},
			"iam_service_role_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of an IAM role with permission to access the data source and required resources.",
				Optional:    true,
			},
			"sync_schedule": schema.StringAttribute{
				Description: "Frequency for Amazon Q to check the documents in your data source repository and update your index.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 998),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrVPCConfig: schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description:  "Information for an VPC to connect to your data source.",
				NestedObject: schema.NestedBlockObject{},
			},
			"document_enrichment_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description:  "Configuration information for altering document metadata and content during the document ingestion process.",
				NestedObject: schema.NestedBlockObject{},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
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

func keySchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 200),
			validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`),
				"must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters")),
	}
}

func valueSchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &sdkschema.Resource{
			Schema: map[string]*sdkschema.Schema{
				"date_value": {
					Type:         sdkschema.TypeString,
					Optional:     true,
					ValidateFunc: validation.IsRFC3339Time,
				},
				"long_value": {
					Type:     sdkschema.TypeInt,
					Optional: true,
				},
				"string_list_value": {
					Type:     sdkschema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 2048,
					Elem:     &sdkschema.Schema{Type: sdkschema.TypeString},
				},
				"string_value": {
					Type:     sdkschema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
					),
				},
			},
		},
	}
}

func documentAttributeConditionSchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &sdkschema.Resource{
			Schema: map[string]*sdkschema.Schema{
				"key": keySchema(),
				"operator": {
					Type:             sdkschema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.DocumentEnrichmentConditionOperator](),
				},
				"value": valueSchema(),
			},
		},
	}
}

func hookConfigurationSchema() *sdkschema.Schema {
	return &sdkschema.Schema{
		Type:     sdkschema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &sdkschema.Resource{Schema: map[string]*sdkschema.Schema{
			"invocation_condition": documentAttributeConditionSchema(),
			"lambda_arn": {
				Type:         sdkschema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"role_arn": {
				Type:         sdkschema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_bucket_name": {
				Type:     sdkschema.TypeString,
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
func resourceDatasourceCreate(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if _, err := waitDatasourceCreated(ctx, conn, d.Id(), d.Timeout(sdkschema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness datasource (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceDatasourceRead(ctx, d, meta)...)
}

func resourceDatasourceRead(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceDatasourceUpdate(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if v, ok := d.GetOk("document_enrichment_configuration"); ok {
		input.DocumentEnrichmentConfiguration = expandDocumentEnrichmentConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("configuration"); ok {
		var conf map[string]interface{}
		err = json.Unmarshal([]byte(v.(string)), &conf)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "unmarshal qbusiness datasource configuration: %s", err)
		}

		input.Configuration = document.NewLazyDocument(conf)
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("iam_service_role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
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

	_, err = conn.UpdateDataSource(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating qbusiness datasource (%s): %s", d.Id(), err)
	}

	if _, err := waitDatasourceUpdated(ctx, conn, d.Id(), d.Timeout(sdkschema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness datasource (%s) to be updated: %s", d.Id(), err)
	}

	return append(diags, resourceDatasourceRead(ctx, d, meta)...)
}

func resourceDatasourceDelete(ctx context.Context, d *sdkschema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting qbusiness datasource (%s): %s", d.Id(), err)
	}

	if _, err := waitDatasourceDeleted(ctx, conn, d.Id(), d.Timeout(sdkschema.TimeoutDelete)); err != nil {
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

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func flattenDocumentEnrichmentConfiguration(cfg *awstypes.DocumentEnrichmentConfiguration) []interface{} {
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

func flattenInlineDocumentEnrichmentConfigurations(cfg []awstypes.InlineDocumentEnrichmentConfiguration) []interface{} {
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

func expandDocumentEnrichmentConfiguration(v []interface{}) *awstypes.DocumentEnrichmentConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &awstypes.DocumentEnrichmentConfiguration{
		InlineConfigurations:            expandInlineDocumentEnrichmentConfigurations(m["inline_configurations"].([]interface{})),
		PostExtractionHookConfiguration: expandHookConfiguration(m["post_extraction_hook_configuration"].([]interface{})),
		PreExtractionHookConfiguration:  expandHookConfiguration(m["pre_extraction_hook_configuration"].([]interface{})),
	}
}

func expandInlineDocumentEnrichmentConfigurations(v []interface{}) []awstypes.InlineDocumentEnrichmentConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	var conf []awstypes.InlineDocumentEnrichmentConfiguration

	for _, c := range m["configuration"].([]interface{}) {
		conf = append(conf, awstypes.InlineDocumentEnrichmentConfiguration{
			Condition:               expandDocumentAttributeCondition(c.(map[string]interface{})["condition"].([]interface{})),
			DocumentContentOperator: awstypes.DocumentContentOperator(c.(map[string]interface{})["document_content_operator"].(string)),
			Target:                  expandDocumentAttributeTarget(c.(map[string]interface{})["target"].([]interface{})),
		})
	}
	return conf
}

func flattenHookConfiguration(cfg *awstypes.HookConfiguration) []interface{} {
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

func expandHookConfiguration(v []interface{}) *awstypes.HookConfiguration {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &awstypes.HookConfiguration{
		InvocationCondition: expandDocumentAttributeCondition(m["invocation_condition"].([]interface{})),
		RoleArn:             aws.String(m["role_arn"].(string)),
		LambdaArn:           aws.String(m["lambda_arn"].(string)),
		S3BucketName:        aws.String(m["s3_bucket_name"].(string)),
	}
}

func flattenDocumentAttributeCondition(cfg *awstypes.DocumentAttributeCondition) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["key"] = aws.ToString(cfg.Key)
	m["operator"] = string(cfg.Operator)
	m["value"] = flattenValueSchema(cfg.Value)

	return []interface{}{m}
}

func expandDocumentAttributeTarget(v []interface{}) *awstypes.DocumentAttributeTarget {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &awstypes.DocumentAttributeTarget{
		Key:                    aws.String(m["key"].(string)),
		AttributeValueOperator: awstypes.AttributeValueOperator(m["attribute_value_operator"].(string)),
		Value:                  expandValueSchema(m["value"].([]interface{})),
	}
}

func flattenDocumentAttributeTarget(cfg *awstypes.DocumentAttributeTarget) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	m["key"] = aws.ToString(cfg.Key)
	m["attribute_value_operator"] = string(cfg.AttributeValueOperator)
	m["value"] = flattenValueSchema(cfg.Value)

	return []interface{}{m}
}

func expandDocumentAttributeCondition(v []interface{}) *awstypes.DocumentAttributeCondition {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	return &awstypes.DocumentAttributeCondition{
		Key:      aws.String(m["key"].(string)),
		Operator: awstypes.DocumentEnrichmentConditionOperator(m["operator"].(string)),
		Value:    expandValueSchema(m["value"].([]interface{})),
	}
}

func flattenValueSchema(v awstypes.DocumentAttributeValue) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	switch v := v.(type) {
	case *awstypes.DocumentAttributeValueMemberDateValue:
		m["date_value"] = v.Value.Format(time.RFC3339)
	case *awstypes.DocumentAttributeValueMemberLongValue:
		m["long_value"] = v.Value
	case *awstypes.DocumentAttributeValueMemberStringListValue:
		m["string_list_value"] = flex.FlattenStringValueList(v.Value)
	case *awstypes.DocumentAttributeValueMemberStringValue:
		m["string_value"] = v.Value
	}

	return []interface{}{m}
}

func expandValueSchema(v []interface{}) awstypes.DocumentAttributeValue {
	if len(v) == 0 || v[0] == nil {
		return nil
	}
	m := v[0].(map[string]interface{})

	if date_value, ok := m["date_value"].(string); ok {
		if t, err := time.Parse(time.RFC3339, date_value); err == nil {
			return &awstypes.DocumentAttributeValueMemberDateValue{
				Value: t,
			}
		}
	}

	if string_list_value, ok := m["string_list_value"].([]interface{}); ok && len(string_list_value) > 0 {
		return &awstypes.DocumentAttributeValueMemberStringListValue{
			Value: flex.ExpandStringValueList(string_list_value),
		}
	}

	if string_value, ok := m["string_value"].(string); ok && string_value != "" {
		return &awstypes.DocumentAttributeValueMemberStringValue{
			Value: string_value,
		}
	}

	if long_value, ok := m["long_value"].(int); ok {
		return &awstypes.DocumentAttributeValueMemberLongValue{
			Value: int64(long_value),
		}
	}

	return nil
}

func expandVPCConfiguration(cfg []interface{}) *awstypes.DataSourceVpcConfiguration {
	if len(cfg) == 0 || cfg[0] == nil {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := &awstypes.DataSourceVpcConfiguration{}

	if v, ok := conf["security_group_ids"].(*sdkschema.Set); ok && v.Len() > 0 {
		out.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := conf["subnet_ids"].(*sdkschema.Set); ok && v.Len() > 0 {
		out.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return out
}

func flattenVPCConfiguration(rs *awstypes.DataSourceVpcConfiguration) []interface{} {
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
