package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/appstream/lister"
)

func init() {
	resource.AddTestSweepers("aws_appstream_image_builder", &resource.Sweeper{
		Name: "aws_appstream_image_builder",
		F:    testSweepAppStreamImageBuilder,
		Dependencies: []string{
			"aws_vpc",
			"aws_subnet",
		},
	})
}

func testSweepAppStreamImageBuilder(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).appstreamconn

	var sweeperErrs *multierror.Error

	input := &appstream.DescribeImageBuildersInput{}

	err = lister.DescribeImageBuildersPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.ImageBuilders {
			id := aws.StringValue(directory.Name)

			r := resourceAwsAppStreamImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting AppStream ImageBuilder (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream ImageBuilder sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing AppStream ImageBuilders: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsAppStreamImageBuilder_basic(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfig(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAwsAppStreamImageBuilder_disappears(t *testing.T) {
	var imageBuilderOutput appstream.ImageBuilder
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfig(instanceType, rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
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
				Config: testAccAwsAppStreamImageBuilderConfigComplete(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccAwsAppStreamImageBuilderConfigComplete(rName, descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
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
				Config: testAccAwsAppStreamImageBuilderConfigWithTags(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccAwsAppStreamImageBuilderConfigWithTags(rName, descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName, &imageBuilderOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccAwsAppStreamImageBuilderConfig(instanceType, name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "Amazon-AppStream2-Sample-Image-02-04-2019"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, name)
}

func testAccAwsAppStreamImageBuilderConfigComplete(name, description, instanceType string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_image_builder" "test" {
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
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

func testAccAwsAppStreamImageBuilderConfigWithTags(name, description, instanceType string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_image_builder" "test" {
  image_name                     = "Amazon-AppStream2-Sample-Image-02-04-2019"
  name                           = %[1]q
  description                    = %[2]q
  enable_default_internet_access = false
  instance_type                  = %[3]q
  vpc_config {
    subnet_ids = [aws_subnet.test.id]
  }
  tags = {
    Key = "value"
  }
}
`, name, description, instanceType))
}
