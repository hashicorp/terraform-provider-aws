// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_instance_types", name="Instance Types")
func dataSourceInstanceTypes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceTypesRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			"instance_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeInstanceTypesInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	output, err := findInstanceTypes(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instance Types: %s", err)
	}

	var instanceTypes []string

	for _, instanceType := range output {
		instanceTypes = append(instanceTypes, string(instanceType.InstanceType))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("instance_types", instanceTypes)

	return diags
}
