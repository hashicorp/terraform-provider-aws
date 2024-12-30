// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_mq_broker_engine_types", name="Broker Engine Types")
func dataSourceBrokerEngineTypes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerEngineTypesRead,

		Schema: map[string]*schema.Schema{
			"broker_engine_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"engine_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"engine_versions": {
							Type:     schema.TypeList,
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
					},
				},
			},
			"engine_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.EngineType](),
			},
		},
	}
}

func dataSourceBrokerEngineTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).MQClient(ctx)

	input := &mq.DescribeBrokerEngineTypesInput{}

	if v, ok := d.GetOk("engine_type"); ok {
		input.EngineType = aws.String(v.(string))
	}

	var engineTypes []types.BrokerEngineType
	for {
		output, err := client.DescribeBrokerEngineTypes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading MQ Broker Engine Types: %s", err)
		}

		engineTypes = append(engineTypes, output.BrokerEngineTypes...)

		if output.NextToken == nil {
			break
		}

		input.NextToken = output.NextToken
	}

	d.SetId(id.UniqueId())

	if err := d.Set("broker_engine_types", flattenBrokerList(engineTypes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting broker_engine_types: %s", err)
	}

	return diags
}

func flattenBrokerList(types []types.BrokerEngineType) (brokers []map[string]interface{}) {
	for _, broker := range types {
		brokers = append(brokers, map[string]interface{}{
			"engine_type":     broker.EngineType,
			"engine_versions": flattenEngineVersions(broker.EngineVersions),
		})
	}
	return
}

func flattenEngineVersions(engines []types.EngineVersion) (versions []map[string]string) {
	for _, engine := range engines {
		versions = append(versions, map[string]string{
			names.AttrName: aws.ToString(engine.Name),
		})
	}
	return
}
