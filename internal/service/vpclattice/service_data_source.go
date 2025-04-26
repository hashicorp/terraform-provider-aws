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

// Caution: Because of cross account usage, using Tags(identifierAttribute="arn") causes Access Denied
// errors because tags need special handling. See crossAccountSetTags().

// @SDKDataSource("aws_vpclattice_service", name="Service")
// @Tags
// @Testing(tagsTest=false)
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
			names.AttrCertificateARN: {
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
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "service_identifier"},
			},
			"service_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "service_identifier"},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	var out *vpclattice.GetServiceOutput
	if v, ok := d.GetOk("service_identifier"); ok {
		serviceID := v.(string)
		service, err := findServiceByID(ctx, conn, serviceID)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		out = service
	} else if v, ok := d.GetOk(names.AttrName); ok {
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

	d.SetId(aws.ToString(out.Id))
	serviceARN := aws.ToString(out.Arn)
	d.Set(names.AttrARN, serviceARN)
	d.Set("auth_type", out.AuthType)
	d.Set(names.AttrCertificateARN, out.CertificateArn)
	d.Set("custom_domain_name", out.CustomDomainName)
	if out.DnsEntry != nil {
		if err := d.Set("dns_entry", []any{flattenDNSEntry(out.DnsEntry)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dns_entry: %s", err)
		}
	} else {
		d.Set("dns_entry", nil)
	}
	d.Set(names.AttrName, out.Name)
	d.Set("service_identifier", out.Id)
	d.Set(names.AttrStatus, out.Status)

	return crossAccountSetTags(ctx, conn, diags, serviceARN, meta.(*conns.AWSClient).AccountID(ctx), "Service")
}
