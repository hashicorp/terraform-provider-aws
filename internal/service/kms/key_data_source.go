// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kms_key", name="Key")
func dataSourceKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceKeyRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAWSAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_master_key_spec": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_key_store_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"expiration_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grant_tokens": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrKeyID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateKeyOrAlias,
			},
			"key_manager": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_spec": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_usage": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"multi_region": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"multi_region_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"multi_region_key_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_key": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"replica_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pending_deletion_window_in_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"valid_to": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xks_key_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID := d.Get(names.AttrKeyID).(string)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	if v, ok := d.GetOk("grant_tokens"); ok && len(v.([]interface{})) > 0 {
		input.GrantTokens = flex.ExpandStringValueList(v.([]interface{}))
	}

	output, err := findKey(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", keyID, err)
	}

	d.SetId(aws.ToString(output.KeyId))
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrAWSAccountID, output.AWSAccountId)
	d.Set("cloud_hsm_cluster_id", output.CloudHsmClusterId)
	d.Set(names.AttrCreationDate, aws.ToTime(output.CreationDate).Format(time.RFC3339))
	d.Set("customer_master_key_spec", output.CustomerMasterKeySpec)
	d.Set("custom_key_store_id", output.CustomKeyStoreId)
	if output.DeletionDate != nil {
		d.Set("deletion_date", aws.ToTime(output.DeletionDate).Format(time.RFC3339))
	}
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set("expiration_model", output.ExpirationModel)
	d.Set("key_manager", output.KeyManager)
	d.Set("key_spec", output.KeySpec)
	d.Set("key_state", output.KeyState)
	d.Set("key_usage", output.KeyUsage)
	d.Set("multi_region", output.MultiRegion)
	if output.MultiRegionConfiguration != nil {
		if err := d.Set("multi_region_configuration", []interface{}{flattenMultiRegionConfiguration(output.MultiRegionConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting multi_region_configuration: %s", err)
		}
	} else {
		d.Set("multi_region_configuration", nil)
	}
	d.Set("origin", output.Origin)
	d.Set("pending_deletion_window_in_days", output.PendingDeletionWindowInDays)
	if output.ValidTo != nil {
		d.Set("valid_to", aws.ToTime(output.ValidTo).Format(time.RFC3339))
	}
	if output.XksKeyConfiguration != nil {
		if err := d.Set("xks_key_configuration", []interface{}{flattenXksKeyConfigurationType(output.XksKeyConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting xks_key_configuration: %s", err)
		}
	} else {
		d.Set("xks_key_configuration", nil)
	}

	return diags
}

func flattenMultiRegionConfiguration(apiObject *awstypes.MultiRegionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"multi_region_key_type": apiObject.MultiRegionKeyType,
	}

	if v := apiObject.PrimaryKey; v != nil {
		tfMap["primary_key"] = []interface{}{flattenMultiRegionKey(v)}
	}

	if v := apiObject.ReplicaKeys; v != nil {
		tfMap["replica_keys"] = flattenMultiRegionKeys(v)
	}

	return tfMap
}

func flattenMultiRegionKey(apiObject *awstypes.MultiRegionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenMultiRegionKeys(apiObjects []awstypes.MultiRegionKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenMultiRegionKey(&apiObject))
	}

	return tfList
}

func flattenXksKeyConfigurationType(apiObject *awstypes.XksKeyConfigurationType) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	return tfMap
}
