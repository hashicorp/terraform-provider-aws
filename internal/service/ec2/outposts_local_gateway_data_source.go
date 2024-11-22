// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_local_gateway", name="Local Gateway")
// @Tags
// @Testing(tagsTest=false)
func dataSourceLocalGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceLocalGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLocalGatewaysInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.LocalGatewayIds = []string{v.(string)}
	}

	input.Filters = newAttributeFilterList(
		map[string]string{
			names.AttrState: d.Get(names.AttrState).(string),
		},
	)

	if tags, tagsOk := d.GetOk(names.AttrTags); tagsOk {
		input.Filters = append(input.Filters, newTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	localGateway, err := findLocalGateway(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Local Gateway", err))
	}

	d.SetId(aws.ToString(localGateway.LocalGatewayId))
	d.Set("outpost_arn", localGateway.OutpostArn)
	d.Set(names.AttrOwnerID, localGateway.OwnerId)
	d.Set(names.AttrState, localGateway.State)

	setTagsOut(ctx, localGateway.Tags)

	return diags
}
