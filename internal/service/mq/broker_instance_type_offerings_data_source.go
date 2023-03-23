package mq

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceBrokerInstanceTypeOfferings() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBrokerInstanceTypeOfferingsRead,

		Schema: map[string]*schema.Schema{
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
		},
	}
}

func dataSourceBrokerInstanceTypeOfferingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MQConn

	input := &mq.DescribeBrokerInstanceOptionsInput{}

	if v, ok := d.GetOk("host_instance_type"); ok {
		input.HostInstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_type"); ok {
		input.EngineType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_type"); ok {
		input.StorageType = aws.String(v.(string))
	}

	bios := make([]*mq.BrokerInstanceOption, 0)
	for {
		output, err := conn.DescribeBrokerInstanceOptions(input)

		if err != nil {
			return fmt.Errorf("error listing MQ Broker Instance Type Offerings: %w", err)
		}

		if output == nil {
			return fmt.Errorf("empty response while reading MQ Broker Instance Type Offerings")
		}

		bios = append(bios, output.BrokerInstanceOptions...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set("broker_instance_options", flattenBrokerInstanceOptions(bios)); err != nil {
		return fmt.Errorf("error setting broker_instance_options: %w", err)
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
