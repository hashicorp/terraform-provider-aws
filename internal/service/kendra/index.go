package kendra

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kendra"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIndex() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIndexCreate,
		ReadContext:   resourceIndexRead,
		UpdateContext: resourceIndexUpdate,
		DeleteContext: resourceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(kendraIndexCreatedTimeout),
			Update: schema.DefaultTimeout(kendraIndexUpdatedTimeout),
			Delete: schema.DefaultTimeout(kendraIndexDeletedTimeout),
		},
		CustomizeDiff: verify.SetTagsDiff,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"storage_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"document_metadata_configuration_updates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"relevance": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"duration": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"freshness": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"importance": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"rank_order": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values_importance_map": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"search": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"displayable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"facetable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"searchable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"sortable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"edition": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(kendra.IndexEdition_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// API returns "ENTERPRISE_EDITION" by default.
					if old == kendra.IndexEditionEnterpriseEdition && new == "" {
						return true
					}

					return old == new
				},
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_statistics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"faq_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_question_answers_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"text_document_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_text_bytes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"indexed_text_documents_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(
						regexp.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`),
						"The name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_context_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(kendra.UserContextPolicy_Values(), false),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// API returns "ATTRIBUTE_FILTER" by default.
					if old == kendra.UserContextPolicyAttributeFilter && new == "" {
						return true
					}

					return old == new
				},
			},
			"user_group_resolution_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_group_resolution_mode": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(kendra.UserGroupResolutionMode_Values(), false),
						},
					},
				},
			},
			"user_token_configurations": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_token_type_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_attribute_field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 2048),
									},
									"user_name_attribute_field": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 2048),
									},
								},
							},
						},
						"jwt_token_type_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"claim_regex": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 100),
									},
									"group_attribute_field": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 100),
									},
									"issuer": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 65),
									},
									"key_location": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(kendra.KeyLocation_Values(), false),
									},
									"secrets_manager_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"url": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 2048),
											validation.StringMatch(
												regexp.MustCompile(`^(https?|ftp|file):\/\/([^\s]*)`),
												"Must be valid URL",
											),
										),
									},
									"user_name_attribute_field": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &kendra.CreateIndexInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(name),
		RoleArn:     aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("edition"); ok {
		input.Edition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_side_encryption_configuration"); ok {
		input.ServerSideEncryptionConfiguration = expandServerSideEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_context_policy"); ok {
		input.UserContextPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_group_resolution_configuration"); ok {
		input.UserGroupResolutionConfiguration = expandUserGroupResolutionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_token_configurations"); ok {
		input.UserTokenConfigurations = expandUserTokenConfigurations(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Kendra Index %s", input)
	output, err := conn.CreateIndexWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Kendra Index (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Kendra Index (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.Id))

	// waiter since the status changes from CREATING to either ACTIVE or FAILED
	if _, err := waitIndexCreated(ctx, conn, d.Timeout(schema.TimeoutCreate), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Index (%s) creation: %w", d.Id(), err))
	}

	return resourceIndexRead(ctx, d, meta)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Id()

	resp, err := conn.DescribeIndexWithContext(ctx, &kendra.DescribeIndexInput{
		Id: aws.String(id),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kendra.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Kendra Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Kendra Index (%s): %w", d.Id(), err))
	}

	if resp == nil {
		return diag.FromErr(fmt.Errorf("error getting Kendra Index (%s): empty response", d.Id()))
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("edition", resp.Edition)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("status", resp.Status)
	d.Set("updated_at", aws.TimeValue(resp.UpdatedAt).Format(time.RFC3339))
	d.Set("user_context_policy", resp.UserContextPolicy)

	if err := d.Set("capacity_units", flattenCapacityUnits(resp.CapacityUnits)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("document_metadata_configuration_updates", flattenDocumentMetadataConfigurations(resp.DocumentMetadataConfigurations)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("index_statistics", flattenIndexStatistics(resp.IndexStatistics)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(resp.ServerSideEncryptionConfiguration)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_group_resolution_configuration", flattenUserGroupResolutionConfiguration(resp.UserGroupResolutionConfiguration)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_token_configurations", flattenUserTokenConfigurations(resp.UserTokenConfigurations)); err != nil {
		return diag.FromErr(err)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing tags for resource (%s): %s", d.Get("arn").(string), err))
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id := d.Id()

	if d.HasChanges("description", "name", "role_arn", "user_context_policy", "user_group_resolution_configuration", "user_token_configurations") {
		input := &kendra.UpdateIndexInput{
			Id: aws.String(id),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}
		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}
		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}
		if d.HasChange("user_context_policy") {
			input.UserContextPolicy = aws.String(d.Get("user_context_policy").(string))
		}
		if d.HasChange("user_group_resolution_configuration") {
			input.UserGroupResolutionConfiguration = expandUserGroupResolutionConfiguration(d.Get("user_group_resolution_configuration").([]interface{}))
		}
		if d.HasChange("user_token_configurations") {
			input.UserTokenConfigurations = expandUserTokenConfigurations(d.Get("user_token_configurations").([]interface{}))
		}

		_, err := conn.UpdateIndexWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("[ERROR] Error updating Index (%s): %w", d.Id(), err))
		}

		// waiter since the status changes from UPDATING to either ACTIVE or FAILED
		if _, err := waitIndexUpdated(ctx, conn, d.Timeout(schema.TimeoutUpdate), d.Id()); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for Index (%s) update: %w", d.Id(), err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	return resourceIndexRead(ctx, d, meta)
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id := d.Id()

	_, err := conn.DeleteIndexWithContext(ctx, &kendra.DeleteIndexInput{
		Id: aws.String(id),
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Index (%s): %w", d.Id(), err))
	}

	if _, err := waitIndexDeleted(ctx, conn, d.Timeout(schema.TimeoutDelete), id); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Index (%s) delete: %w", d.Id(), err))
	}

	return nil
}

func expandServerSideEncryptionConfiguration(serverSideEncryptionConfiguration []interface{}) *kendra.ServerSideEncryptionConfiguration {
	if len(serverSideEncryptionConfiguration) == 0 || serverSideEncryptionConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := serverSideEncryptionConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &kendra.ServerSideEncryptionConfiguration{}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		result.KmsKeyId = aws.String(v)
	}

	return result
}

func expandUserGroupResolutionConfiguration(userGroupResolutionConfiguration []interface{}) *kendra.UserGroupResolutionConfiguration {
	if len(userGroupResolutionConfiguration) == 0 || userGroupResolutionConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := userGroupResolutionConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &kendra.UserGroupResolutionConfiguration{
		UserGroupResolutionMode: aws.String(tfMap["user_group_resolution_mode"].(string)),
	}

	return result
}

func expandUserTokenConfigurations(userTokenConfigurations []interface{}) []*kendra.UserTokenConfiguration {
	if len(userTokenConfigurations) == 0 {
		return nil
	}

	userTokenConfigurationsConfigs := []*kendra.UserTokenConfiguration{}

	for _, userTokenConfiguration := range userTokenConfigurations {
		tfMap := userTokenConfiguration.(map[string]interface{})
		userTokenConfigurationConfig := &kendra.UserTokenConfiguration{}

		if v, ok := tfMap["json_token_type_configuration"].([]interface{}); ok && len(v) > 0 {
			userTokenConfigurationConfig.JsonTokenTypeConfiguration = expandJsonTokenTypeConfiguration(v)
		}

		if v, ok := tfMap["jwt_token_type_configuration"].([]interface{}); ok && len(v) > 0 {
			userTokenConfigurationConfig.JwtTokenTypeConfiguration = expandJwtTokenTypeConfiguration(v)
		}

		userTokenConfigurationsConfigs = append(userTokenConfigurationsConfigs, userTokenConfigurationConfig)
	}

	return userTokenConfigurationsConfigs
}

func expandJsonTokenTypeConfiguration(jsonTokenTypeConfiguration []interface{}) *kendra.JsonTokenTypeConfiguration {
	if len(jsonTokenTypeConfiguration) == 0 || jsonTokenTypeConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := jsonTokenTypeConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &kendra.JsonTokenTypeConfiguration{
		GroupAttributeField:    aws.String(tfMap["group_attribute_field"].(string)),
		UserNameAttributeField: aws.String(tfMap["user_name_attribute_field"].(string)),
	}

	return result
}

func expandJwtTokenTypeConfiguration(jwtTokenTypeConfiguration []interface{}) *kendra.JwtTokenTypeConfiguration {
	if len(jwtTokenTypeConfiguration) == 0 || jwtTokenTypeConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := jwtTokenTypeConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &kendra.JwtTokenTypeConfiguration{
		KeyLocation: aws.String(tfMap["key_location"].(string)),
	}

	if v, ok := tfMap["claim_regex"].(string); ok && v != "" {
		result.ClaimRegex = aws.String(v)
	}

	if v, ok := tfMap["group_attribute_field"].(string); ok && v != "" {
		result.GroupAttributeField = aws.String(v)
	}

	if v, ok := tfMap["issuer"].(string); ok && v != "" {
		result.Issuer = aws.String(v)
	}

	if v, ok := tfMap["secrets_manager_arn"].(string); ok && v != "" {
		result.SecretManagerArn = aws.String(v)
	}

	if v, ok := tfMap["url"].(string); ok && v != "" {
		result.URL = aws.String(v)
	}

	if v, ok := tfMap["user_name_attribute_field"].(string); ok && v != "" {
		result.UserNameAttributeField = aws.String(v)
	}

	return result
}

func flattenCapacityUnits(capacityUnits *kendra.CapacityUnitsConfiguration) []interface{} {
	if capacityUnits == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"query_capacity_units":   aws.Int64Value(capacityUnits.QueryCapacityUnits),
		"storage_capacity_units": aws.Int64Value(capacityUnits.StorageCapacityUnits),
	}

	return []interface{}{values}
}

func flattenDocumentMetadataConfigurations(documentMetadataConfigurations []*kendra.DocumentMetadataConfiguration) []interface{} {
	documentMetadataConfigurationsList := []interface{}{}

	for _, documentMetadataConfiguration := range documentMetadataConfigurations {
		values := map[string]interface{}{
			"name":      aws.StringValue(documentMetadataConfiguration.Name),
			"relevance": flattenRelevance(documentMetadataConfiguration.Relevance),
			"search":    flattenSearch(documentMetadataConfiguration.Search),
			"type":      aws.StringValue(documentMetadataConfiguration.Type),
		}

		documentMetadataConfigurationsList = append(documentMetadataConfigurationsList, values)
	}

	return documentMetadataConfigurationsList
}

func flattenRelevance(relevance *kendra.Relevance) []interface{} {
	if relevance == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := relevance.Duration; v != nil {
		values["duration"] = aws.StringValue(v)
	}

	if v := relevance.Freshness; v != nil {
		values["freshness"] = aws.BoolValue(v)
	}

	if v := relevance.Importance; v != nil {
		values["importance"] = aws.Int64Value(v)
	}

	if v := relevance.RankOrder; v != nil {
		values["rank_order"] = aws.StringValue(v)
	}

	if v := relevance.ValueImportanceMap; v != nil {
		values["values_importance_map"] = aws.Int64ValueMap(v)
	}

	return []interface{}{values}
}

func flattenSearch(search *kendra.Search) []interface{} {
	if search == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := search.Displayable; v != nil {
		values["displayable"] = aws.BoolValue(v)
	}

	if v := search.Facetable; v != nil {
		values["facetable"] = aws.BoolValue(v)
	}

	if v := search.Searchable; v != nil {
		values["searchable"] = aws.BoolValue(v)
	}

	if v := search.Sortable; v != nil {
		values["sortable"] = aws.BoolValue(v)
	}

	return []interface{}{values}
}

func flattenIndexStatistics(indexStatistics *kendra.IndexStatistics) []interface{} {
	if indexStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"faq_statistics":           flattenFaqStatistics(indexStatistics.FaqStatistics),
		"text_document_statistics": flattenTextDocumentStatistics(indexStatistics.TextDocumentStatistics),
	}

	return []interface{}{values}
}

func flattenFaqStatistics(faqStatistics *kendra.FaqStatistics) []interface{} {
	if faqStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"indexed_question_answers_count": aws.Int64Value(faqStatistics.IndexedQuestionAnswersCount),
	}

	return []interface{}{values}
}

func flattenTextDocumentStatistics(textDocumentStatistics *kendra.TextDocumentStatistics) []interface{} {
	if textDocumentStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"indexed_text_bytes":           aws.Int64Value(textDocumentStatistics.IndexedTextBytes),
		"indexed_text_documents_count": aws.Int64Value(textDocumentStatistics.IndexedTextDocumentsCount),
	}

	return []interface{}{values}
}

func flattenServerSideEncryptionConfiguration(serverSideEncryptionConfiguration *kendra.ServerSideEncryptionConfiguration) []interface{} {
	if serverSideEncryptionConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := serverSideEncryptionConfiguration.KmsKeyId; v != nil {
		values["kms_key_id"] = aws.StringValue(v)
	}

	return []interface{}{values}
}

func flattenUserGroupResolutionConfiguration(userGroupResolutionConfiguration *kendra.UserGroupResolutionConfiguration) []interface{} {
	if userGroupResolutionConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"user_group_resolution_configuration": aws.StringValue(userGroupResolutionConfiguration.UserGroupResolutionMode),
	}

	return []interface{}{values}
}

func flattenUserTokenConfigurations(userTokenConfigurations []*kendra.UserTokenConfiguration) []interface{} {
	userTokenConfigurationsList := []interface{}{}

	for _, userTokenConfiguration := range userTokenConfigurations {
		values := map[string]interface{}{}

		if v := userTokenConfiguration.JsonTokenTypeConfiguration; v != nil {
			values["json_token_type_configuration"] = flattenJsonTokenTypeConfiguration(v)
		}

		if v := userTokenConfiguration.JwtTokenTypeConfiguration; v != nil {
			values["jwt_token_type_configuration"] = flattenJwtTokenTypeConfiguration(v)
		}

		userTokenConfigurationsList = append(userTokenConfigurationsList, values)
	}

	return userTokenConfigurationsList
}

func flattenJsonTokenTypeConfiguration(jsonTokenTypeConfiguration *kendra.JsonTokenTypeConfiguration) []interface{} {
	if jsonTokenTypeConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"group_attribute_field":     aws.StringValue(jsonTokenTypeConfiguration.GroupAttributeField),
		"user_name_attribute_field": aws.StringValue(jsonTokenTypeConfiguration.UserNameAttributeField),
	}

	return []interface{}{values}
}

func flattenJwtTokenTypeConfiguration(jwtTokenTypeConfiguration *kendra.JwtTokenTypeConfiguration) []interface{} {
	if jwtTokenTypeConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"key_location": aws.StringValue(jwtTokenTypeConfiguration.KeyLocation),
	}

	if v := jwtTokenTypeConfiguration.ClaimRegex; v != nil {
		values["claim_regex"] = aws.StringValue(v)
	}

	if v := jwtTokenTypeConfiguration.GroupAttributeField; v != nil {
		values["group_attribute_field"] = aws.StringValue(v)
	}

	if v := jwtTokenTypeConfiguration.Issuer; v != nil {
		values["issuer"] = aws.StringValue(v)
	}

	if v := jwtTokenTypeConfiguration.SecretManagerArn; v != nil {
		values["secrets_manager_arn"] = aws.StringValue(v)
	}

	if v := jwtTokenTypeConfiguration.URL; v != nil {
		values["url"] = aws.StringValue(v)
	}

	if v := jwtTokenTypeConfiguration.UserNameAttributeField; v != nil {
		values["user_name_attribute_field"] = aws.StringValue(v)
	}

	return []interface{}{values}
}
