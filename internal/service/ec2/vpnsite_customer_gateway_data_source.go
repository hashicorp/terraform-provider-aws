// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_customer_gateway", name="Customer Gateway")
// @Tags
// @Testing(tagsTest=false)
func dataSourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomerGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"bgp_asn_extended": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrCertificateARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDeviceName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomerGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeCustomerGatewaysInput{}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filters = newCustomFilterList(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.CustomerGatewayIds = []string{v.(string)}
	}

	cgw, err := findCustomerGateway(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Customer Gateway", err))
	}

	d.SetId(aws.ToString(cgw.CustomerGatewayId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if v := aws.ToString(cgw.BgpAsn); v != "" {
		v, err := strconv.ParseInt(v, 0, 0)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("bgp_asn", v)
	} else {
		d.Set("bgp_asn", nil)
	}
	if v := aws.ToString(cgw.BgpAsnExtended); v != "" {
		v, err := strconv.ParseInt(v, 0, 0)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("bgp_asn_extended", v)
	} else {
		d.Set("bgp_asn_extended", nil)
	}
	d.Set(names.AttrCertificateARN, cgw.CertificateArn)
	d.Set(names.AttrDeviceName, cgw.DeviceName)
	d.Set(names.AttrIPAddress, cgw.IpAddress)
	d.Set(names.AttrType, cgw.Type)

	setTagsOut(ctx, cgw.Tags)

	return diags
}
