// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_service")
func dataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_entry": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "service_identifier"},
			},
			"service_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "service_identifier"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameService = "Service Data Source"
)

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var out *vpclattice.GetServiceOutput
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if v, ok := d.GetOk("service_identifier"); ok {
		serviceID := v.(string)
		service, err := findServiceByID(ctx, conn, serviceID)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		out = service
	} else if v, ok := d.GetOk("name"); ok {
		filter := func(x types.ServiceSummary) bool {
			return aws.ToString(x.Name) == v.(string)
		}
		output, err := findService(ctx, conn, filter)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		service, err := findServiceByID(ctx, conn, aws.ToString(output.Id))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		out = service
	}

	//
	// If you don't set the ID, the data source will not be stored in state. In
	// fact, that's how a resource can be removed from state - clearing its ID.
	//
	// If this data source is a companion to a resource, often both will use the
	// same ID. Otherwise, the ID will be a unique identifier such as an AWS
	// identifier, ARN, or name.
	d.SetId(aws.ToString(out.Id))

	d.Set("arn", out.Arn)
	d.Set("auth_type", out.AuthType)
	d.Set("certificate_arn", out.CertificateArn)
	d.Set("custom_domain_name", out.CustomDomainName)
	if out.DnsEntry != nil {
		if err := d.Set("dns_entry", []interface{}{flattenDNSEntry(out.DnsEntry)}); err != nil {
			return diag.Errorf("setting dns_entry: %s", err)
		}
	} else {
		d.Set("dns_entry", nil)
	}
	d.Set("name", out.Name)
	d.Set("service_identifier", out.Id)
	d.Set("status", out.Status)

	return nil
}
