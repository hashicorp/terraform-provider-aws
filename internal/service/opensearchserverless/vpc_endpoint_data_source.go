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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_opensearchserverless_vpc_endpoint", name="VPC Endpoint")
func dataSourceVPCEndpoint() *schema.Resource {
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

func dataSourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	id := d.Get(names.AttrVPCEndpointID).(string)
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Serverless VPC Endpoint (%s): %s", id, err)
	}

	d.SetId(aws.ToString(vpcEndpoint.Id))
	createdDate := time.UnixMilli(aws.ToInt64(vpcEndpoint.CreatedDate))
	d.Set(names.AttrCreatedDate, createdDate.Format(time.RFC3339))
	d.Set(names.AttrName, vpcEndpoint.Name)
	d.Set(names.AttrSubnetIDs, vpcEndpoint.SubnetIds)
	d.Set(names.AttrVPCID, vpcEndpoint.VpcId)

	// Security Group IDs are not returned and must be retrieved from the EC2 API.
	vpce, err := tfec2.FindVPCEndpointByID(ctx, meta.(*conns.AWSClient).EC2Client(ctx), d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Serverless VPC Endpoint (%s): %s", id, err)
	}

	var securityGroupIDs []*string
	for _, group := range vpce.Groups {
		securityGroupIDs = append(securityGroupIDs, group.GroupId)
	}

	d.Set(names.AttrSecurityGroupIDs, aws.ToStringSlice(securityGroupIDs))

	return diags
}
