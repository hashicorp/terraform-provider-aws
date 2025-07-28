// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_servicequotas_service", name="Service")
func dataSourceService() *schema.Resource {
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

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceQuotasClient(ctx)

	service, err := findServiceByName(ctx, conn, d.Get(names.AttrServiceName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Service Quotas Service", err))
	}

	serviceCode := aws.ToString(service.ServiceCode)
	d.SetId(serviceCode)
	d.Set("service_code", serviceCode)
	d.Set(names.AttrServiceName, service.ServiceName)

	return diags
}

func findServiceByName(ctx context.Context, conn *servicequotas.Client, name string) (*awstypes.ServiceInfo, error) {
	input := servicequotas.ListServicesInput{}

	return findService(ctx, conn, &input, func(v *awstypes.ServiceInfo) bool {
		return aws.ToString(v.ServiceName) == name
	})
}

func findService(ctx context.Context, conn *servicequotas.Client, input *servicequotas.ListServicesInput, filter tfslices.Predicate[*awstypes.ServiceInfo]) (*awstypes.ServiceInfo, error) {
	output, err := findServices(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findServices(ctx context.Context, conn *servicequotas.Client, input *servicequotas.ListServicesInput, filter tfslices.Predicate[*awstypes.ServiceInfo]) ([]awstypes.ServiceInfo, error) {
	var output []awstypes.ServiceInfo

	pages := servicequotas.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Services {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
