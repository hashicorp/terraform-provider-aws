// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_image")
func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"operating_system_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"required_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	imageID := d.Get("image_id").(string)
	input := &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []string{imageID},
	}

	resp, err := conn.DescribeWorkspaceImages(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describe workspaces images: %s", err)
	}
	if len(resp.Images) == 0 {
		return sdkdiag.AppendErrorf(diags, "Workspace image %s was not found", imageID)
	}

	image := resp.Images[0]
	d.SetId(imageID)
	d.Set(names.AttrName, image.Name)
	d.Set(names.AttrDescription, image.Description)
	d.Set("operating_system_type", image.OperatingSystem.Type)
	d.Set("required_tenancy", image.RequiredTenancy)
	d.Set(names.AttrState, image.State)

	return diags
}
