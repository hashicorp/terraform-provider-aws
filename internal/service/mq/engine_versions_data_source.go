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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_mq_engine_versions", name="Engine Versions")
func DataSourceEngineVersions() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEngineVersionsRead,
		Schema: map[string]*schema.Schema{
			"filters": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"engine_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{"ACTIVEMQ", "RABBITMQ"}, false),
						},
					},
				},
			},
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
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	DSNameEngineVersions = "Engine Versions Data Source"
)

func dataSourceEngineVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*conns.AWSClient).MQClient(ctx)

	input := &mq.DescribeBrokerEngineTypesInput{}
	if v, ok := d.GetOk("filters"); ok {
		filters := v.(*schema.Set).List()
		for _, filter := range filters {
			f := filter.(map[string]interface{})
			if v, ok := f["engine_type"]; ok {
				input.EngineType = aws.String(v.(string))
			}
		}
	}
	d.SetId(id.UniqueId())

	var engineTypes []types.BrokerEngineType
	for {
		out, err := client.DescribeBrokerEngineTypes(ctx, input)
		if err != nil {
			return append(diags, create.DiagError(names.MQ, create.ErrActionReading, DSNameEngineVersions, "", err)...)
		}

		engineTypes = append(engineTypes, out.BrokerEngineTypes...)
		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	if err := d.Set("broker_engine_types", flattenBrokerList(engineTypes)); err != nil {
		return append(diags, create.DiagError(names.MQ, create.ErrActionSetting, DSNameEngineVersions, d.Id(), err)...)
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
			"name": aws.ToString(engine.Name),
		})
	}
	return
}
