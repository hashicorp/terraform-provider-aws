// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKDataSource("aws_mq_broker_instance_type_offerings")
func DataSourceBrokerInstanceTypeOfferings() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerInstanceTypeOfferingsRead,

		Schema: map[string]*schema.Schema{
			"broker_instance_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
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
						"storage_type": {
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(mq.EngineType_Values(), false),
			},
			"host_instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"storage_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(mq.BrokerStorageType_Values(), false),
			},
		},
	}
}

func dataSourceBrokerInstanceTypeOfferingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MQConn(ctx)

	input := &mq.DescribeBrokerInstanceOptionsInput{}

	if v, ok := d.GetOk("engine_type"); ok {
		input.EngineType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("host_instance_type"); ok {
		input.HostInstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_type"); ok {
		input.StorageType = aws.String(v.(string))
	}

	var output []*mq.BrokerInstanceOption

	err := describeBrokerInstanceOptionsPages(ctx, conn, input, func(page *mq.DescribeBrokerInstanceOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.BrokerInstanceOptions...)

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("reading MQ Broker Instance Options: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("broker_instance_options", flattenBrokerInstanceOptions(output)); err != nil {
		return diag.Errorf("setting broker_instance_options: %s", err)
	}

	return nil
}

func flattenBrokerInstanceOptions(bios []*mq.BrokerInstanceOption) []interface{} {
	if len(bios) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, bio := range bios {
		if bio == nil {
			continue
		}

		tfMap := map[string]interface{}{}

		if bio.EngineType != nil {
			tfMap["engine_type"] = aws.StringValue(bio.EngineType)
		}

		if bio.StorageType != nil {
			tfMap["storage_type"] = aws.StringValue(bio.StorageType)
		}

		if bio.HostInstanceType != nil {
			tfMap["host_instance_type"] = aws.StringValue(bio.HostInstanceType)
		}

		if bio.AvailabilityZones != nil {
			tfMap["availability_zones"] = flattenAvailabilityZones(bio.AvailabilityZones)
		}

		if bio.SupportedDeploymentModes != nil {
			tfMap["supported_deployment_modes"] = flex.FlattenStringSet(bio.SupportedDeploymentModes)
		}

		if bio.SupportedEngineVersions != nil {
			tfMap["supported_engine_versions"] = flex.FlattenStringList(bio.SupportedEngineVersions)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAvailabilityZones(azs []*mq.AvailabilityZone) []interface{} {
	if len(azs) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, az := range azs {
		if az == nil {
			continue
		}

		tfMap := map[string]interface{}{}

		if az.Name != nil {
			tfMap["name"] = aws.StringValue(az.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
