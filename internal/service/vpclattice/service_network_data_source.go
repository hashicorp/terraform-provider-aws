// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_service_network")
// @Tags
func dataSourceServiceNetwork() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceNetworkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
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
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceServiceNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceNetworkID := d.Get("service_network_identifier").(string)
	out, err := findServiceNetworkByID(ctx, conn, serviceNetworkID)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(aws.ToString(out.Id))
	serviceNetworkARN := aws.ToString(out.Arn)
	d.Set(names.AttrARN, serviceNetworkARN)
	d.Set("auth_type", out.AuthType)
	d.Set(names.AttrCreatedAt, aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set(names.AttrName, out.Name)
	d.Set("number_of_associated_services", out.NumberOfAssociatedServices)
	d.Set("number_of_associated_vpcs", out.NumberOfAssociatedVPCs)
	d.Set("service_network_identifier", out.Id)

	// https://docs.aws.amazon.com/vpc-lattice/latest/ug/sharing.html#sharing-perms
	// Owners and consumers can list tags and can tag/untag resources in a service network that the account created.
	// They can't list tags and tag/untag resources in a service network that aren't created by the account.
	parsedARN, err := arn.Parse(serviceNetworkARN)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedARN.AccountID == meta.(*conns.AWSClient).AccountID {
		tags, err := listTags(ctx, conn, serviceNetworkARN)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for VPC Lattice Service Network (%s): %s", serviceNetworkARN, err)
		}

		setTagsOut(ctx, Tags(tags))
	}

	return diags
}
