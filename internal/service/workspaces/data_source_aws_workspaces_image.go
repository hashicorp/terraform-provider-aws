package workspaces

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceImageRead,

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

func dataSourceImageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn

	imageID := d.Get("image_id").(string)
	input := &workspaces.DescribeWorkspaceImagesInput{
		ImageIds: []*string{aws.String(imageID)},
	}

	resp, err := conn.DescribeWorkspaceImages(input)
	if err != nil {
		return fmt.Errorf("Failed describe workspaces images: %w", err)
	}
	if len(resp.Images) == 0 {
		return fmt.Errorf("Workspace image %s was not found", imageID)
	}

	image := resp.Images[0]
	d.SetId(imageID)
	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("operating_system_type", image.OperatingSystem.Type)
	d.Set("required_tenancy", image.RequiredTenancy)
	d.Set("state", image.State)

	return nil
}
