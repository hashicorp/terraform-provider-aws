// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasConn(ctx)

	serviceName := d.Get("service_name").(string)

	input := &servicequotas.ListServicesInput{}

	var service *servicequotas.ServiceInfo
	err := conn.ListServicesPagesWithContext(ctx, input, func(page *servicequotas.ListServicesOutput, lastPage bool) bool {
		for _, s := range page.Services {
			if aws.StringValue(s.ServiceName) == serviceName {
				service = s
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Services: %s", err)
	}

	if service == nil {
		return sdkdiag.AppendErrorf(diags, "finding Service (%s): no results found", serviceName)
	}

	d.Set("service_code", service.ServiceCode)
	d.Set("service_name", service.ServiceName)
	d.SetId(aws.StringValue(service.ServiceCode))

	return diags
}
