package s3control

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStorageLensConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStorageLensConfigurationCreate,
		ReadWithoutTimeout:   resourceStorageLensConfigurationRead,
		UpdateWithoutTimeout: resourceStorageLensConfigurationUpdate,
		DeleteWithoutTimeout: resourceStorageLensConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"arn": {
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
												"enabled": {
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
															"enabled": {
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
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"enabled": {
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
																						Type:     schema.TypeFloat,
																						Optional: true,
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
								},
							},
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStorageLensConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}
	configID := d.Get("config_id").(string)
	id := StorageLensConfigurationCreateResourceID(accountID, configID)

	input := &s3control.PutStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	}

	if v, ok := d.GetOk("storage_lens_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.StorageLensConfiguration = expandStorageLensConfiguration(v.([]interface{})[0].(map[string]interface{}))
		input.StorageLensConfiguration.Id = aws.String(configID)
	}

	if len(tags) > 0 {
		input.Tags = StorageLensTags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating S3 Storage Lens Configuration: %s", input)
	_, err := conn.PutStorageLensConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Storage Lens Configuration (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := FindStorageLensConfigurationByAccountIDAndConfigID(ctx, conn, accountID, configID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Storage Lens Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("arn", output.StorageLensArn)
	d.Set("config_id", configID)
	if err := d.Set("storage_lens_configuration", []interface{}{flattenStorageLensConfiguration(output)}); err != nil {
		return diag.Errorf("setting storage_lens_configuration: %s", err)
	}

	tags, err := storageLensConfigurationListTags(ctx, conn, accountID, configID)

	if err != nil {
		return diag.Errorf("listing tags for S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceStorageLensConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChangesExcept("tags", "tags_all") {
		input := &s3control.PutStorageLensConfigurationInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
		}

		if v, ok := d.GetOk("storage_lens_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.StorageLensConfiguration = expandStorageLensConfiguration(v.([]interface{})[0].(map[string]interface{}))
			input.StorageLensConfiguration.Id = aws.String(configID)
		}

		log.Printf("[DEBUG] Updating S3 Storage Lens Configuration: %s", input)
		_, err := conn.PutStorageLensConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating S3 Storage Lens Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := storageLensConfigurationUpdateTags(ctx, conn, accountID, configID, o, n); err != nil {
			return diag.Errorf("updating S3 Storage Lens Configuration (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceStorageLensConfigurationRead(ctx, d, meta)
}

func resourceStorageLensConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, configID, err := StorageLensConfigurationParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting S3 Storage Lens Configuration: %s", d.Id())
	_, err = conn.DeleteStorageLensConfigurationWithContext(ctx, &s3control.DeleteStorageLensConfigurationInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchConfiguration) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Storage Lens Configuration (%s): %s", d.Id(), err)
	}

	return nil
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

func expandStorageLensConfiguration(tfMap map[string]interface{}) *s3control.StorageLensConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.StorageLensConfiguration{}

	if v, ok := tfMap["account_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AccountLevel = expandAccountLevel(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["enabled"].(bool); ok && v {
		apiObject.IsEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandAccountLevel(tfMap map[string]interface{}) *s3control.AccountLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.AccountLevel{}

	if v, ok := tfMap["activity_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActivityMetrics = expandActivityMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["bucket_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.BucketLevel = expandBucketLevel(v[0].(map[string]interface{}))
	} else {
		apiObject.BucketLevel = &s3control.BucketLevel{}
	}

	return apiObject
}

func expandActivityMetrics(tfMap map[string]interface{}) *s3control.ActivityMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.ActivityMetrics{}

	if v, ok := tfMap["enabled"].(bool); ok && v {
		apiObject.IsEnabled = aws.Bool(v)
	}

	return apiObject
}

func expandBucketLevel(tfMap map[string]interface{}) *s3control.BucketLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.BucketLevel{}

	if v, ok := tfMap["activity_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.ActivityMetrics = expandActivityMetrics(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["prefix_level"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PrefixLevel = expandPrefixLevel(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrefixLevel(tfMap map[string]interface{}) *s3control.PrefixLevel {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PrefixLevel{}

	if v, ok := tfMap["activity_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.StorageMetrics = expandPrefixLevelStorageMetrics(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPrefixLevelStorageMetrics(tfMap map[string]interface{}) *s3control.PrefixLevelStorageMetrics {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PrefixLevelStorageMetrics{}

	if v, ok := tfMap["enabled"].(bool); ok && v {
		apiObject.IsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["selection_criteria"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SelectionCriteria = expandSelectionCriteria(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandSelectionCriteria(tfMap map[string]interface{}) *s3control.SelectionCriteria {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.SelectionCriteria{}

	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		apiObject.Delimiter = aws.String(v)
	}

	if v, ok := tfMap["max_depth"].(int); ok && v != 0 {
		apiObject.MaxDepth = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_storage_bytes_percentage"].(float64); ok && v != 0.0 {
		apiObject.MinStorageBytesPercentage = aws.Float64(v)
	}

	return apiObject
}

func flattenStorageLensConfiguration(apiObject *s3control.StorageLensConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AccountLevel; v != nil {
		tfMap["account_level"] = []interface{}{flattenAccountLevel(v)}
	}

	if v := apiObject.IsEnabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenAccountLevel(apiObject *s3control.AccountLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ActivityMetrics; v != nil {
		tfMap["activity_metrics"] = []interface{}{flattenActivityMetrics(v)}
	}

	if v := apiObject.BucketLevel; v != nil {
		tfMap["bucket_level"] = []interface{}{flattenBucketLevel(v)}
	}

	return tfMap
}

func flattenActivityMetrics(apiObject *s3control.ActivityMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IsEnabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenBucketLevel(apiObject *s3control.BucketLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ActivityMetrics; v != nil {
		tfMap["activity_metrics"] = []interface{}{flattenActivityMetrics(v)}
	}

	if v := apiObject.PrefixLevel; v != nil {
		tfMap["prefix_level"] = []interface{}{flattenPrefixLevel(v)}
	}

	return tfMap
}

func flattenPrefixLevel(apiObject *s3control.PrefixLevel) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.StorageMetrics; v != nil {
		tfMap["storage_metrics"] = []interface{}{flattenPrefixLevelStorageMetrics(v)}
	}

	return tfMap
}

func flattenPrefixLevelStorageMetrics(apiObject *s3control.PrefixLevelStorageMetrics) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IsEnabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.SelectionCriteria; v != nil {
		tfMap["selection_criteria"] = []interface{}{flattenSelectionCriteria(v)}
	}

	return tfMap
}

func flattenSelectionCriteria(apiObject *s3control.SelectionCriteria) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Delimiter; v != nil {
		tfMap["delimiter"] = aws.StringValue(v)
	}

	if v := apiObject.MaxDepth; v != nil {
		tfMap["max_depth"] = aws.Int64Value(v)
	}

	if v := apiObject.MinStorageBytesPercentage; v != nil {
		tfMap["min_storage_bytes_percentage"] = aws.Float64Value(v)
	}

	return tfMap
}
