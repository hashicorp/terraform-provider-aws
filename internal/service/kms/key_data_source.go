package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceKeyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_master_key_spec": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
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
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: ValidateKeyOrAlias,
			},
			"key_manager": {
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
									"arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"region": {
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
									"arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"region": {
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
			"valid_to": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn()

	keyID := d.Get("key_id").(string)
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	if v, ok := d.GetOk("grant_tokens"); ok && len(v.([]interface{})) > 0 {
		input.GrantTokens = flex.ExpandStringList(v.([]interface{}))
	}

	output, err := conn.DescribeKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", keyID, err)
	}

	keyMetadata := output.KeyMetadata
	d.SetId(aws.StringValue(keyMetadata.KeyId))
	d.Set("arn", keyMetadata.Arn)
	d.Set("aws_account_id", keyMetadata.AWSAccountId)
	d.Set("creation_date", aws.TimeValue(keyMetadata.CreationDate).Format(time.RFC3339))
	d.Set("customer_master_key_spec", keyMetadata.CustomerMasterKeySpec)
	if keyMetadata.DeletionDate != nil {
		d.Set("deletion_date", aws.TimeValue(keyMetadata.DeletionDate).Format(time.RFC3339))
	}
	d.Set("description", keyMetadata.Description)
	d.Set("enabled", keyMetadata.Enabled)
	d.Set("expiration_model", keyMetadata.ExpirationModel)
	d.Set("key_manager", keyMetadata.KeyManager)
	d.Set("key_state", keyMetadata.KeyState)
	d.Set("key_usage", keyMetadata.KeyUsage)
	d.Set("multi_region", keyMetadata.MultiRegion)
	if keyMetadata.MultiRegionConfiguration != nil {
		if err := d.Set("multi_region_configuration", []interface{}{flattenMultiRegionConfiguration(keyMetadata.MultiRegionConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting multi_region_configuration: %s", err)
		}
	} else {
		d.Set("multi_region_configuration", nil)
	}
	d.Set("origin", keyMetadata.Origin)
	if keyMetadata.ValidTo != nil {
		d.Set("valid_to", aws.TimeValue(keyMetadata.ValidTo).Format(time.RFC3339))
	}

	return diags
}

func flattenMultiRegionConfiguration(apiObject *kms.MultiRegionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MultiRegionKeyType; v != nil {
		tfMap["multi_region_key_type"] = aws.StringValue(v)
	}

	if v := apiObject.PrimaryKey; v != nil {
		tfMap["primary_key"] = []interface{}{flattenMultiRegionKey(v)}
	}

	if v := apiObject.ReplicaKeys; v != nil {
		tfMap["replica_keys"] = flattenMultiRegionKeys(v)
	}

	return tfMap
}

func flattenMultiRegionKey(apiObject *kms.MultiRegionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = aws.StringValue(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap["region"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenMultiRegionKeys(apiObjects []*kms.MultiRegionKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenMultiRegionKey(apiObject))
	}

	return tfList
}
