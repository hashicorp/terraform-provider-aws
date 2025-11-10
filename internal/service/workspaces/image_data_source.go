// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_workspaces_image", name="Image")
func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
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

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesClient(ctx)

	imageID := d.Get("image_id").(string)
	image, err := findImageByID(ctx, conn, imageID)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("WorkSpaces Image", err))
	}

	d.SetId(imageID)
	d.Set(names.AttrDescription, image.Description)
	d.Set(names.AttrName, image.Name)
	d.Set("operating_system_type", image.OperatingSystem.Type)
	d.Set("required_tenancy", image.RequiredTenancy)
	d.Set(names.AttrState, image.State)

	return diags
}

func findImageByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspaceImage, error) {
	input := &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []string{id},
	}

	return findImage(ctx, conn, input)
}

func findImage(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceImagesInput) (*types.WorkspaceImage, error) {
	output, err := findImages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findImages(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeWorkspaceImagesInput) ([]types.WorkspaceImage, error) {
	var output []types.WorkspaceImage

	err := describeWorkspaceImagesPages(ctx, conn, input, func(page *workspaces.DescribeWorkspaceImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Images...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
