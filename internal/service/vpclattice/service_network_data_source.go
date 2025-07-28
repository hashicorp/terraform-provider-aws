// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ErrCodeAccessDeniedException = "AccessDeniedException"
)

// Caution: Because of cross account usage, using Tags(identifierAttribute="arn") causes Access Denied
// errors because tags need special handling. See crossAccountSetTags().

// @SDKDataSource("aws_vpclattice_service_network", name="Service Network")
// @Tags
// @Testing(tagsTest=false)
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

func dataSourceServiceNetworkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceNetworkID := d.Get("service_network_identifier").(string)
	output, err := findServiceNetworkByID(ctx, conn, serviceNetworkID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Service Network (%s): %s", serviceNetworkID, err)
	}

	d.SetId(aws.ToString(output.Id))
	serviceNetworkARN := aws.ToString(output.Arn)
	d.Set(names.AttrARN, serviceNetworkARN)
	d.Set("auth_type", output.AuthType)
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(output.LastUpdatedAt).String())
	d.Set(names.AttrName, output.Name)
	d.Set("number_of_associated_services", output.NumberOfAssociatedServices)
	d.Set("number_of_associated_vpcs", output.NumberOfAssociatedVPCs)
	d.Set("service_network_identifier", output.Id)

	return crossAccountSetTags(ctx, conn, diags, serviceNetworkARN, meta.(*conns.AWSClient).AccountID(ctx), "Service Network")
}

func crossAccountSetTags(ctx context.Context, conn *vpclattice.Client, diags diag.Diagnostics, resARN, accountID, resName string) diag.Diagnostics {
	// https://docs.aws.amazon.com/vpc-lattice/latest/ug/sharing.html#sharing-perms
	// Owners and consumers can list tags and can tag/untag resources in a service network that the account created.
	// They can't list tags and tag/untag resources in a service network that aren't created by the account.
	parsedARN, err := arn.Parse(resARN)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if parsedARN.AccountID == accountID {
		tags, err := listTags(ctx, conn, resARN)

		if errs.Contains(err, ErrCodeAccessDeniedException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for VPC Lattice %s (%s): %s", resName, resARN, err)
		}

		setTagsOut(ctx, svcTags(tags))
	}

	return diags
}
