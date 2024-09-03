// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var image types.WorkspaceImage
	imageID := os.Getenv("AWS_WORKSPACES_IMAGE_ID")
	dataSourceName := "data.aws_workspaces_image.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccImagePreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_basic(imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageExists(ctx, dataSourceName, &image),
					testAccCheckImageAttributes(dataSourceName, &image),
				),
			},
		},
	})
}

func testAccImagePreCheck(t *testing.T) {
	if os.Getenv("AWS_WORKSPACES_IMAGE_ID") == "" {
		t.Skip("AWS_WORKSPACES_IMAGE_ID env var must be set for AWS WorkSpaces image acceptance tests. This is required until AWS provides ubiquitous (Windows, Linux) import image API.")
	}
}

func testAccImageDataSourceConfig_basic(imageID string) string {
	return fmt.Sprintf(`
# TODO: Create aws_workspaces_image resource when API will be provided

data aws_workspaces_image test {
  image_id = %q
}
`, imageID)
}

func testAccCheckImageExists(ctx context.Context, n string, image *types.WorkspaceImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)
		resp, err := conn.DescribeWorkspaceImages(ctx, &workspaces.DescribeWorkspaceImagesInput{
			ImageIds: []string{rs.Primary.ID},
		})
		if err != nil {
			return fmt.Errorf("Failed describe workspaces images: %w", err)
		}
		if resp == nil || len(resp.Images) == 0 || reflect.DeepEqual(resp.Images[0], (types.WorkspaceImage{})) {
			return fmt.Errorf("Workspace image %s was not found", rs.Primary.ID)
		}
		if aws.ToString(resp.Images[0].ImageId) != rs.Primary.ID {
			return fmt.Errorf("Workspace image ID mismatch - existing: %q, state: %q", *resp.Images[0].ImageId, rs.Primary.ID)
		}

		*image = resp.Images[0]

		return nil
	}
}

func testAccCheckImageAttributes(n string, image *types.WorkspaceImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if err := resource.TestCheckResourceAttr(n, names.AttrID, *image.ImageId)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, names.AttrName, *image.Name)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, names.AttrDescription, *image.Description)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "operating_system_type", string(image.OperatingSystem.Type))(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "required_tenancy", string(image.RequiredTenancy))(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, names.AttrState, string(image.State))(s); err != nil {
			return err
		}

		return nil
	}
}
