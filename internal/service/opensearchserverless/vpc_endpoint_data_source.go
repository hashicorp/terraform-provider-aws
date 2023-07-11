// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_opensearchserverless_vpc_endpoint")
func DataSourceVPCEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCEndpointRead,

		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^vpce-[0-9a-z]*$`), `must start with "vpce-" and can include any lower case letter or number`),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	id := d.Get("vpc_endpoint_id").(string)
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Serverless VPC Endpoint with id (%s): %s", id, err)
	}

	d.SetId(aws.ToString(vpcEndpoint.Id))

	createdDate := time.UnixMilli(aws.ToInt64(vpcEndpoint.CreatedDate))
	d.Set("created_date", createdDate.Format(time.RFC3339))

	d.Set("name", vpcEndpoint.Name)
	d.Set("security_group_ids", vpcEndpoint.SecurityGroupIds)
	d.Set("subnet_ids", vpcEndpoint.SubnetIds)
	d.Set("vpc_id", vpcEndpoint.VpcId)

	return diags
}
