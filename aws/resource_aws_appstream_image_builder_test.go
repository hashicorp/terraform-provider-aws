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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/appstream/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/appstream/lister"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeImageBuildersInput{}

	err = lister.DescribeImageBuildersPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageBuilder := range page.ImageBuilders {
			if imageBuilder == nil {
				continue
			}

			id := aws.StringValue(imageBuilder.Name)

			r := resourceAwsAppStreamImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Image Builders: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Image Builders for %s: %w", region, err))
	}

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Image Builders sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppStreamImageBuilder_basic(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfig(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
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

func TestAccAwsAppStreamImageBuilder_disappears(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfig(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsAppStreamImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_complete(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigComplete(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
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
				Config: testAccAwsAppStreamImageBuilderConfigComplete(rName, descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
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

func TestAccAwsAppStreamImageBuilder_Tags(t *testing.T) {
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamImageBuilderDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamImageBuilderConfigTags1(instanceType, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
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
				Config: testAccAwsAppStreamImageBuilderConfigTags2(instanceType, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsAppStreamImageBuilderConfigTags1(instanceType, rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsAppStreamImageBuilderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn

		imageBuilder, err := finder.ImageBuilderByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if imageBuilder == nil {
			return fmt.Errorf("appstream imageBuilder %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsAppStreamImageBuilderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_image_builder" {
			continue
		}

		imageBuilder, err := finder.ImageBuilderByName(context.Background(), conn, rs.Primary.ID)

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

func testAccAwsAppStreamImageBuilderConfig(instanceType, name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2012R2-07-19-2021"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, name)
}

func testAccAwsAppStreamImageBuilderConfigComplete(name, description, instanceType string) string {
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

func testAccAwsAppStreamImageBuilderConfigTags1(instanceType, name, key, value string) string {
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

func testAccAwsAppStreamImageBuilderConfigTags2(instanceType, name, key1, value1, key2, value2 string) string {
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
