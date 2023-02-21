package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMConn()

	input := &ssm.DescribeInstanceInformationInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = expandInstanceInformationStringFilters(v.(*schema.Set).List())
	}

	var results []*ssm.InstanceInformation

	err := conn.DescribeInstanceInformationPagesWithContext(ctx, input, func(page *ssm.DescribeInstanceInformationOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceInformation := range page.InstanceInformationList {
			if instanceInformation == nil {
				continue
			}

			results = append(results, instanceInformation)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSM Instances: %s", err)
	}

	var instanceIDs []string

	for _, r := range results {
		instanceIDs = append(instanceIDs, aws.StringValue(r.InstanceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", instanceIDs)

	return diags
}

func expandInstanceInformationStringFilters(tfList []interface{}) []*ssm.InstanceInformationStringFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ssm.InstanceInformationStringFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandInstanceInformationStringFilter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInstanceInformationStringFilter(tfMap map[string]interface{}) *ssm.InstanceInformationStringFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &ssm.InstanceInformationStringFilter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringList(v)
	}

	return apiObject
}
