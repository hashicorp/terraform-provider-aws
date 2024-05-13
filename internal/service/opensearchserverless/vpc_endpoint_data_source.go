// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_opensearchserverless_vpc_endpoint")
func DataSourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCEndpointRead,

		Schema: map[string]*schema.Schema{
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^vpce-[0-9a-z]*$`), `must start with "vpce-" and can include any lower case letter or number`),
				),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroupIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	id := d.Get(names.AttrVPCEndpointID).(string)
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Serverless VPC Endpoint with id (%s): %s", id, err)
	}

	d.SetId(aws.ToString(vpcEndpoint.Id))

	createdDate := time.UnixMilli(aws.ToInt64(vpcEndpoint.CreatedDate))
	d.Set(names.AttrCreatedDate, createdDate.Format(time.RFC3339))

	d.Set(names.AttrName, vpcEndpoint.Name)
	d.Set(names.AttrSecurityGroupIDs, vpcEndpoint.SecurityGroupIds)
	d.Set(names.AttrSubnetIDs, vpcEndpoint.SubnetIds)
	d.Set(names.AttrVPCID, vpcEndpoint.VpcId)

	return diags
}
