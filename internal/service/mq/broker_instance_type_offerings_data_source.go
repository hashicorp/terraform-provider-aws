// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_mq_broker_instance_type_offerings", name="Broker Instance Type Offerings")
func dataSourceBrokerInstanceTypeOfferings() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerInstanceTypeOfferingsRead,

		Schema: map[string]*schema.Schema{
			"broker_instance_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"engine_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStorageType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"supported_deployment_modes": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"supported_engine_versions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"engine_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.EngineType](),
			},
			"host_instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrStorageType: {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.BrokerStorageType](),
			},
		},
	}
}

func dataSourceBrokerInstanceTypeOfferingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MQClient(ctx)

	input := &mq.DescribeBrokerInstanceOptionsInput{}

	if v, ok := d.GetOk("engine_type"); ok {
		input.EngineType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_instance_type"); ok {
		input.HostInstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrStorageType); ok {
		input.StorageType = aws.String(v.(string))
	}

	var output []types.BrokerInstanceOption

	err := describeBrokerInstanceOptionsPages(ctx, conn, input, func(page *mq.DescribeBrokerInstanceOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.BrokerInstanceOptions...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MQ Broker Instance Options: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("broker_instance_options", flattenBrokerInstanceOptions(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting broker_instance_options: %s", err)
	}

	return diags
}

func flattenBrokerInstanceOptions(bios []types.BrokerInstanceOption) []interface{} {
	if len(bios) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, bio := range bios {
		tfMap := map[string]interface{}{
			"engine_type":                bio.EngineType,
			names.AttrStorageType:        bio.StorageType,
			"supported_deployment_modes": bio.SupportedDeploymentModes,
			"supported_engine_versions":  bio.SupportedEngineVersions,
		}

		if bio.HostInstanceType != nil {
			tfMap["host_instance_type"] = aws.ToString(bio.HostInstanceType)
		}

		if bio.AvailabilityZones != nil {
			tfMap[names.AttrAvailabilityZones] = flattenAvailabilityZones(bio.AvailabilityZones)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAvailabilityZones(azs []types.AvailabilityZone) []interface{} {
	if len(azs) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, az := range azs {
		tfMap := map[string]interface{}{}

		if az.Name != nil {
			tfMap[names.AttrName] = aws.ToString(az.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
