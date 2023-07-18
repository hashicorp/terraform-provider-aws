// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_vpc_endpoint_service")
func DataSourceVPCEndpointService() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner": {
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
				ConflictsWith: []string{"service_name"},
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"service"},
			},
			"service_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(ec2.ServiceType_Values(), false),
			},
			"supported_ip_address_types": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_endpoint_policy_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCEndpointServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeVpcEndpointServicesInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"service-type": d.Get("service_type").(string),
			},
		),
	}

	var serviceName string

	if v, ok := d.GetOk("service_name"); ok {
		serviceName = v.(string)
	} else if v, ok := d.GetOk("service"); ok {
		serviceName = fmt.Sprintf("com.amazonaws.%s.%s", meta.(*conns.AWSClient).Region, v.(string))
	}

	if serviceName != "" {
		input.ServiceNames = aws.StringSlice([]string{serviceName})
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	serviceDetails, serviceNames, err := FindVPCEndpointServices(ctx, conn, input)

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
				d.Set("service_name", name)
				return diags
			}
		}

		return sdkdiag.AppendErrorf(diags, "no matching EC2 VPC Endpoint Service found")
	}

	if len(serviceDetails) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple EC2 VPC Endpoint Services matched; use additional constraints to reduce matches to a single EC2 VPC Endpoint Service")
	}

	sd := serviceDetails[0]
	serviceID := aws.StringValue(sd.ServiceId)
	serviceName = aws.StringValue(sd.ServiceName)

	d.SetId(strconv.Itoa(create.StringHashcode(serviceName)))

	d.Set("acceptance_required", sd.AcceptanceRequired)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("vpc-endpoint-service/%s", serviceID),
	}.String()
	d.Set("arn", arn)

	d.Set("availability_zones", aws.StringValueSlice(sd.AvailabilityZones))
	d.Set("base_endpoint_dns_names", aws.StringValueSlice(sd.BaseEndpointDnsNames))
	d.Set("manages_vpc_endpoints", sd.ManagesVpcEndpoints)
	d.Set("owner", sd.Owner)
	d.Set("private_dns_name", sd.PrivateDnsName)
	d.Set("service_id", serviceID)
	d.Set("service_name", serviceName)
	if len(sd.ServiceType) > 0 {
		d.Set("service_type", sd.ServiceType[0].ServiceType)
	} else {
		d.Set("service_type", nil)
	}
	d.Set("supported_ip_address_types", aws.StringValueSlice(sd.SupportedIpAddressTypes))
	d.Set("vpc_endpoint_policy_supported", sd.VpcEndpointPolicySupported)

	err = d.Set("tags", KeyValueTags(ctx, sd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
