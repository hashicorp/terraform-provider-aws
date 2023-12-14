// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_s3_bucket_inventory")
func ResourceBucketInventory() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketInventoryPut,
		ReadWithoutTimeout:   resourceBucketInventoryRead,
		UpdateWithoutTimeout: resourceBucketInventoryPut,
		DeleteWithoutTimeout: resourceBucketInventoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidAccountID,
									},
									"bucket_arn": {
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
													Type:          schema.TypeList,
													Optional:      true,
													MaxItems:      1,
													ConflictsWith: []string{"destination.0.bucket.0.encryption.0.sse_s3"},
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"key_id": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
														},
													},
												},
												"sse_s3": {
													Type:          schema.TypeList,
													Optional:      true,
													MaxItems:      1,
													ConflictsWith: []string{"destination.0.bucket.0.encryption.0.sse_kms"},
													Elem: &schema.Resource{
														// No options currently; just existence of "sse_s3".
														Schema: map[string]*schema.Schema{},
													},
												},
											},
										},
									},
									"format": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.InventoryFormat](),
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"included_object_versions": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.InventoryIncludedObjectVersions](),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			"optional_fields": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[types.InventoryOptionalField](),
				},
			},
			"schedule": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"frequency": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.InventoryFrequency](),
						},
					},
				},
			},
		},
	}
}

func resourceBucketInventoryPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	name := d.Get("name").(string)
	inventoryConfiguration := &types.InventoryConfiguration{
		Id:        aws.String(name),
		IsEnabled: aws.Bool(d.Get("enabled").(bool)),
	}

	if v, ok := d.GetOk("destination"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})["bucket"].([]interface{})[0].(map[string]interface{})
		inventoryConfiguration.Destination = &types.InventoryDestination{
			S3BucketDestination: expandInventoryBucketDestination(tfMap),
		}
	}

	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		inventoryConfiguration.Filter = expandInventoryFilter(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("included_object_versions"); ok {
		inventoryConfiguration.IncludedObjectVersions = types.InventoryIncludedObjectVersions(v.(string))
	}

	if v, ok := d.GetOk("optional_fields"); ok && v.(*schema.Set).Len() > 0 {
		inventoryConfiguration.OptionalFields = flex.ExpandStringyValueSet[types.InventoryOptionalField](v.(*schema.Set))
	}

	if v, ok := d.GetOk("schedule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		inventoryConfiguration.Schedule = &types.InventorySchedule{
			Frequency: types.InventoryFrequency(tfMap["frequency"].(string)),
		}
	}

	bucket := d.Get("bucket").(string)
	input := &s3.PutBucketInventoryConfigurationInput{
		Bucket:                 aws.String(bucket),
		Id:                     aws.String(name),
		InventoryConfiguration: inventoryConfiguration,
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketInventoryConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "InventoryConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return diag.Errorf("creating S3 Bucket (%s) Inventory: %s", bucket, err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("%s:%s", bucket, name))

		_, err = tfresource.RetryWhenNotFound(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
			return findInventoryConfiguration(ctx, conn, bucket, name)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Inventory (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketInventoryRead(ctx, d, meta)...)
}

func resourceBucketInventoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketInventoryParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	ic, err := findInventoryConfiguration(ctx, conn, bucket, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Inventory (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Bucket Inventory (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	if v := ic.Destination; v != nil {
		tfMap := map[string]interface{}{
			"bucket": flattenInventoryBucketDestination(v.S3BucketDestination),
		}
		if err := d.Set("destination", []map[string]interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting destination: %s", err)
		}
	}
	d.Set("enabled", ic.IsEnabled)
	if err := d.Set("filter", flattenInventoryFilter(ic.Filter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filter: %s", err)
	}
	d.Set("included_object_versions", ic.IncludedObjectVersions)
	d.Set("name", name)
	d.Set("optional_fields", ic.OptionalFields)
	if err := d.Set("schedule", flattenInventorySchedule(ic.Schedule)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
	}

	return diags
}

func resourceBucketInventoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, name, err := BucketInventoryParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.DeleteBucketInventoryConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(name),
	}

	log.Printf("[DEBUG] Deleting S3 Bucket Inventory: %s", d.Id())
	_, err = conn.DeleteBucketInventoryConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Inventory (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return findInventoryConfiguration(ctx, conn, bucket, name)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for S3 Bucket Inventory (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandInventoryFilter(m map[string]interface{}) *types.InventoryFilter {
	v, ok := m["prefix"]
	if !ok {
		return nil
	}
	return &types.InventoryFilter{
		Prefix: aws.String(v.(string)),
	}
}

func flattenInventoryFilter(filter *types.InventoryFilter) []map[string]interface{} {
	if filter == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, 1)

	m := make(map[string]interface{})
	if filter.Prefix != nil {
		m["prefix"] = aws.ToString(filter.Prefix)
	}

	result = append(result, m)

	return result
}

func flattenInventorySchedule(schedule *types.InventorySchedule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)
	m := map[string]interface{}{
		"frequency": schedule.Frequency,
	}
	result = append(result, m)

	return result
}

func expandInventoryBucketDestination(m map[string]interface{}) *types.InventoryS3BucketDestination {
	destination := &types.InventoryS3BucketDestination{
		Format: types.InventoryFormat(m["format"].(string)),
		Bucket: aws.String(m["bucket_arn"].(string)),
	}

	if v, ok := m["account_id"]; ok && v.(string) != "" {
		destination.AccountId = aws.String(v.(string))
	}

	if v, ok := m["prefix"]; ok && v.(string) != "" {
		destination.Prefix = aws.String(v.(string))
	}

	if v, ok := m["encryption"].([]interface{}); ok && len(v) > 0 {
		encryptionMap := v[0].(map[string]interface{})

		encryption := &types.InventoryEncryption{}

		for k, v := range encryptionMap {
			data := v.([]interface{})

			if len(data) == 0 {
				continue
			}

			switch k {
			case "sse_kms":
				m := data[0].(map[string]interface{})
				encryption.SSEKMS = &types.SSEKMS{
					KeyId: aws.String(m["key_id"].(string)),
				}
			case "sse_s3":
				encryption.SSES3 = &types.SSES3{}
			}
		}

		destination.Encryption = encryption
	}

	return destination
}

func flattenInventoryBucketDestination(destination *types.InventoryS3BucketDestination) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	m := map[string]interface{}{
		"format":     destination.Format,
		"bucket_arn": aws.ToString(destination.Bucket),
	}

	if destination.AccountId != nil {
		m["account_id"] = aws.ToString(destination.AccountId)
	}
	if destination.Prefix != nil {
		m["prefix"] = aws.ToString(destination.Prefix)
	}

	if destination.Encryption != nil {
		encryption := make(map[string]interface{}, 1)
		if destination.Encryption.SSES3 != nil {
			encryption["sse_s3"] = []map[string]interface{}{{}}
		} else if destination.Encryption.SSEKMS != nil {
			encryption["sse_kms"] = []map[string]interface{}{
				{
					"key_id": aws.ToString(destination.Encryption.SSEKMS.KeyId),
				},
			}
		}
		m["encryption"] = []map[string]interface{}{encryption}
	}

	result = append(result, m)

	return result
}

func BucketInventoryParseID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("please make sure the ID is in the form BUCKET:NAME (i.e. my-bucket:EntireBucket")
	}
	bucket := idParts[0]
	name := idParts[1]
	return bucket, name, nil
}

func findInventoryConfiguration(ctx context.Context, conn *s3.Client, bucket, id string) (*types.InventoryConfiguration, error) {
	input := &s3.GetBucketInventoryConfigurationInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	}

	output, err := conn.GetBucketInventoryConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InventoryConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InventoryConfiguration, nil
}
