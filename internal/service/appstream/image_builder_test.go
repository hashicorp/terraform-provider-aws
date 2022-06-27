package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamImageBuilder_basic(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_basic(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

func TestAccAppStreamImageBuilder_disappears(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_basic(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamImageBuilder_complete(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_complete(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
			{
				Config: testAccImageBuilderConfig_complete(rName, descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

func TestAccAppStreamImageBuilder_tags(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
			{
				Config: testAccImageBuilderConfig_tags2(instanceType, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAppStreamImageBuilder_imageARN(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	// imageName selected from the available AWS Managed AppStream 2.0 Base Images
	// Reference: https://docs.aws.amazon.com/appstream2/latest/developerguide/base-image-version-history.html
	imageName := "AppStream-WinServer2012R2-07-19-2021"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_byARN(rName, imageName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "image_arn", "appstream", fmt.Sprintf("image/%s", imageName)),
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

func testAccCheckImageBuilderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

		imageBuilder, err := tfappstream.FindImageBuilderByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if imageBuilder == nil {
			return fmt.Errorf("appstream imageBuilder %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckImageBuilderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_image_builder" {
			continue
		}

		imageBuilder, err := tfappstream.FindImageBuilderByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if imageBuilder != nil {
			return fmt.Errorf("appstream imageBuilder %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccImageBuilderConfig_basic(instanceType, name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2012R2-07-19-2021"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, name)
}

func testAccImageBuilderConfig_complete(name, description, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.1.0.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_image_builder" "test" {
  image_name                     = "AppStream-WinServer2012R2-07-19-2021"
  name                           = %[1]q
  description                    = %[2]q
  enable_default_internet_access = false
  instance_type                  = %[3]q
  vpc_config {
    subnet_ids = [aws_subnet.test.id]
  }
}
`, name, description, instanceType))
}

func testAccImageBuilderConfig_tags1(instanceType, name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2012R2-07-19-2021"
  instance_type = %[1]q
  name          = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, instanceType, name, key, value)
}

func testAccImageBuilderConfig_tags2(instanceType, name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2012R2-07-19-2021"
  instance_type = %[1]q
  name          = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, instanceType, name, key1, value1, key2, value2)
}

func testAccImageBuilderConfig_byARN(rName, imageName, instanceType string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_appstream_image_builder" "test" {
  image_arn     = "arn:${data.aws_partition.current.partition}:appstream:%[1]s::image/%[2]s"
  instance_type = %[3]q
  name          = %[4]q
}
`, acctest.Region(), imageName, instanceType, rName)
}
