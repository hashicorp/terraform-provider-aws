// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/licensemanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResReceivedLicense = "Received License"
)

// @SDKDataSource("aws_licensemanager_received_license")
func DataSourceReceivedLicense() *schema.Resource {
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

func dataSourceReceivedLicenseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LicenseManagerClient(ctx)

	arn := d.Get("license_arn").(string)

	in := &licensemanager.ListReceivedLicensesInput{
		LicenseArns: []string{arn},
	}

	out, err := FindReceivedLicenseByARN(ctx, conn, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Received License (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(out.LicenseArn))
	d.Set("beneficiary", out.Beneficiary)
	d.Set("consumption_configuration", []interface{}{flattenConsumptionConfiguration(out.ConsumptionConfiguration)})
	d.Set("entitlements", flattenEntitlements(out.Entitlements))
	d.Set("home_region", out.HomeRegion)
	d.Set(names.AttrIssuer, []interface{}{flattenIssuer(out.Issuer)})
	d.Set("license_arn", out.LicenseArn)
	d.Set("license_metadata", flattenMetadatas(out.LicenseMetadata))
	d.Set("license_name", out.LicenseName)
	d.Set("product_name", out.ProductName)
	d.Set("product_sku", out.ProductSKU)
	d.Set("received_metadata", []interface{}{flattenReceivedMetadata(out.ReceivedMetadata)})
	d.Set(names.AttrStatus, out.Status)
	d.Set("validity", []interface{}{flattenDateTimeRange(out.Validity)})
	d.Set(names.AttrVersion, out.Version)

	if v := aws.ToString(out.CreateTime); v != "" {
		seconds, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading License Manager Received License (%s): %s", arn, err)
		}
		d.Set(names.AttrCreateTime, time.Unix(seconds, 0).UTC().Format(time.RFC3339))
	}

	return diags
}

func FindReceivedLicenseByARN(ctx context.Context, conn *licensemanager.Client, in *licensemanager.ListReceivedLicensesInput) (*awstypes.GrantedLicense, error) {
	out, err := conn.ListReceivedLicenses(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(out.Licenses) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if len(out.Licenses) > 1 {
		return nil, create.Error(names.LicenseManager, create.ErrActionReading, ResReceivedLicense, in.LicenseArns[0], errors.New("More than one License Returned by the API."))
	}

	return &out.Licenses[0], nil
}

func flattenConsumptionConfiguration(apiObject *awstypes.ConsumptionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BorrowConfiguration; v != nil {
		tfMap["borrow_configuration"] = map[string]interface{}{
			"max_time_to_live_in_minutes": v.MaxTimeToLiveInMinutes,
			"allow_early_check_in":        v.AllowEarlyCheckIn,
		}
	}

	if v := apiObject.ProvisionalConfiguration.MaxTimeToLiveInMinutes; v != nil {
		tfMap["provisional_configuration"] = []interface{}{map[string]interface{}{
			"max_time_to_live_in_minutes": int(aws.ToInt32(v)),
		}}
	}

	tfMap["renew_type"] = apiObject.RenewType

	return tfMap
}

func flattenEntitlements(apiObjects []awstypes.Entitlement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		out := flattenEntitlement(apiObject)

		if len(out) > 0 {
			tfList = append(tfList, out)
		}
	}

	return tfList
}

func flattenEntitlement(apiObject awstypes.Entitlement) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AllowCheckIn; v != nil {
		tfMap["allow_check_in"] = v
	}

	if v := apiObject.MaxCount; v != nil {
		tfMap["max_count"] = v
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Overage; v != nil {
		tfMap["overage"] = v
	}

	tfMap[names.AttrUnit] = apiObject.Unit

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = v
	}

	return tfMap
}

func flattenIssuer(apiObject *awstypes.IssuerDetails) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.KeyFingerprint; v != nil {
		tfMap["key_fingerprint"] = v
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.SignKey; v != nil {
		tfMap["sign_key"] = v
	}

	return tfMap
}

func flattenMetadatas(apiObjects []awstypes.Metadata) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		out := flattenLicenseMetadata(apiObject)

		if len(out) > 0 {
			tfList = append(tfList, out)
		}
	}

	return tfList
}

func flattenLicenseMetadata(apiObject awstypes.Metadata) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = v
	}

	return tfMap
}

func flattenReceivedMetadata(apiObject *awstypes.ReceivedMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowedOperations; v != nil {
		tfMap["allowed_operations"] = v
	}

	tfMap["received_status"] = apiObject.ReceivedStatus

	if v := apiObject.ReceivedStatusReason; v != nil {
		tfMap["received_status_reason"] = v
	}

	return tfMap
}

func flattenDateTimeRange(apiObject *awstypes.DatetimeRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Begin; v != nil {
		tfMap["begin"] = v
	}

	if v := apiObject.End; v != nil {
		tfMap["end"] = v
	}

	return tfMap
}
