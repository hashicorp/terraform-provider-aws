package sfn_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSFNActivity_basic(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActivityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
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

func TestAccSFNActivity_tags(t *testing.T) {
	name := sdkacctest.RandString(10)
	resourceName := "aws_sfn_activity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sfn.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActivityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActivityConfig_basicTags1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActivityConfig_basicTags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccActivityConfig_basicTags1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActivityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckActivityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SFNConn

		_, err := conn.DescribeActivity(&sfn.DescribeActivityInput{
			ActivityArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckActivityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SFNConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_activity" {
			continue
		}

		// Retrying as Read after Delete is not always consistent
		retryErr := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error

			_, err = conn.DescribeActivity(&sfn.DescribeActivityInput{
				ActivityArn: aws.String(rs.Primary.ID),
			})

			if tfawserr.ErrCodeEquals(err, sfn.ErrCodeActivityDoesNotExist) {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			// If there are no errors, the removal failed
			// and the object is not yet removed.
			return resource.RetryableError(fmt.Errorf("Expected AWS Step Function Activity to be destroyed, but was still found, retrying"))
		})

		return retryErr
	}

	return fmt.Errorf("Default error in Step Function Test")
}

func testAccActivityConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = "%s"
}
`, rName)
}

func testAccActivityConfig_basicTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccActivityConfig_basicTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "test" {
  name = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
