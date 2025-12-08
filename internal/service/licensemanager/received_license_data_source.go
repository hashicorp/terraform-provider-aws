// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_licensemanager_received_license", name="Received License")
func dataSourceReceivedLicense() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceReceivedLicenseRead,
		Schema: map[string]*schema.Schema{
			"beneficiary": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"consumption_configuration": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"borrow_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow_early_check_in": {
										Computed: true,
										Type:     schema.TypeBool,
									},
									"max_time_to_live_in_minutes": {
										Computed: true,
										Type:     schema.TypeInt,
									},
								},
							},
						},
						"provisional_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_time_to_live_in_minutes": {
										Computed: true,
										Type:     schema.TypeInt,
									},
								},
							},
						},
						"renew_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrCreateTime: {
				Computed: true,
				Type:     schema.TypeString,
			},
			"entitlements": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_check_in": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"overage": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrUnit: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"home_region": {
				Computed: true,
				Type:     schema.TypeString,
			},
			names.AttrIssuer: {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_fingerprint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sign_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"license_arn": {
				Required:     true,
				Type:         schema.TypeString,
				ValidateFunc: verify.ValidARN,
			},
			"license_metadata": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"license_name": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"product_name": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"product_sku": {
				Computed: true,
				Type:     schema.TypeString,
			},
			"received_metadata": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_operations": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"received_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"received_status_reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Computed: true,
				Type:     schema.TypeString,
			},
			"validity": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"begin": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrVersion: {
				Computed: true,
				Type:     schema.TypeString,
			},
		},
	}
}

func dataSourceReceivedLicenseRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	arn := d.Get("license_arn").(string)
	license, err := findReceivedLicenseByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Received License (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(license.LicenseArn))
	d.Set("beneficiary", license.Beneficiary)
	if err := d.Set("consumption_configuration", []any{flattenConsumptionConfiguration(license.ConsumptionConfiguration)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting consumption_configuration: %s", err)
	}
	if v := aws.ToString(license.CreateTime); v != "" {
		d.Set(names.AttrCreateTime, time.Unix(flex.StringValueToInt64Value(v), 0).UTC().Format(time.RFC3339))
	}
	if err := d.Set("entitlements", flattenEntitlements(license.Entitlements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting entitlements: %s", err)
	}
	d.Set("home_region", license.HomeRegion)
	if err := d.Set(names.AttrIssuer, []any{flattenIssuerDetails(license.Issuer)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting issuer: %s", err)
	}
	d.Set("license_arn", license.LicenseArn)
	if err := d.Set("license_metadata", flattenMetadatas(license.LicenseMetadata)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting license_metadata: %s", err)
	}
	d.Set("license_name", license.LicenseName)
	d.Set("product_name", license.ProductName)
	d.Set("product_sku", license.ProductSKU)
	if err := d.Set("received_metadata", []any{flattenReceivedMetadata(license.ReceivedMetadata)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting received_metadata: %s", err)
	}
	d.Set(names.AttrStatus, license.Status)
	if err := d.Set("validity", []any{flattenDateTimeRange(license.Validity)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting validity: %s", err)
	}
	d.Set(names.AttrVersion, license.Version)

	return diags
}

func findReceivedLicenseByARN(ctx context.Context, conn *licensemanager.Client, arn string) (*awstypes.GrantedLicense, error) {
	input := &licensemanager.ListReceivedLicensesInput{
		LicenseArns: []string{arn},
	}

	return findReceivedLicense(ctx, conn, input)
}

func findReceivedLicense(ctx context.Context, conn *licensemanager.Client, input *licensemanager.ListReceivedLicensesInput) (*awstypes.GrantedLicense, error) {
	output, err := findReceivedLicenses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func flattenConsumptionConfiguration(apiObject *awstypes.ConsumptionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.BorrowConfiguration; v != nil {
		tfMap["borrow_configuration"] = map[string]any{
			"allow_early_check_in":        aws.ToBool(v.AllowEarlyCheckIn),
			"max_time_to_live_in_minutes": aws.ToInt32(v.MaxTimeToLiveInMinutes),
		}
	}

	if v := apiObject.ProvisionalConfiguration.MaxTimeToLiveInMinutes; v != nil {
		tfMap["provisional_configuration"] = []any{map[string]any{
			"max_time_to_live_in_minutes": aws.ToInt32(v),
		}}
	}

	tfMap["renew_type"] = apiObject.RenewType

	return tfMap
}

func flattenEntitlements(apiObjects []awstypes.Entitlement) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := flattenEntitlement(&apiObject)

		if len(tfMap) > 0 {
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}

func flattenEntitlement(apiObject *awstypes.Entitlement) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.AllowCheckIn; v != nil {
		tfMap["allow_check_in"] = aws.ToBool(v)
	}

	if v := apiObject.MaxCount; v != nil {
		tfMap["max_count"] = aws.ToInt64(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Overage; v != nil {
		tfMap["overage"] = aws.ToBool(v)
	}

	tfMap[names.AttrUnit] = apiObject.Unit

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenIssuerDetails(apiObject *awstypes.IssuerDetails) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.KeyFingerprint; v != nil {
		tfMap["key_fingerprint"] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.SignKey; v != nil {
		tfMap["sign_key"] = aws.ToString(v)
	}

	return tfMap
}

func flattenMetadatas(apiObjects []awstypes.Metadata) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := flattenMetadata(&apiObject)

		if len(tfMap) > 0 {
			tfList = append(tfList, tfMap)
		}
	}

	return tfList
}

func flattenMetadata(apiObject *awstypes.Metadata) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenReceivedMetadata(apiObject *awstypes.ReceivedMetadata) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.AllowedOperations; v != nil {
		tfMap["allowed_operations"] = v
	}

	tfMap["received_status"] = apiObject.ReceivedStatus

	if v := apiObject.ReceivedStatusReason; v != nil {
		tfMap["received_status_reason"] = aws.ToString(v)
	}

	return tfMap
}

func flattenDateTimeRange(apiObject *awstypes.DatetimeRange) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Begin; v != nil {
		tfMap["begin"] = aws.ToString(v)
	}

	if v := apiObject.End; v != nil {
		tfMap["end"] = aws.ToString(v)
	}

	return tfMap
}
