// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_dx_router_configuration")
func DataSourceRouterConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouterConfigurationRead,

		Schema: map[string]*schema.Schema{
			"customer_router_config": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"router": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"platform": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"router_type_identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"software": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vendor": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"xslt_template_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"xslt_template_name_for_mac_sec": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"router_type_identifier": {
				Type: schema.TypeString,
				// even though the API Reference shows this as optional, the API call will fail without this argument
				Required: true,
			},
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"virtual_interface_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameRouterConfiguration = "Router Configuration Data Source"
)

func dataSourceRouterConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	routerTypeIdentifier := d.Get("router_type_identifier").(string)
	virtualInterfaceId := d.Get("virtual_interface_id").(string)

	out, err := findRouterConfigurationByTypeAndVif(ctx, conn, routerTypeIdentifier, virtualInterfaceId)
	if err != nil {
		return create.AppendDiagError(diags, names.DirectConnect, create.ErrActionReading, DSNameRouterConfiguration, virtualInterfaceId, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", virtualInterfaceId, routerTypeIdentifier))

	d.Set("customer_router_config", out.CustomerRouterConfig)
	d.Set("router_type_identifier", out.Router.RouterTypeIdentifier)
	d.Set("virtual_interface_id", out.VirtualInterfaceId)
	d.Set("virtual_interface_name", out.VirtualInterfaceName)

	if err := d.Set("router", flattenRouter(out.Router)); err != nil {
		return create.AppendDiagError(diags, names.DirectConnect, create.ErrActionSetting, DSNameRouterConfiguration, d.Id(), err)
	}

	return diags
}

func findRouterConfigurationByTypeAndVif(ctx context.Context, conn *directconnect.Client, routerTypeIdentifier string, virtualInterfaceId string) (*directconnect.DescribeRouterConfigurationOutput, error) {
	input := &directconnect.DescribeRouterConfigurationInput{
		RouterTypeIdentifier: aws.String(routerTypeIdentifier),
		VirtualInterfaceId:   aws.String(virtualInterfaceId),
	}

	output, err := conn.DescribeRouterConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func flattenRouter(apiObject *awstypes.RouterType) []interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Platform; v != nil {
		tfMap["platform"] = aws.ToString(v)
	}

	if v := apiObject.RouterTypeIdentifier; v != nil {
		tfMap["router_type_identifier"] = aws.ToString(v)
	}

	if v := apiObject.Software; v != nil {
		tfMap["software"] = aws.ToString(v)
	}

	if v := apiObject.Vendor; v != nil {
		tfMap["vendor"] = aws.ToString(v)
	}

	if v := apiObject.XsltTemplateName; v != nil {
		tfMap["xslt_template_name"] = aws.ToString(v)
	}

	if v := apiObject.XsltTemplateNameForMacSec; v != nil {
		tfMap["xslt_template_name_for_mac_sec"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
