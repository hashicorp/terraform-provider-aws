// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_service_network")
func DataSourceServiceNetwork() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceNetworkRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_associated_services": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"number_of_associated_vpcs": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_network_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameServiceNetwork = "Service Network Data Source"
)

func dataSourceServiceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceNetworkID := d.Get("service_network_identifier").(string)
	out, err := findServiceNetworkByID(ctx, conn, serviceNetworkID)

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameServiceNetwork, serviceNetworkID, err)
	}

	d.SetId(aws.ToString(out.Id))
	d.Set("arn", out.Arn)
	d.Set("auth_type", out.AuthType)
	d.Set("created_at", aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set("name", out.Name)
	d.Set("number_of_associated_services", out.NumberOfAssociatedServices)
	d.Set("number_of_associated_vpcs", out.NumberOfAssociatedVPCs)
	d.Set("service_network_identifier", out.Id)

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags, err := listTags(ctx, conn, aws.ToString(out.Arn))

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameServiceNetwork, serviceNetworkID, err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, DSNameServiceNetwork, d.Id(), err)
	}

	return nil
}
