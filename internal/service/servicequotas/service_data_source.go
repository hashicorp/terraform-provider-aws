// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicequotas_service")
func DataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			"service_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	serviceName := d.Get(names.AttrServiceName).(string)

	input := &servicequotas.ListServicesInput{}

	var service *types.ServiceInfo
	paginator := servicequotas.NewListServicesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Services: %s", err)
		}

		for _, s := range page.Services {
			s := s
			if aws.ToString(s.ServiceName) == serviceName {
				service = &s
				break
			}
		}

		if service != nil {
			break // stop paging once found
		}
	}

	if service == nil {
		return sdkdiag.AppendErrorf(diags, "finding Service (%s): no results found", serviceName)
	}

	d.Set("service_code", service.ServiceCode)
	d.Set(names.AttrServiceName, service.ServiceName)
	d.SetId(aws.ToString(service.ServiceCode))

	return diags
}
