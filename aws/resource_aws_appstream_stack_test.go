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
	resource.AddTestSweepers("aws_appstream_stack", &resource.Sweeper{
		Name: "aws_appstream_stack",
		F:    testSweepAppStreamStack,
	})
}

func testSweepAppStreamStack(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).appstreamconn

	var sweeperErrs *multierror.Error

	input := &appstream.DescribeStacksInput{}

	err = lister.DescribeStacksPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.Stacks {
			id := aws.StringValue(directory.Name)

			r := resourceAwsAppStreamStack()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting AppStream Stack (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Stack sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing AppStream Stacks: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsAppStreamStack_basic(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
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

func TestAccAwsAppStreamStack_disappears(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAppStreamStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamStack_complete(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
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

func TestAccAwsAppStreamStack_withTags(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        testAccErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigWithTags(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRfc3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
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

func testAccCheckAwsAppStreamStackExists(resourceName string, appStreamStack *appstream.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).appstreamconn
		resp, err := conn.DescribeStacks(&appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return fmt.Errorf("problem checking for AppStream Stack existence: %w", err)
		}

		if resp == nil && len(resp.Stacks) == 0 {
			return fmt.Errorf("appstream stack %q does not exist", rs.Primary.ID)
		}

		*appStreamStack = *resp.Stacks[0]

		return nil
	}
}

func testAccCheckAwsAppStreamStackDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appstreamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_stack" {
			continue
		}

		resp, err := conn.DescribeStacks(&appstream.DescribeStacksInput{Names: []*string{aws.String(rs.Primary.ID)}})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking AppStream Stack was destroyed: %w", err)
		}

		if resp != nil && len(resp.Stacks) > 0 {
			return fmt.Errorf("appstream stack %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccAwsAppStreamStackConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}
`, name)
}

func testAccAwsAppStreamStackConfigComplete(name, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  user_settings {
    action     = "FILE_UPLOAD"
    permission = "ENABLED"
  }

  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }

  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }
}
`, name, description)
}

func testAccAwsAppStreamStackConfigWithTags(name, description string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name        = %[1]q
  description = %[2]q

  storage_connectors {
    connector_type = "HOMEFOLDERS"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_FROM_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  user_settings {
    action     = "CLIPBOARD_COPY_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  user_settings {
    action     = "FILE_UPLOAD"
    permission = "DISABLED"
  }

  user_settings {
    action     = "FILE_DOWNLOAD"
    permission = "ENABLED"
  }

  user_settings {
    action     = "PRINTING_TO_LOCAL_DEVICE"
    permission = "ENABLED"
  }

  user_settings {
    action     = "DOMAIN_PASSWORD_SIGNIN"
    permission = "ENABLED"
  }

  application_settings {
    enabled        = true
    settings_group = "SettingsGroup"
  }

  tags = {
    Key = "value"
  }
}
`, name, description)
}
