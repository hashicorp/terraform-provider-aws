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
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpc_endpoint_service", name="Endpoint Service")
func dataSourceVPCEndpointService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVPCEndpointServiceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrServiceName},
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrServiceName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"service"},
			},
			"service_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ServiceType](),
			},
			"supported_ip_address_types": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"vpc_endpoint_policy_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEndpointServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpcEndpointServicesInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"service-type": d.Get("service_type").(string),
			},
		),
	}

	var serviceName string

	if v, ok := d.GetOk(names.AttrServiceName); ok {
		serviceName = v.(string)
	} else if v, ok := d.GetOk("service"); ok {
		serviceName = fmt.Sprintf("com.amazonaws.%s.%s", meta.(*conns.AWSClient).Region, v.(string))
	}

	if serviceName != "" {
		input.ServiceNames = []string{serviceName}
	}

	if v, ok := d.GetOk(names.AttrTags); ok {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))))...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set))...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	serviceDetails, serviceNames, err := findVPCEndpointServices(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoint Services: %s", err)
	}

	if len(serviceDetails) == 0 && len(serviceNames) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching EC2 VPC Endpoint Service found")
	}

	// Note: AWS Commercial now returns a response with `ServiceNames` and
	// `ServiceDetails`, but GovCloud responses only include `ServiceNames`
	if len(serviceDetails) == 0 {
		// GovCloud doesn't respect the filter.
		for _, name := range serviceNames {
			if name == serviceName {
				d.SetId(strconv.Itoa(create.StringHashcode(name)))
				d.Set(names.AttrServiceName, name)
				return diags
			}
		}

		return sdkdiag.AppendErrorf(diags, "no matching EC2 VPC Endpoint Service found")
	}

	if len(serviceDetails) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple EC2 VPC Endpoint Services matched; use additional constraints to reduce matches to a single EC2 VPC Endpoint Service")
	}

	sd := serviceDetails[0]
	serviceID := aws.ToString(sd.ServiceId)
	serviceName = aws.ToString(sd.ServiceName)

	d.SetId(strconv.Itoa(create.StringHashcode(serviceName)))

	d.Set("acceptance_required", sd.AcceptanceRequired)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-endpoint-service/%s", serviceID),
	}.String()
	d.Set(names.AttrARN, arn)

	d.Set(names.AttrAvailabilityZones, sd.AvailabilityZones)
	d.Set("base_endpoint_dns_names", sd.BaseEndpointDnsNames)
	d.Set("manages_vpc_endpoints", sd.ManagesVpcEndpoints)
	d.Set(names.AttrOwner, sd.Owner)
	d.Set("private_dns_name", sd.PrivateDnsName)
	d.Set("service_id", serviceID)
	d.Set(names.AttrServiceName, serviceName)
	if len(sd.ServiceType) > 0 {
		d.Set("service_type", sd.ServiceType[0].ServiceType)
	} else {
		d.Set("service_type", nil)
	}
	d.Set("supported_ip_address_types", sd.SupportedIpAddressTypes)
	d.Set("vpc_endpoint_policy_supported", sd.VpcEndpointPolicySupported)

	err = d.Set(names.AttrTags, keyValueTags(ctx, sd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
