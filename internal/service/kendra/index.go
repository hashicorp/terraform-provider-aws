package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Allow IAM role to become visible to the index
	propagationTimeout = 2 * time.Minute

	// validationExceptionMessage describes the error returned when the IAM role has not yet propagated
	validationExceptionMessage = "Please make sure your role exists and has `kendra.amazonaws.com` as trusted entity"
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
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
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
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_capacity_units": {
							Type:         schema.TypeInt,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"storage_capacity_units": {
							Type:         schema.TypeInt,
							Computed:     true,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
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
				Default:      string(types.IndexEditionEnterpriseEdition),
				ValidateFunc: validation.StringInSlice(indexEditionValues(types.IndexEdition("").Values()...), false),
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
				Default:      string(types.UserContextPolicyAttributeFilter),
				ValidateFunc: validation.StringInSlice(userContextPolicyValues(types.UserContextPolicy("").Values()...), false),
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
							ValidateFunc: validation.StringInSlice(userGroupResolutionModeValues(types.UserGroupResolutionMode("").Values()...), false),
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
										ValidateFunc: validation.StringInSlice(keyLocationValues(types.KeyLocation("").Values()...), false),
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
		input.Edition = types.IndexEdition(v.(string))
	}

	if v, ok := d.GetOk("server_side_encryption_configuration"); ok {
		input.ServerSideEncryptionConfiguration = expandServerSideEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("user_context_policy"); ok {
		input.UserContextPolicy = types.UserContextPolicy(v.(string))
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

	log.Printf("[DEBUG] Creating Kendra Index %#v", input)

	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateIndex(ctx, input)
		},
		func(err error) (bool, error) {
			var validationException *types.ValidationException

			if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return diag.Errorf("error creating Kendra Index (%s): %s", name, err)
	}

	if outputRaw == nil {
		return diag.Errorf("error creating Kendra Index (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateIndexOutput)

	d.SetId(aws.ToString(output.Id))

	// waiter since the status changes from CREATING to either ACTIVE or FAILED
	if _, err := waitIndexCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("error waiting for Index (%s) creation: %s", d.Id(), err)
	}

	// CreateIndex API does not support capacity_units but UpdateIndex does
	if v, ok := d.GetOk("capacity_units"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		return resourceIndexUpdate(ctx, d, meta)
	}

	return resourceIndexRead(ctx, d, meta)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := findIndexByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting Kendra Index (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("edition", resp.Edition)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("status", resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))
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

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return diag.Errorf("error listing tags for resource (%s): %s", d.Get("arn").(string), err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id := d.Id()

	if d.HasChanges("capacity_units", "description", "name", "role_arn", "user_context_policy", "user_group_resolution_configuration", "user_token_configurations") {
		input := &kendra.UpdateIndexInput{
			Id: aws.String(id),
		}
		if d.HasChange("capacity_units") {
			input.CapacityUnits = expandCapacityUnits(d.Get("capacity_units").([]interface{}))
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
			input.UserContextPolicy = types.UserContextPolicy(d.Get("user_context_policy").(string))
		}
		if d.HasChange("user_group_resolution_configuration") {
			input.UserGroupResolutionConfiguration = expandUserGroupResolutionConfiguration(d.Get("user_group_resolution_configuration").([]interface{}))
		}
		if d.HasChange("user_token_configurations") {
			input.UserTokenConfigurations = expandUserTokenConfigurations(d.Get("user_token_configurations").([]interface{}))
		}

		_, err := tfresource.RetryWhen(
			propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateIndex(ctx, input)
			},
			func(err error) (bool, error) {
				var validationException *types.ValidationException

				if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return diag.Errorf("error updating Index (%s): %s", d.Id(), err)
		}

		// waiter since the status changes from UPDATING to either ACTIVE or FAILED
		if _, err := waitIndexUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("error waiting for Index (%s) update: %s", d.Id(), err)
		}
	}

	if !d.IsNewResource() && d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceIndexRead(ctx, d, meta)
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KendraConn

	id := d.Id()

	_, err := conn.DeleteIndex(ctx, &kendra.DeleteIndexInput{
		Id: aws.String(id),
	})

	if err != nil {
		return diag.Errorf("error deleting Index (%s): %s", d.Id(), err)
	}

	if _, err := waitIndexDeleted(ctx, conn, id, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for Index (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func findIndexByID(ctx context.Context, conn *kendra.Client, id string) (*kendra.DescribeIndexOutput, error) {
	input := &kendra.DescribeIndexInput{
		Id: aws.String(id),
	}

	output, err := conn.DescribeIndex(ctx, input)

	if err != nil {
		var resourceNotFoundException *types.ResourceNotFoundException

		if errors.As(err, &resourceNotFoundException) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusIndex(ctx context.Context, conn *kendra.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findIndexByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitIndexCreated(ctx context.Context, conn *kendra.Client, id string, timeout time.Duration) (*kendra.DescribeIndexOutput, error) {

	stateConf := &resource.StateChangeConf{
		Pending: IndexStatusValues(types.IndexStatusCreating),
		Target:  IndexStatusValues(types.IndexStatusActive),
		Timeout: timeout,
		Refresh: statusIndex(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		if output.Status == types.IndexStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitIndexUpdated(ctx context.Context, conn *kendra.Client, id string, timeout time.Duration) (*kendra.DescribeIndexOutput, error) {

	stateConf := &resource.StateChangeConf{
		Pending: IndexStatusValues(types.IndexStatusUpdating),
		Target:  IndexStatusValues(types.IndexStatusActive),
		Timeout: timeout,
		Refresh: statusIndex(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		if output.Status == types.IndexStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitIndexDeleted(ctx context.Context, conn *kendra.Client, id string, timeout time.Duration) (*kendra.DescribeIndexOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: IndexStatusValues(types.IndexStatusDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusIndex(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func expandCapacityUnits(capacityUnits []interface{}) *types.CapacityUnitsConfiguration {
	if len(capacityUnits) == 0 || capacityUnits[0] == nil {
		return nil
	}

	tfMap, ok := capacityUnits[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.CapacityUnitsConfiguration{
		QueryCapacityUnits:   aws.Int32(int32(tfMap["query_capacity_units"].(int))),
		StorageCapacityUnits: aws.Int32(int32(tfMap["storage_capacity_units"].(int))),
	}

	return result
}

func expandServerSideEncryptionConfiguration(serverSideEncryptionConfiguration []interface{}) *types.ServerSideEncryptionConfiguration {
	if len(serverSideEncryptionConfiguration) == 0 || serverSideEncryptionConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := serverSideEncryptionConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ServerSideEncryptionConfiguration{}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		result.KmsKeyId = aws.String(v)
	}

	return result
}

func expandUserGroupResolutionConfiguration(userGroupResolutionConfiguration []interface{}) *types.UserGroupResolutionConfiguration {
	if len(userGroupResolutionConfiguration) == 0 || userGroupResolutionConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := userGroupResolutionConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.UserGroupResolutionConfiguration{
		UserGroupResolutionMode: types.UserGroupResolutionMode(tfMap["user_group_resolution_mode"].(string)),
	}

	return result
}

func expandUserTokenConfigurations(userTokenConfigurations []interface{}) []types.UserTokenConfiguration {
	if len(userTokenConfigurations) == 0 {
		return nil
	}

	userTokenConfigurationsConfigs := []types.UserTokenConfiguration{}

	for _, userTokenConfiguration := range userTokenConfigurations {
		tfMap := userTokenConfiguration.(map[string]interface{})
		userTokenConfigurationConfig := types.UserTokenConfiguration{}

		if v, ok := tfMap["json_token_type_configuration"].([]interface{}); ok && len(v) > 0 {
			userTokenConfigurationConfig.JsonTokenTypeConfiguration = expandJSONTokenTypeConfiguration(v)
		}

		if v, ok := tfMap["jwt_token_type_configuration"].([]interface{}); ok && len(v) > 0 {
			userTokenConfigurationConfig.JwtTokenTypeConfiguration = expandJwtTokenTypeConfiguration(v)
		}

		userTokenConfigurationsConfigs = append(userTokenConfigurationsConfigs, userTokenConfigurationConfig)
	}

	return userTokenConfigurationsConfigs
}

func expandJSONTokenTypeConfiguration(jsonTokenTypeConfiguration []interface{}) *types.JsonTokenTypeConfiguration {
	if len(jsonTokenTypeConfiguration) == 0 || jsonTokenTypeConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := jsonTokenTypeConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.JsonTokenTypeConfiguration{
		GroupAttributeField:    aws.String(tfMap["group_attribute_field"].(string)),
		UserNameAttributeField: aws.String(tfMap["user_name_attribute_field"].(string)),
	}

	return result
}

func expandJwtTokenTypeConfiguration(jwtTokenTypeConfiguration []interface{}) *types.JwtTokenTypeConfiguration {
	if len(jwtTokenTypeConfiguration) == 0 || jwtTokenTypeConfiguration[0] == nil {
		return nil
	}

	tfMap, ok := jwtTokenTypeConfiguration[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.JwtTokenTypeConfiguration{
		KeyLocation: types.KeyLocation(tfMap["key_location"].(string)),
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

func flattenCapacityUnits(capacityUnits *types.CapacityUnitsConfiguration) []interface{} {
	if capacityUnits == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"query_capacity_units":   aws.ToInt32(capacityUnits.QueryCapacityUnits),
		"storage_capacity_units": aws.ToInt32(capacityUnits.StorageCapacityUnits),
	}

	return []interface{}{values}
}

func flattenDocumentMetadataConfigurations(documentMetadataConfigurations []types.DocumentMetadataConfiguration) []interface{} {
	documentMetadataConfigurationsList := []interface{}{}

	for _, documentMetadataConfiguration := range documentMetadataConfigurations {
		values := map[string]interface{}{
			"name":      documentMetadataConfiguration.Name,
			"relevance": flattenRelevance(documentMetadataConfiguration.Relevance),
			"search":    flattenSearch(documentMetadataConfiguration.Search),
			"type":      documentMetadataConfiguration.Type,
		}

		documentMetadataConfigurationsList = append(documentMetadataConfigurationsList, values)
	}

	return documentMetadataConfigurationsList
}

func flattenRelevance(relevance *types.Relevance) []interface{} {
	if relevance == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"rank_order": relevance.RankOrder,
	}

	if v := relevance.Duration; v != nil {
		values["duration"] = aws.ToString(v)
	}

	if v := relevance.Freshness; v != nil {
		values["freshness"] = aws.ToBool(v)
	}

	if v := relevance.Importance; v != nil {
		values["importance"] = aws.ToInt32(v)
	}

	if v := relevance.ValueImportanceMap; v != nil {
		values["values_importance_map"] = v
	}

	return []interface{}{values}
}

func flattenSearch(search *types.Search) []interface{} {
	if search == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"displayable": search.Displayable,
		"facetable":   search.Facetable,
		"searchable":  search.Searchable,
		"sortable":    search.Sortable,
	}

	return []interface{}{values}
}

func flattenIndexStatistics(indexStatistics *types.IndexStatistics) []interface{} {
	if indexStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"faq_statistics":           flattenFaqStatistics(indexStatistics.FaqStatistics),
		"text_document_statistics": flattenTextDocumentStatistics(indexStatistics.TextDocumentStatistics),
	}

	return []interface{}{values}
}

func flattenFaqStatistics(faqStatistics *types.FaqStatistics) []interface{} {
	if faqStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"indexed_question_answers_count": aws.ToInt32(&faqStatistics.IndexedQuestionAnswersCount),
	}

	return []interface{}{values}
}

func flattenTextDocumentStatistics(textDocumentStatistics *types.TextDocumentStatistics) []interface{} {
	if textDocumentStatistics == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"indexed_text_bytes":           aws.ToInt64(&textDocumentStatistics.IndexedTextBytes),
		"indexed_text_documents_count": aws.ToInt32(&textDocumentStatistics.IndexedTextDocumentsCount),
	}

	return []interface{}{values}
}

func flattenServerSideEncryptionConfiguration(serverSideEncryptionConfiguration *types.ServerSideEncryptionConfiguration) []interface{} {
	if serverSideEncryptionConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := serverSideEncryptionConfiguration.KmsKeyId; v != nil {
		values["kms_key_id"] = aws.ToString(v)
	}

	return []interface{}{values}
}

func flattenUserGroupResolutionConfiguration(userGroupResolutionConfiguration *types.UserGroupResolutionConfiguration) []interface{} {
	if userGroupResolutionConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"user_group_resolution_configuration": userGroupResolutionConfiguration.UserGroupResolutionMode,
	}

	return []interface{}{values}
}

func flattenUserTokenConfigurations(userTokenConfigurations []types.UserTokenConfiguration) []interface{} {
	userTokenConfigurationsList := []interface{}{}

	for _, userTokenConfiguration := range userTokenConfigurations {
		values := map[string]interface{}{}

		if v := userTokenConfiguration.JsonTokenTypeConfiguration; v != nil {
			values["json_token_type_configuration"] = flattenJSONTokenTypeConfiguration(v)
		}

		if v := userTokenConfiguration.JwtTokenTypeConfiguration; v != nil {
			values["jwt_token_type_configuration"] = flattenJwtTokenTypeConfiguration(v)
		}

		userTokenConfigurationsList = append(userTokenConfigurationsList, values)
	}

	return userTokenConfigurationsList
}

func flattenJSONTokenTypeConfiguration(jsonTokenTypeConfiguration *types.JsonTokenTypeConfiguration) []interface{} {
	if jsonTokenTypeConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"group_attribute_field":     jsonTokenTypeConfiguration.GroupAttributeField,
		"user_name_attribute_field": jsonTokenTypeConfiguration.UserNameAttributeField,
	}

	return []interface{}{values}
}

func flattenJwtTokenTypeConfiguration(jwtTokenTypeConfiguration *types.JwtTokenTypeConfiguration) []interface{} {
	if jwtTokenTypeConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"key_location": jwtTokenTypeConfiguration.KeyLocation,
	}

	if v := jwtTokenTypeConfiguration.ClaimRegex; v != nil {
		values["claim_regex"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.GroupAttributeField; v != nil {
		values["group_attribute_field"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.Issuer; v != nil {
		values["issuer"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.SecretManagerArn; v != nil {
		values["secrets_manager_arn"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.URL; v != nil {
		values["url"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.UserNameAttributeField; v != nil {
		values["user_name_attribute_field"] = aws.ToString(v)
	}

	return []interface{}{values}
}

// Helpers added. Could be generated or somehow use go 1.18 generics?
func indexEditionValues(input ...types.IndexEdition) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func userContextPolicyValues(input ...types.UserContextPolicy) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func userGroupResolutionModeValues(input ...types.UserGroupResolutionMode) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func keyLocationValues(input ...types.KeyLocation) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}

func IndexStatusValues(input ...types.IndexStatus) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}
