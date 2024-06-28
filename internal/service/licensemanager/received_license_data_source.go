// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/licensemanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
	conn := meta.(*conns.AWSClient).LicenseManagerConn(ctx)

	arn := d.Get("license_arn").(string)

	in := &licensemanager.ListReceivedLicensesInput{
		LicenseArns: aws.StringSlice([]string{arn}),
	}

	out, err := FindReceivedLicenseByARN(ctx, conn, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading License Manager Received License (%s): %s", arn, err)
	}

	d.SetId(aws.StringValue(out.LicenseArn))
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

	if v := aws.StringValue(out.CreateTime); v != "" {
		seconds, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading License Manager Received License (%s): %s", arn, err)
		}
		d.Set(names.AttrCreateTime, time.Unix(seconds, 0).UTC().Format(time.RFC3339))
	}

	return diags
}

func FindReceivedLicenseByARN(ctx context.Context, conn *licensemanager.LicenseManager, in *licensemanager.ListReceivedLicensesInput) (*licensemanager.GrantedLicense, error) {
	out, err := conn.ListReceivedLicensesWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, licensemanager.ErrCodeResourceNotFoundException) {
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
		return nil, create.Error(names.LicenseManager, create.ErrActionReading, ResReceivedLicense, *in.LicenseArns[0], errors.New("More than one License Returned by the API."))
	}

	return out.Licenses[0], nil
}

func flattenConsumptionConfiguration(apiObject *licensemanager.ConsumptionConfiguration) map[string]interface{} {
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
			"max_time_to_live_in_minutes": int(aws.Int64Value(v)),
		}}
	}

	if v := apiObject.RenewType; v != nil {
		tfMap["renew_type"] = v
	}

	return tfMap
}

func flattenEntitlements(apiObjects []*licensemanager.Entitlement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		out := flattenEntitlement(apiObject)

		if len(out) > 0 {
			tfList = append(tfList, out)
		}
	}

	return tfList
}

func flattenEntitlement(apiObject *licensemanager.Entitlement) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

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

	if v := apiObject.Unit; v != nil {
		tfMap[names.AttrUnit] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = v
	}

	return tfMap
}

func flattenIssuer(apiObject *licensemanager.IssuerDetails) map[string]interface{} {
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

func flattenMetadatas(apiObjects []*licensemanager.Metadata) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		out := flattenLicenseMetadata(apiObject)

		if len(out) > 0 {
			tfList = append(tfList, out)
		}
	}

	return tfList
}

func flattenLicenseMetadata(apiObject *licensemanager.Metadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = v
	}

	return tfMap
}

func flattenReceivedMetadata(apiObject *licensemanager.ReceivedMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllowedOperations; v != nil {
		tfMap["allowed_operations"] = v
	}

	if v := apiObject.ReceivedStatus; v != nil {
		tfMap["received_status"] = v
	}

	if v := apiObject.ReceivedStatusReason; v != nil {
		tfMap["received_status_reason"] = v
	}

	return tfMap
}

func flattenDateTimeRange(apiObject *licensemanager.DatetimeRange) map[string]interface{} {
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
