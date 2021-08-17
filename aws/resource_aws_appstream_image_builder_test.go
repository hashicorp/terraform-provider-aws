package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAwsAppStreamImageBuilder_basic(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_nameGenerated(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_namePrefix(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	namePrefix := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigNamePrefix(instanceType, namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", namePrefix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", namePrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_disappears(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigNameGenerated(instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppStreamImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_complete(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigComplete(description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamImageBuilderConfigComplete(descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_withTags(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigWithTags(description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamImageBuilderConfigWithTags(descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", "value"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
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

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.ImageBuilders) > 0 {
			return fmt.Errorf("appstream imageBuilder %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamImageBuilderConfigNameGenerated(instanceType string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = %[1]q
}
`, instanceType)
}

func testAccAwsAppStreamImageBuilderConfigNamePrefix(instanceType, namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = %[1]q
  name_prefix   = %[2]q
}
`, instanceType, namePrefix)
}

func testAccAwsAppStreamImageBuilderConfigComplete(description, instanceType string) string {
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
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  description                    = %[1]q
  enable_default_internet_access = false
  instance_type                  = %[2]q
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
}
`, description, instanceType)
}

func testAccAwsAppStreamImageBuilderConfigWithTags(description, instanceType string) string {
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
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  description                    = %[1]q
  enable_default_internet_access = false
  instance_type                  = %[2]q
  vpc_config {
    subnet_ids = [aws_subnet.example.id]
  }
  tags = {
    Key = "value"
  }
}
`, description, instanceType)
}
