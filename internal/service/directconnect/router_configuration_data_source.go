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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_dx_router_configuration", name="Router Configuration")
func dataSourceRouterConfiguration() *schema.Resource {
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

func dataSourceRouterConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	routerTypeIdentifier := d.Get("router_type_identifier").(string)
	vifID := d.Get("virtual_interface_id").(string)
	id := fmt.Sprintf("%s:%s", vifID, routerTypeIdentifier)
	output, err := findRouterConfigurationByTwoPartKey(ctx, conn, routerTypeIdentifier, vifID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Router Configuration (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("customer_router_config", output.CustomerRouterConfig)
	if err := d.Set("router", flattenRouter(output.Router)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting router: %s", err)
	}
	d.Set("router_type_identifier", output.Router.RouterTypeIdentifier)
	d.Set("virtual_interface_id", output.VirtualInterfaceId)
	d.Set("virtual_interface_name", output.VirtualInterfaceName)

	return diags
}

func findRouterConfigurationByTwoPartKey(ctx context.Context, conn *directconnect.Client, routerTypeIdentifier, vifID string) (*directconnect.DescribeRouterConfigurationOutput, error) {
	input := &directconnect.DescribeRouterConfigurationInput{
		RouterTypeIdentifier: aws.String(routerTypeIdentifier),
		VirtualInterfaceId:   aws.String(vifID),
	}

	output, err := conn.DescribeRouterConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Router == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func flattenRouter(apiObject *awstypes.RouterType) []any {
	tfMap := map[string]any{}

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

	return []any{tfMap}
}
