// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Allow IAM role to become visible to the index
	propagationTimeout = 2 * time.Minute

	// validationExceptionMessage describes the error returned when the IAM role has not yet propagated
	validationExceptionMessage = "Please make sure your role exists and has `kendra.amazonaws.com` as trusted entity"
)

// @SDKResource("aws_kendra_index", name="Index")
// @Tags(identifierAttribute="arn")
func ResourceIndex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIndexCreate,
		ReadWithoutTimeout:   resourceIndexRead,
		UpdateWithoutTimeout: resourceIndexUpdate,
		DeleteWithoutTimeout: resourceIndexDelete,
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
			names.AttrARN: {
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
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"document_metadata_configuration_updates": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				MinItems: 0,
				MaxItems: 500,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 30),
						},
						"relevance": {
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDuration: {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 10),
											validation.StringMatch(
												regexache.MustCompile(`[0-9]+[s]`),
												"numeric string followed by the character \"s\"",
											),
										),
									},
									"freshness": {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
									"importance": {
										Type:         schema.TypeInt,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 10),
									},
									"rank_order": {
										Type:             schema.TypeString,
										Computed:         true,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Order](),
									},
									"values_importance_map": {
										Type:     schema.TypeMap,
										Computed: true,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"search": {
							Type:     schema.TypeList,
							Computed: true,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"displayable": {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
									"facetable": {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
									"searchable": {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
									"sortable": {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
								},
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.DocumentAttributeValueType](),
						},
					},
				},
			},
			"edition": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          string(types.IndexEditionEnterpriseEdition),
				ValidateDiagFunc: enum.Validate[types.IndexEdition](),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(
						regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`),
						"The name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			names.AttrRoleARN: {
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
						names.AttrKMSKeyID: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_context_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(types.UserContextPolicyAttributeFilter),
				ValidateDiagFunc: enum.Validate[types.UserContextPolicy](),
			},
			"user_group_resolution_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_group_resolution_mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.UserGroupResolutionMode](),
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
									names.AttrIssuer: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 65),
									},
									"key_location": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.KeyLocation](),
									},
									"secrets_manager_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrURL: {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 2048),
											validation.StringMatch(
												regexache.MustCompile(`^(https?|ftp|file):\/\/([^\s]*)`),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kendra.CreateIndexInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(name),
		RoleArn:     aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
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

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
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
		return sdkdiag.AppendErrorf(diags, "creating Kendra Index (%s): %s", name, err)
	}

	if outputRaw == nil {
		return sdkdiag.AppendErrorf(diags, "creating Kendra Index (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateIndexOutput)

	d.SetId(aws.ToString(output.Id))

	// waiter since the status changes from CREATING to either ACTIVE or FAILED
	if _, err := waitIndexCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Index (%s) creation: %s", d.Id(), err)
	}

	callUpdateIndex := false

	// CreateIndex API does not support capacity_units but UpdateIndex does
	if v, ok := d.GetOk("capacity_units"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		callUpdateIndex = true
	}

	// CreateIndex API does not support document_metadata_configuration_updates but UpdateIndex does
	if v, ok := d.GetOk("document_metadata_configuration_updates"); ok && v.(*schema.Set).Len() >= 13 {
		callUpdateIndex = true
	}

	if callUpdateIndex {
		return append(diags, resourceIndexUpdate(ctx, d, meta)...)
	}

	return append(diags, resourceIndexRead(ctx, d, meta)...)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	resp, err := findIndexByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Index (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Kendra Index (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s", d.Id()),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("edition", resp.Edition)
	d.Set("error_message", resp.ErrorMessage)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrRoleARN, resp.RoleArn)
	d.Set(names.AttrStatus, resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))
	d.Set("user_context_policy", resp.UserContextPolicy)

	if err := d.Set("capacity_units", flattenCapacityUnits(resp.CapacityUnits)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("document_metadata_configuration_updates", flattenDocumentMetadataConfigurations(resp.DocumentMetadataConfigurations)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("index_statistics", flattenIndexStatistics(resp.IndexStatistics)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(resp.ServerSideEncryptionConfiguration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("user_group_resolution_configuration", flattenUserGroupResolutionConfiguration(resp.UserGroupResolutionConfiguration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("user_token_configurations", flattenUserTokenConfigurations(resp.UserTokenConfigurations)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id := d.Id()

	if d.HasChanges("capacity_units", names.AttrDescription, "document_metadata_configuration_updates", names.AttrName, names.AttrRoleARN, "user_context_policy", "user_group_resolution_configuration", "user_token_configurations") {
		input := &kendra.UpdateIndexInput{
			Id: aws.String(id),
		}
		if d.HasChange("capacity_units") {
			input.CapacityUnits = expandCapacityUnits(d.Get("capacity_units").([]interface{}))
		}
		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}
		if d.HasChange("document_metadata_configuration_updates") {
			input.DocumentMetadataConfigurationUpdates = expandDocumentMetadataConfigurationUpdates(d.Get("document_metadata_configuration_updates").(*schema.Set).List())
		}
		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}
		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
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

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
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
			return sdkdiag.AppendErrorf(diags, "updating Index (%s): %s", d.Id(), err)
		}

		// waiter since the status changes from UPDATING to either ACTIVE or FAILED
		if _, err := waitIndexUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Index (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIndexRead(ctx, d, meta)...)
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id := d.Id()

	_, err := conn.DeleteIndex(ctx, &kendra.DeleteIndexInput{
		Id: aws.String(id),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Index (%s): %s", d.Id(), err)
	}

	if _, err := waitIndexDeleted(ctx, conn, id, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Index (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findIndexByID(ctx context.Context, conn *kendra.Client, id string) (*kendra.DescribeIndexOutput, error) {
	input := &kendra.DescribeIndexInput{
		Id: aws.String(id),
	}

	output, err := conn.DescribeIndex(ctx, input)

	if err != nil {
		var resourceNotFoundException *types.ResourceNotFoundException

		if errors.As(err, &resourceNotFoundException) {
			return nil, &retry.NotFoundError{
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

func statusIndex(ctx context.Context, conn *kendra.Client, id string) retry.StateRefreshFunc {
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
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IndexStatusCreating),
		Target:  enum.Slice(types.IndexStatusActive),
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
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IndexStatusUpdating),
		Target:  enum.Slice(types.IndexStatusActive),
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
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IndexStatusDeleting),
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

func expandDocumentMetadataConfigurationUpdates(documentMetadataConfigurationUpdates []interface{}) []types.DocumentMetadataConfiguration {
	if len(documentMetadataConfigurationUpdates) == 0 {
		return nil
	}

	documentMetadataConfigurationUpdateConfigs := []types.DocumentMetadataConfiguration{}

	for _, documentMetadataConfigurationUpdate := range documentMetadataConfigurationUpdates {
		tfMap := documentMetadataConfigurationUpdate.(map[string]interface{})
		documentMetadataConfigurationUpdateConfig := types.DocumentMetadataConfiguration{
			Name: aws.String(tfMap[names.AttrName].(string)),
			Type: types.DocumentAttributeValueType(tfMap[names.AttrType].(string)),
		}

		documentMetadataConfigurationUpdateConfig.Relevance = expandRelevance(tfMap["relevance"].([]interface{}), tfMap[names.AttrType].(string))
		documentMetadataConfigurationUpdateConfig.Search = expandSearch(tfMap["search"].([]interface{}))

		documentMetadataConfigurationUpdateConfigs = append(documentMetadataConfigurationUpdateConfigs, documentMetadataConfigurationUpdateConfig)
	}

	return documentMetadataConfigurationUpdateConfigs
}

func expandRelevance(relevance []interface{}, documentAttributeValueType string) *types.Relevance {
	if len(relevance) == 0 || relevance[0] == nil {
		return nil
	}

	tfMap, ok := relevance[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Relevance{}

	if v, ok := tfMap[names.AttrDuration].(string); ok && v != "" {
		result.Duration = aws.String(v)
	}

	// You can only set the Freshness field on one DATE type field
	if v, ok := tfMap["freshness"].(bool); ok && documentAttributeValueType == string(types.DocumentAttributeValueTypeDateValue) {
		result.Freshness = aws.Bool(v)
	}

	if v, ok := tfMap["importance"].(int); ok {
		result.Importance = aws.Int32(int32(v))
	}

	if v, ok := tfMap["rank_order"].(string); ok && v != "" {
		result.RankOrder = types.Order(v)
	}

	if v, ok := tfMap["values_importance_map"].(map[string]interface{}); ok && len(v) > 0 {
		result.ValueImportanceMap = flex.ExpandInt32Map(v)
	}

	return result
}

func expandSearch(search []interface{}) *types.Search {
	if len(search) == 0 || search[0] == nil {
		return nil
	}

	tfMap, ok := search[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Search{}

	if v, ok := tfMap["displayable"].(bool); ok {
		result.Displayable = v
	}

	if v, ok := tfMap["facetable"].(bool); ok {
		result.Facetable = v
	}

	if v, ok := tfMap["searchable"].(bool); ok {
		result.Searchable = v
	}

	if v, ok := tfMap["sortable"].(bool); ok {
		result.Sortable = v
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

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
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

	if v, ok := tfMap[names.AttrIssuer].(string); ok && v != "" {
		result.Issuer = aws.String(v)
	}

	if v, ok := tfMap["secrets_manager_arn"].(string); ok && v != "" {
		result.SecretManagerArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrURL].(string); ok && v != "" {
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
			names.AttrName: documentMetadataConfiguration.Name,
			"relevance":    flattenRelevance(documentMetadataConfiguration.Relevance, string(documentMetadataConfiguration.Type)),
			"search":       flattenSearch(documentMetadataConfiguration.Search),
			names.AttrType: documentMetadataConfiguration.Type,
		}

		documentMetadataConfigurationsList = append(documentMetadataConfigurationsList, values)
	}

	return documentMetadataConfigurationsList
}

func flattenRelevance(relevance *types.Relevance, documentAttributeValueType string) []interface{} {
	if relevance == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"rank_order": relevance.RankOrder,
	}

	if v := relevance.Duration; v != nil {
		values[names.AttrDuration] = aws.ToString(v)
	}

	if v := relevance.Freshness; v != nil && documentAttributeValueType == string(types.DocumentAttributeValueTypeDateValue) {
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
		values[names.AttrKMSKeyID] = aws.ToString(v)
	}

	return []interface{}{values}
}

func flattenUserGroupResolutionConfiguration(userGroupResolutionConfiguration *types.UserGroupResolutionConfiguration) []interface{} {
	if userGroupResolutionConfiguration == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"user_group_resolution_mode": userGroupResolutionConfiguration.UserGroupResolutionMode,
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
		values[names.AttrIssuer] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.SecretManagerArn; v != nil {
		values["secrets_manager_arn"] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.URL; v != nil {
		values[names.AttrURL] = aws.ToString(v)
	}

	if v := jwtTokenTypeConfiguration.UserNameAttributeField; v != nil {
		values["user_name_attribute_field"] = aws.ToString(v)
	}

	return []interface{}{values}
}
