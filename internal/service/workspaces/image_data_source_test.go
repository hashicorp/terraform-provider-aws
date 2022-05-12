package workspaces_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccImageDataSource_basic(t *testing.T) {
	var image workspaces.WorkspaceImage
	imageID := os.Getenv("AWS_WORKSPACES_IMAGE_ID")
	dataSourceName := "data.aws_workspaces_image.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccImagePreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, workspaces.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig(imageID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageExists(dataSourceName, &image),
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

func testAccImageDataSourceConfig(imageID string) string {
	return fmt.Sprintf(`
# TODO: Create aws_workspaces_image resource when API will be provided

data aws_workspaces_image test {
  image_id = %q
}
`, imageID)
}

func testAccCheckImageExists(n string, image *workspaces.WorkspaceImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesConn
		resp, err := conn.DescribeWorkspaceImages(&workspaces.DescribeWorkspaceImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return fmt.Errorf("Failed describe workspaces images: %w", err)
		}
		if resp == nil || len(resp.Images) == 0 || resp.Images[0] == nil {
			return fmt.Errorf("Workspace image %s was not found", rs.Primary.ID)
		}
		if aws.StringValue(resp.Images[0].ImageId) != rs.Primary.ID {
			return fmt.Errorf("Workspace image ID mismatch - existing: %q, state: %q", *resp.Images[0].ImageId, rs.Primary.ID)
		}

		*image = *resp.Images[0]

		return nil
	}
}

func testAccCheckImageAttributes(n string, image *workspaces.WorkspaceImage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if err := resource.TestCheckResourceAttr(n, "id", *image.ImageId)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "name", *image.Name)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "description", *image.Description)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "operating_system_type", *image.OperatingSystem.Type)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "required_tenancy", *image.RequiredTenancy)(s); err != nil {
			return err
		}

		if err := resource.TestCheckResourceAttr(n, "state", *image.State)(s); err != nil {
			return err
		}

		return nil
	}
}
