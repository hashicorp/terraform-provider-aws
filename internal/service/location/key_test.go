package location_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"testing"
)

func TestAccLocationKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_location_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, locationservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "create_time"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "key_arn", "geo", fmt.Sprintf("api-key/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "update_time"),
					resource.TestCheckResourceAttr(resourceName, "no_expiry", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

// Additional test functions (e.g., TestAccLocationKey_disappears, TestAccLocationKey_description, TestAccLocationKey_tags) would follow a similar pattern.

func testAccCheckKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_location_key" {
				continue
			}

			input := &locationservice.DescribeKeyInput{
				KeyName: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeKeyWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Location Service Key (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Location Service Key (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckKeyExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LocationConn(ctx)

		input := &locationservice.DescribeKeyInput{
			KeyName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeKeyWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Location Service Key (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

resource "aws_location_key" "test" {
  key_name = %[1]q
  no_expiry = true

  restrictions {
    allow_actions = [
      "geo:SearchPlaceIndexForText",
    ]
    allow_referers = [
      "*",
	]
    allow_resources = [
      aws_location_place_index.test.index_arn,
    ]
  }
}
`, rName)
}

// Additional configuration functions (e.g., testAccKeyConfig_description, testAccKeyConfig_tags) would be defined here.
