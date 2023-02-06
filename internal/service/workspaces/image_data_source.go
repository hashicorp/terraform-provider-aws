package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
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
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WorkSpacesConn()

	imageID := d.Get("image_id").(string)
	input := &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []*string{aws.String(imageID)},
	}

	resp, err := conn.DescribeWorkspaceImagesWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describe workspaces images: %s", err)
	}
	if len(resp.Images) == 0 {
		return sdkdiag.AppendErrorf(diags, "Workspace image %s was not found", imageID)
	}

	image := resp.Images[0]
	d.SetId(imageID)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("operating_system_type", image.OperatingSystem.Type)
	d.Set("required_tenancy", image.RequiredTenancy)
	d.Set("state", image.State)

	return diags
}
