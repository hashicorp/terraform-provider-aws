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
	resource.AddTestSweepers("aws_appstream_stack", &resource.Sweeper{
		Name: "aws_appstream_stack",
		F:    testSweepAppStreamStack,
	})
}

func testSweepAppStreamStack(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppStreamConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &appstream.DescribeStacksInput{}

	err = tfappstream.DescribeStacksPagesWithContext(context.TODO(), conn, input, func(page *appstream.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, stack := range page.Stacks {
			if stack == nil {
				continue
			}

			id := aws.StringValue(stack.Name)

			r := ResourceImageBuilder()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppStream Stacks: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppStream Stacks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppStream Stacks sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func TestAccAwsAppStreamStack_basic(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAppStreamStack_complete(t *testing.T) {
	var stackOutput appstream.Stack
	resourceName := "aws_appstream_stack.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsAppStreamStackDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAppStreamStackConfigComplete(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				Config: testAccAwsAppStreamStackConfigWithTags(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAppStreamStackExists(resourceName, &stackOutput),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

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
