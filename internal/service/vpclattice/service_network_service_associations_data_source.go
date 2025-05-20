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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_vpclattice_service_network_service_associations", name="Service Network Service Associations")
func DataSourceServiceNetworkServiceAssociations() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceNetworkServiceAssociationsRead,
		Schema: map[string]*schema.Schema{
			"service_network_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"service_network_identifier", "service_identifier"},
			},
			"service_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"service_network_identifier", "service_identifier"},
			},
			"associations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_by": {
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
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrServiceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_network_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_network_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_network_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const (
	DSNameServiceNetworkServiceAssociations = "Service Network Service Associations Data Source"
)

func dataSourceServiceNetworkServiceAssociationsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	if v, ok := d.GetOk("service_network_identifier"); ok {
		service_network_identifier := v.(string)

		d.SetId(v.(string))

		// Checking if the Service Network exists, since list-service-network-service-associations returns an empty list if the Service Network is not found
		if _, err := findServiceNetworkByID(ctx, conn, service_network_identifier); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

		assoc, err := listServiceNetworkServiceAssociationsByServiceNetworkIdentifier(ctx, conn, service_network_identifier)

		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

		if err := d.Set("associations", flattenAssociations(assoc)); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

	} else if v, ok := d.GetOk("service_identifier"); ok {
		service_identifier := v.(string)

		d.SetId(v.(string))

		// Checking if the Service exists, since list-service-network-service-associations returns an empty list if the Service Network is not found
		if _, err := findServiceByID(ctx, conn, service_identifier); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

		assoc, err := listServiceNetworkServiceAssociationsByServiceIdentifier(ctx, conn, service_identifier)

		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

		if err := d.Set("associations", flattenAssociations(assoc)); err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, DSNameServiceNetworkServiceAssociations, d.Id(), err)
		}

	}

	return diags
}

func listServiceNetworkServiceAssociationsByServiceIdentifier(ctx context.Context, conn *vpclattice.Client, ServiceIdentifier string) ([]*types.ServiceNetworkServiceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkServiceAssociationsInput{
		ServiceIdentifier: aws.String(ServiceIdentifier),
	}

	return listServiceNetworkServiceAssociations(ctx, conn, &input)
}

func listServiceNetworkServiceAssociationsByServiceNetworkIdentifier(ctx context.Context, conn *vpclattice.Client, ServiceNetworkIdentifier string) ([]*types.ServiceNetworkServiceAssociationSummary, error) {
	input := vpclattice.ListServiceNetworkServiceAssociationsInput{
		ServiceNetworkIdentifier: aws.String(ServiceNetworkIdentifier),
	}

	return listServiceNetworkServiceAssociations(ctx, conn, &input)
}

func listServiceNetworkServiceAssociations(ctx context.Context, conn *vpclattice.Client, input *vpclattice.ListServiceNetworkServiceAssociationsInput) ([]*types.ServiceNetworkServiceAssociationSummary, error) {
	var output []*types.ServiceNetworkServiceAssociationSummary
	paginator := vpclattice.NewListServiceNetworkServiceAssociationsPaginator(conn, input, func(options *vpclattice.ListServiceNetworkServiceAssociationsPaginatorOptions) {
		options.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}
		for _, v := range page.Items {
			output = append(output, &v)
		}
	}

	return output, nil
}

func flattenAssociation(apiObject *types.ServiceNetworkServiceAssociationSummary) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}
	if v := apiObject.CreatedBy; v != nil {
		tfMap["created_by"] = aws.ToString(v)
	}
	if v := apiObject.CustomDomainName; v != nil {
		tfMap["custom_domain_name"] = aws.ToString(v)
	}
	if v := apiObject.DnsEntry; v != nil {
		tfMap["dns_entry"] = []any{flattenDNSEntry(v)}
	}
	if v := apiObject.Id; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}
	if v := apiObject.ServiceArn; v != nil {
		tfMap["service_arn"] = aws.ToString(v)
	}
	if v := apiObject.ServiceId; v != nil {
		tfMap["service_id"] = aws.ToString(v)
	}
	if v := apiObject.ServiceName; v != nil {
		tfMap[names.AttrServiceName] = aws.ToString(v)
	}
	if v := apiObject.ServiceNetworkArn; v != nil {
		tfMap["service_network_arn"] = aws.ToString(v)
	}
	if v := apiObject.ServiceNetworkId; v != nil {
		tfMap["service_network_id"] = aws.ToString(v)
	}
	if v := apiObject.ServiceNetworkName; v != nil {
		tfMap["service_network_name"] = aws.ToString(v)
	}

	tfMap[names.AttrStatus] = apiObject.Status

	return tfMap
}

func flattenAssociations(apiObjects []*types.ServiceNetworkServiceAssociationSummary) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenAssociation(apiObject))
	}

	return tfList
}
