package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsAppStreamImageBuilder_basic(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	imageBuilderName := acctest.RandomWithPrefix("tf-acc-test")
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigBasic(imageBuilderName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
				),
			},
			{
				Config:            testAccAwsAppStreamImageBuilderConfigBasic(imageBuilderName, instanceType),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsAppStreamImageBuilder_disappears(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	imageBuilderName := acctest.RandomWithPrefix("tf-acc-test")
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigBasic(imageBuilderName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppstreamImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsAppStreamImageBuilder_withTags(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	imageBuilderName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a imageBuilder"
	instanceType := "stream.standard.small"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigWithTags(imageBuilderName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsAppStreamImageBuilderExists(resourceName string, appStreamImageBuilder *appstream.ImageBuilder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn
		resp, err := conn.DescribeImageBuilders(&appstream.DescribeImageBuildersInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.ImageBuilders) == 0 {
			return fmt.Errorf("appstream imageBuilder %q does not exist", rs.Primary.ID)
		}

		*appStreamImageBuilder = *resp.ImageBuilders[0]

		return nil
	}
}

func testAccCheckAwsAppStreamImageBuilderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_image_builder" {
			continue
		}

		resp, err := conn.DescribeImageBuilders(&appstream.DescribeImageBuildersInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp != nil && len(resp.ImageBuilders) > 0 {
			return fmt.Errorf("appstream imageBuilder %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamImageBuilderConfigBasic(imageBuilderName, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  name          = %[1]q
  instance_type = %[2]q
}
`, imageBuilderName, instanceType)
}

func testAccAwsAppStreamImageBuilderConfigWithTags(imageBuilderName, description, instanceType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.example.id
}

resource "aws_appstream_image_builder" "test" {
  name                           = %[1]q
  access_endpoints {
    endpoint_type = "STREAMING"
  }
  description                    = %[2]q
  display_name                   = %[1]q
  enable_default_internet_access = false
  instance_type                  = %[3]q
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
  tags = {
    Key = "value"
  }
}
`, imageBuilderName, description, instanceType)
}
