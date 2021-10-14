package appstream_test

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
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_appstream_image_builder", &resource.Sweeper{
		Name: "aws_appstream_image_builder",
		F:    sweepImageBuilder,
		Dependencies: []string{
			"aws_vpc",
			"aws_subnet",
		},
	})
}

func sweepImageBuilder(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeImageBuildersInput{}

	err = tfappstream.DescribeImageBuildersPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageBuilder := range page.ImageBuilders {
			if imageBuilder == nil {
				continue
			}

			id := aws.StringValue(imageBuilder.Name)

			r := tfappstream.ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Image Builders: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Image Builders for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Image Builders sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppStreamImageBuilder_basic(t *testing.T) {
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
				Config: testAccImageBuilderConfig(instanceType, rName),
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

func TestAccAwsAppStreamImageBuilder_disappears(t *testing.T) {
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
				Config: testAccImageBuilderConfig(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamImageBuilder_complete(t *testing.T) {
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
				Config: testAccImageBuilderCompleteConfig(rName, description, instanceType),
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
				Config: testAccImageBuilderCompleteConfig(rName, descriptionUpdated, instanceTypeUpdate),
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

func TestAccAwsAppStreamImageBuilder_Tags(t *testing.T) {
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
				Config: testAccImageBuilderTags1Config(instanceType, rName, "key1", "value1"),
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
				Config: testAccImageBuilderTags2Config(instanceType, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImageBuilderTags1Config(instanceType, rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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

func testAccImageBuilderConfig(instanceType, name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2012R2-07-19-2021"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, name)
}

func testAccImageBuilderCompleteConfig(name, description, instanceType string) string {
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

func testAccImageBuilderTags1Config(instanceType, name, key, value string) string {
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

func testAccImageBuilderTags2Config(instanceType, name, key1, value1, key2, value2 string) string {
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
