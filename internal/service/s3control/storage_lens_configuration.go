// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_s3control_storage_lens_configuration", name="Storage Lens Configuration")
// @Tags
func resourceStorageLensConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStorageLensConfigurationCreate,
		ReadWithoutTimeout:   resourceStorageLensConfigurationRead,
		UpdateWithoutTimeout: resourceStorageLensConfigurationUpdate,
		DeleteWithoutTimeout: resourceStorageLensConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"storage_lens_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_level": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"activity_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"advanced_cost_optimization_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"advanced_data_protection_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"bucket_level": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"activity_metrics": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"advanced_cost_optimization_metrics": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"advanced_data_protection_metrics": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"detailed_status_code_metrics": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrEnabled: {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"prefix_level": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"storage_metrics": {
																Type:     schema.TypeList,
																Required: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrEnabled: {
																			Type:     schema.TypeBool,
																			Optional: true,
																		},
																		"selection_criteria": {
																			Type:     schema.TypeList,
																			Optional: true,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"delimiter": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"max_depth": {
																						Type:     schema.TypeInt,
																						Optional: true,
																					},
																					"min_storage_bytes_percentage": {
																						Type:         schema.TypeFloat,
																						Optional:     true,
																						ValidateFunc: validation.FloatBetween(1.0, 100),
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									"detailed_status_code_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"aws_org": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"data_export": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cloud_watch_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Required: true,
												},
											},
										},
									},
									"s3_bucket_destination": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrAccountID: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidAccountID,
												},
												names.AttrARN: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidARN,
												},
												"encryption": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"sse_kms": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrKeyID: {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: verify.ValidARN,
																		},
																	},
																},
															},
															"sse_s3": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{},
																},
															},
														},
													},
												},
												names.AttrFormat: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.Format](),
												},
												"output_schema_version": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.OutputSchemaVersion](),
												},
												names.AttrPrefix: {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"exclude": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"buckets": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
									"regions": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidRegionName,
										},
									},
								},
							},
						},
						"include": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"buckets": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidARN,
										},
									},
									"regions": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: verify.ValidRegionName,
										},
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStorageLensConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAccountID); ok {
		accountID = v.(string)
	}
	configID := d.Get("config_id").(string)
	id := StorageLensConfigurationCreateResourceID(accountID, configID)
	input := &s3control.PutStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
		Tags:      storageLensTags(keyValueTagsS3(ctx, getTagsInS3(ctx))),
	}

	if v, ok := d.GetOk("storage_lens_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.StorageLensConfiguration = expandStorageLensConfiguration(v.([]interface{})[0].(map[string]interface{}))
		input.StorageLensConfiguration.Id = aws.String(configID)
	}

	_, err := conn.PutStorageLensConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Storage Lens Configuration (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceStorageLensConfigurationRead(ctx, d, meta)...)
}

func resourceStorageLensConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Storage Lens Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrARN, output.StorageLensArn)
	d.Set("config_id", configID)
	if err := d.Set("storage_lens_configuration", []interface{}{flattenStorageLensConfiguration(output)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_lens_configuration: %s", err)
	}

	tags, err := storageLensConfigurationListTags(ctx, conn, accountID, configID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	setTagsOutS3(ctx, tagsS3(tags))

	return diags
}

func resourceStorageLensConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &s3control.PutStorageLensConfigurationInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
		}

		if v, ok := d.GetOk("storage_lens_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.StorageLensConfiguration = expandStorageLensConfiguration(v.([]interface{})[0].(map[string]interface{}))
			input.StorageLensConfiguration.Id = aws.String(configID)
		}

		_, err := conn.PutStorageLensConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating S3 Storage Lens Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)

		if err := storageLensConfigurationUpdateTags(ctx, conn, accountID, configID, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating S3 Storage Lens Configuration (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStorageLensConfigurationRead(ctx, d, meta)...)
}

func resourceStorageLensConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting S3 Storage Lens Configuration: %s", d.Id())
	_, err = conn.DeleteStorageLensConfiguration(ctx, &s3control.DeleteStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	})

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

const storageLensConfigurationResourceIDSeparator = ":"

func StorageLensConfigurationCreateResourceID(accountID, configID string) string {
	parts := []string{accountID, configID}
	id := strings.Join(parts, storageLensConfigurationResourceIDSeparator)

	return id
}

func StorageLensConfigurationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, storageLensConfigurationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]sconfig-id", id, storageLensConfigurationResourceIDSeparator)
}

func findStorageLensConfigurationByAccountIDAndConfigID(ctx context.Context, conn *s3control.Client, accountID, configID string) (*types.StorageLensConfiguration, error) {
	input := &s3control.GetStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	}

	output, err := conn.GetStorageLensConfiguration(ctx, input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.StorageLensConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.StorageLensConfiguration, nil
}

func storageLensTags(tags tftags.KeyValueTags) []types.StorageLensTag {
	result := make([]types.StorageLensTag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := types.StorageLensTag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

func keyValueTagsFromStorageLensTags(ctx context.Context, tags []types.StorageLensTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

func storageLensConfigurationListTags(ctx context.Context, conn *s3control.Client, accountID, configID string) (tftags.KeyValueTags, error) {
	input := &s3control.GetStorageLensConfigurationTaggingInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	}

	output, err := conn.GetStorageLensConfigurationTagging(ctx, input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTagsFromStorageLensTags(ctx, output.Tags), nil
}

func storageLensConfigurationUpdateTags(ctx context.Context, conn *s3control.Client, accountID, configID string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := storageLensConfigurationListTags(ctx, conn, accountID, configID)

	if err != nil {
		return fmt.Errorf("listing tags: %s", err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3control.PutStorageLensConfigurationTaggingInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
			Tags:      storageLensTags(newTags.Merge(ignoredTags)),
		}

		_, err := conn.PutStorageLensConfigurationTagging(ctx, input)

		if err != nil {
			return fmt.Errorf("setting tags: %s", err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3control.DeleteStorageLensConfigurationTaggingInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
		}

		_, err := conn.DeleteStorageLensConfigurationTagging(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting tags: %s", err)
		}
	}

	return nil
}

func expandStorageLensConfiguration(tfMap map[string]interface{}) *types.StorageLensConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StorageLensConfiguration{}

	if v, ok := tfMap["account_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccountLevel = expandAccountLevel(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["aws_org"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AwsOrg = expandStorageLensAwsOrg(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["data_export"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DataExport = expandStorageLensDataExport(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	if v, ok := tfMap["exclude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Exclude = expandExclude(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["include"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Include = expandInclude(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAccountLevel(tfMap map[string]interface{}) *types.AccountLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AccountLevel{}

	if v, ok := tfMap["activity_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActivityMetrics = expandActivityMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["advanced_cost_optimization_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AdvancedCostOptimizationMetrics = expandAdvancedCostOptimizationMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["advanced_data_protection_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AdvancedDataProtectionMetrics = expandAdvancedDataProtectionMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["bucket_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.BucketLevel = expandBucketLevel(v[0].(map[string]interface{}))
	} else {
		apiObject.BucketLevel = &types.BucketLevel{}
	}

	if v, ok := tfMap["detailed_status_code_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DetailedStatusCodesMetrics = expandDetailedStatusCodesMetrics(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandActivityMetrics(tfMap map[string]interface{}) *types.ActivityMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ActivityMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	return apiObject
}

func expandBucketLevel(tfMap map[string]interface{}) *types.BucketLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BucketLevel{}

	if v, ok := tfMap["activity_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActivityMetrics = expandActivityMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["advanced_cost_optimization_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AdvancedCostOptimizationMetrics = expandAdvancedCostOptimizationMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["advanced_data_protection_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AdvancedDataProtectionMetrics = expandAdvancedDataProtectionMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["detailed_status_code_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.DetailedStatusCodesMetrics = expandDetailedStatusCodesMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["prefix_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PrefixLevel = expandPrefixLevel(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAdvancedCostOptimizationMetrics(tfMap map[string]interface{}) *types.AdvancedCostOptimizationMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AdvancedCostOptimizationMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	return apiObject
}

func expandAdvancedDataProtectionMetrics(tfMap map[string]interface{}) *types.AdvancedDataProtectionMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.AdvancedDataProtectionMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	return apiObject
}

func expandDetailedStatusCodesMetrics(tfMap map[string]interface{}) *types.DetailedStatusCodesMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DetailedStatusCodesMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	return apiObject
}

func expandPrefixLevel(tfMap map[string]interface{}) *types.PrefixLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PrefixLevel{}

	if v, ok := tfMap["storage_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StorageMetrics = expandPrefixLevelStorageMetrics(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrefixLevelStorageMetrics(tfMap map[string]interface{}) *types.PrefixLevelStorageMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PrefixLevelStorageMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	if v, ok := tfMap["selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SelectionCriteria = expandSelectionCriteria(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandSelectionCriteria(tfMap map[string]interface{}) *types.SelectionCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SelectionCriteria{}

	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		apiObject.Delimiter = aws.String(v)
	}

	if v, ok := tfMap["max_depth"].(int); ok && v != 0 {
		apiObject.MaxDepth = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_storage_bytes_percentage"].(float64); ok && v != 0.0 {
		apiObject.MinStorageBytesPercentage = aws.Float64(v)
	}

	return apiObject
}

func expandStorageLensAwsOrg(tfMap map[string]interface{}) *types.StorageLensAwsOrg { // nosemgrep:ci.aws-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StorageLensAwsOrg{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func expandStorageLensDataExport(tfMap map[string]interface{}) *types.StorageLensDataExport {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StorageLensDataExport{}

	if v, ok := tfMap["cloud_watch_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchMetrics = expandCloudWatchMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3_bucket_destination"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3BucketDestination = expandS3BucketDestination(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchMetrics(tfMap map[string]interface{}) *types.CloudWatchMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CloudWatchMetrics{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.IsEnabled = v
	}

	return apiObject
}

func expandS3BucketDestination(tfMap map[string]interface{}) *types.S3BucketDestination {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.S3BucketDestination{}

	if v, ok := tfMap[names.AttrAccountID].(string); ok && v != "" {
		apiObject.AccountId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["encryption"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Encryption = expandStorageLensDataExportEncryption(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrFormat].(string); ok && v != "" {
		apiObject.Format = types.Format(v)
	}

	if v, ok := tfMap["output_schema_version"].(string); ok && v != "" {
		apiObject.OutputSchemaVersion = types.OutputSchemaVersion(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandStorageLensDataExportEncryption(tfMap map[string]interface{}) *types.StorageLensDataExportEncryption {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StorageLensDataExportEncryption{}

	if v, ok := tfMap["sse_kms"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SSEKMS = expandSSEKMS(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sse_s3"].([]interface{}); ok && len(v) > 0 {
		apiObject.SSES3 = &types.SSES3{}
	}

	return apiObject
}

func expandSSEKMS(tfMap map[string]interface{}) *types.SSEKMS {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SSEKMS{}

	if v, ok := tfMap[names.AttrKeyID].(string); ok && v != "" {
		apiObject.KeyId = aws.String(v)
	}

	return apiObject
}

func expandExclude(tfMap map[string]interface{}) *types.Exclude {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Exclude{}

	if v, ok := tfMap["buckets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Buckets = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["regions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Regions = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandInclude(tfMap map[string]interface{}) *types.Include {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Include{}

	if v, ok := tfMap["buckets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Buckets = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["regions"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Regions = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenStorageLensConfiguration(apiObject *types.StorageLensConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountLevel; v != nil {
		tfMap["account_level"] = []interface{}{flattenAccountLevel(v)}
	}

	if v := apiObject.AwsOrg; v != nil {
		tfMap["aws_org"] = []interface{}{flattenStorageLensAwsOrg(v)}
	}

	if v := apiObject.DataExport; v != nil {
		tfMap["data_export"] = []interface{}{flattenStorageLensDataExport(v)}
	}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	if v := apiObject.Exclude; v != nil {
		tfMap["exclude"] = []interface{}{flattenExclude(v)}
	}

	if v := apiObject.Include; v != nil {
		tfMap["include"] = []interface{}{flattenInclude(v)}
	}

	return tfMap
}

func flattenAccountLevel(apiObject *types.AccountLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ActivityMetrics; v != nil {
		tfMap["activity_metrics"] = []interface{}{flattenActivityMetrics(v)}
	}

	if v := apiObject.AdvancedCostOptimizationMetrics; v != nil {
		tfMap["advanced_cost_optimization_metrics"] = []interface{}{flattenAdvancedCostOptimizationMetrics(v)}
	}

	if v := apiObject.AdvancedDataProtectionMetrics; v != nil {
		tfMap["advanced_data_protection_metrics"] = []interface{}{flattenAdvancedDataProtectionMetrics(v)}
	}

	if v := apiObject.BucketLevel; v != nil {
		tfMap["bucket_level"] = []interface{}{flattenBucketLevel(v)}
	}

	if v := apiObject.DetailedStatusCodesMetrics; v != nil {
		tfMap["detailed_status_code_metrics"] = []interface{}{flattenDetailedStatusCodesMetrics(v)}
	}

	return tfMap
}

func flattenActivityMetrics(apiObject *types.ActivityMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	return tfMap
}

func flattenBucketLevel(apiObject *types.BucketLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ActivityMetrics; v != nil {
		tfMap["activity_metrics"] = []interface{}{flattenActivityMetrics(v)}
	}

	if v := apiObject.AdvancedCostOptimizationMetrics; v != nil {
		tfMap["advanced_cost_optimization_metrics"] = []interface{}{flattenAdvancedCostOptimizationMetrics(v)}
	}

	if v := apiObject.AdvancedDataProtectionMetrics; v != nil {
		tfMap["advanced_data_protection_metrics"] = []interface{}{flattenAdvancedDataProtectionMetrics(v)}
	}

	if v := apiObject.DetailedStatusCodesMetrics; v != nil {
		tfMap["detailed_status_code_metrics"] = []interface{}{flattenDetailedStatusCodesMetrics(v)}
	}

	if v := apiObject.PrefixLevel; v != nil {
		tfMap["prefix_level"] = []interface{}{flattenPrefixLevel(v)}
	}

	return tfMap
}

func flattenAdvancedCostOptimizationMetrics(apiObject *types.AdvancedCostOptimizationMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	return tfMap
}

func flattenAdvancedDataProtectionMetrics(apiObject *types.AdvancedDataProtectionMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	return tfMap
}

func flattenDetailedStatusCodesMetrics(apiObject *types.DetailedStatusCodesMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	return tfMap
}

func flattenPrefixLevel(apiObject *types.PrefixLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.StorageMetrics; v != nil {
		tfMap["storage_metrics"] = []interface{}{flattenPrefixLevelStorageMetrics(v)}
	}

	return tfMap
}

func flattenPrefixLevelStorageMetrics(apiObject *types.PrefixLevelStorageMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	if v := apiObject.SelectionCriteria; v != nil {
		tfMap["selection_criteria"] = []interface{}{flattenSelectionCriteria(v)}
	}

	return tfMap
}

func flattenSelectionCriteria(apiObject *types.SelectionCriteria) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Delimiter; v != nil {
		tfMap["delimiter"] = aws.ToString(v)
	}

	tfMap["max_depth"] = aws.ToInt32(apiObject.MaxDepth)
	tfMap["min_storage_bytes_percentage"] = aws.ToFloat64(apiObject.MinStorageBytesPercentage)

	return tfMap
}

func flattenStorageLensAwsOrg(apiObject *types.StorageLensAwsOrg) map[string]interface{} { // nosemgrep:ci.aws-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	return tfMap
}

func flattenStorageLensDataExport(apiObject *types.StorageLensDataExport) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchMetrics; v != nil {
		tfMap["cloud_watch_metrics"] = []interface{}{flattenCloudWatchMetrics(v)}
	}

	if v := apiObject.S3BucketDestination; v != nil {
		tfMap["s3_bucket_destination"] = []interface{}{flattenS3BucketDestination(v)}
	}

	return tfMap
}

func flattenCloudWatchMetrics(apiObject *types.CloudWatchMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrEnabled] = apiObject.IsEnabled

	return tfMap
}

func flattenS3BucketDestination(apiObject *types.S3BucketDestination) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountId; v != nil {
		tfMap[names.AttrAccountID] = aws.ToString(v)
	}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	if v := apiObject.Encryption; v != nil {
		tfMap["encryption"] = []interface{}{flattenStorageLensDataExportEncryption(v)}
	}

	tfMap[names.AttrFormat] = apiObject.Format
	tfMap["output_schema_version"] = apiObject.OutputSchemaVersion

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return tfMap
}

func flattenStorageLensDataExportEncryption(apiObject *types.StorageLensDataExportEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SSEKMS; v != nil {
		tfMap["sse_kms"] = []interface{}{flattenSSEKMS(v)}
	}

	if v := apiObject.SSES3; v != nil {
		tfMap["sse_s3"] = []interface{}{flattenSSES3(v)}
	}

	return tfMap
}

func flattenSSEKMS(apiObject *types.SSEKMS) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.KeyId; v != nil {
		tfMap[names.AttrKeyID] = aws.ToString(v)
	}

	return tfMap
}

func flattenSSES3(apiObject *types.SSES3) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	return tfMap
}

func flattenExclude(apiObject *types.Exclude) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["buckets"] = apiObject.Buckets
	tfMap["regions"] = apiObject.Regions

	return tfMap
}

func flattenInclude(apiObject *types.Include) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["buckets"] = apiObject.Buckets
	tfMap["regions"] = apiObject.Regions

	return tfMap
}
